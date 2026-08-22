package deploymentbackup

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"filippo.io/age"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var postgres16Version = regexp.MustCompile(`^pg_(?:dump|restore) \(PostgreSQL\) 16(?:\.[0-9]+)?\n?$`)
var databaseName = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

type PostgresBackupInput struct {
	Directory        string
	BackupID         string
	ServerID         string
	StartedAt        time.Time
	Now              func() time.Time
	Recipient        string
	DatabaseURLFile  string
	ReleaseManifest  string
	EpochDeclaration string
	PreUpgrade       bool
	Runner           CommandRunner
}

type PostgresRestoreInput struct {
	BackupPath             string
	ExpectedManifestSHA256 string
	IdentityFile           string
	TargetDatabaseURLFile  string
	Runner                 CommandRunner
}

type CommandRunner interface {
	Run(context.Context, string, []string, []string, io.Writer) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args, environment []string, output io.Writer) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), environment...)
	command.Stdout = output
	var stderr strings.Builder
	command.Stderr = &boundedWriter{Writer: &stderr, Remaining: 32 << 10}
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s failed: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func CreatePostgresBackup(ctx context.Context, input PostgresBackupInput) (Header, string, error) {
	if input.Now == nil || input.DatabaseURLFile == "" || input.ReleaseManifest == "" || input.EpochDeclaration == "" {
		return Header{}, "", ErrInvalid
	}
	runner := input.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	workspace, err := os.MkdirTemp("", ".cloud-clicker-backup-work-*")
	if err != nil {
		return Header{}, "", err
	}
	defer os.RemoveAll(workspace)
	if err := os.Chmod(workspace, 0o700); err != nil {
		return Header{}, "", err
	}
	manifestBytes, err := os.ReadFile(input.ReleaseManifest)
	if err != nil {
		return Header{}, "", err
	}
	epochBytes, err := os.ReadFile(input.EpochDeclaration)
	if err != nil {
		return Header{}, "", err
	}
	manifest, _, _, _, err := validateReleaseInputs(manifestBytes, epochBytes)
	if err != nil {
		return Header{}, "", err
	}
	environment, err := databaseServiceEnvironment(input.DatabaseURLFile, workspace)
	if err != nil {
		return Header{}, "", err
	}
	if err := requirePostgres16(ctx, runner, "pg_dump", environment); err != nil {
		return Header{}, "", err
	}
	databaseURL, err := readSecretFile(input.DatabaseURLFile)
	if err != nil {
		return Header{}, "", err
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return Header{}, "", err
	}
	if err := requireMigrationIdentity(ctx, database, manifest.DatabaseMigration); err != nil {
		_ = database.Close()
		return Header{}, "", err
	}
	if err := database.Close(); err != nil {
		return Header{}, "", err
	}
	dumpPath := filepath.Join(workspace, dumpName)
	if err := runner.Run(ctx, "pg_dump", []string{"--format=custom", "--no-owner", "--no-privileges", "--file", dumpPath, "--dbname=service=cloud_clicker"}, environment, io.Discard); err != nil {
		return Header{}, "", err
	}
	dump, err := os.Open(dumpPath)
	if err != nil {
		return Header{}, "", err
	}
	defer dump.Close()
	return CreatePackage(PackageInput{
		Directory: input.Directory, BackupID: input.BackupID, ServerID: input.ServerID,
		StartedAt: input.StartedAt, Now: input.Now, Recipient: input.Recipient, Dump: dump,
		ReleaseManifest: manifestBytes, EpochDeclaration: epochBytes, PreUpgrade: input.PreUpgrade,
	})
}

func RestorePostgresBackup(ctx context.Context, input PostgresRestoreInput) (Header, error) {
	if input.BackupPath == "" || input.IdentityFile == "" || input.TargetDatabaseURLFile == "" || !hash.MatchString(input.ExpectedManifestSHA256) {
		return Header{}, ErrInvalid
	}
	runner := input.Runner
	if runner == nil {
		runner = ExecRunner{}
	}
	identityFile, err := os.Open(input.IdentityFile)
	if err != nil {
		return Header{}, err
	}
	identities, err := age.ParseIdentities(identityFile)
	closeErr := identityFile.Close()
	if err != nil || closeErr != nil || len(identities) != 1 {
		return Header{}, ErrInvalid
	}
	workspace, err := os.MkdirTemp("", ".cloud-clicker-restore-work-*")
	if err != nil {
		return Header{}, err
	}
	defer os.RemoveAll(workspace)
	if err := os.Chmod(workspace, 0o700); err != nil {
		return Header{}, err
	}
	extractedDir := filepath.Join(workspace, "extracted")
	if err := os.Mkdir(extractedDir, 0o700); err != nil {
		return Header{}, err
	}
	extracted, err := ExtractPackage(input.BackupPath, input.ExpectedManifestSHA256, identities[0], extractedDir)
	if err != nil {
		return Header{}, err
	}
	environment, err := databaseServiceEnvironment(input.TargetDatabaseURLFile, workspace)
	if err != nil {
		return Header{}, err
	}
	if err := requirePostgres16(ctx, runner, "pg_restore", environment); err != nil {
		return Header{}, err
	}
	if err := runner.Run(ctx, "pg_restore", []string{"--list", extracted.DumpPath}, environment, io.Discard); err != nil {
		return Header{}, errors.Join(ErrInvalid, err)
	}
	databaseURL, err := readSecretFile(input.TargetDatabaseURLFile)
	if err != nil {
		return Header{}, err
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return Header{}, err
	}
	defer database.Close()
	if err := RequireCleanTarget(ctx, database); err != nil {
		return Header{}, err
	}
	if err := runner.Run(ctx, "pg_restore", []string{"--exit-on-error", "--no-owner", "--no-privileges", "--dbname=service=cloud_clicker", extracted.DumpPath}, environment, io.Discard); err != nil {
		return Header{}, err
	}
	if err := requireMigrationIdentity(ctx, database, extracted.Manifest.DatabaseMigration); err != nil {
		return Header{}, err
	}
	return extracted.Header, nil
}

func RequireCleanTarget(ctx context.Context, database *sql.DB) error {
	if database == nil {
		return ErrInvalid
	}
	var objects int
	err := database.QueryRowContext(ctx, `
		SELECT count(*)
		FROM pg_catalog.pg_class c
		JOIN pg_catalog.pg_namespace n ON n.oid=c.relnamespace
		WHERE n.nspname NOT IN ('pg_catalog','information_schema')
		  AND n.nspname !~ '^pg_toast'
		  AND c.relkind IN ('r','p','v','m','S','f')`).Scan(&objects)
	if err != nil || objects != 0 {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

func requireMigrationIdentity(ctx context.Context, database *sql.DB, expected int) error {
	var migration int
	err := database.QueryRowContext(ctx, `SELECT COALESCE(max(version_id) FILTER (WHERE is_applied),0) FROM goose_db_version`).Scan(&migration)
	if err != nil || migration != expected {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

func requirePostgres16(ctx context.Context, runner CommandRunner, binary string, environment []string) error {
	var output strings.Builder
	if err := runner.Run(ctx, binary, []string{"--version"}, environment, &output); err != nil {
		return err
	}
	if !postgres16Version.MatchString(output.String()) {
		return fmt.Errorf("%w: unsupported %s version %q", ErrInvalid, binary, strings.TrimSpace(output.String()))
	}
	return nil
}

func databaseServiceEnvironment(secretPath, workspace string) ([]string, error) {
	databaseURL, err := readSecretFile(secretPath)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" || parsed.Hostname() == "" || parsed.User == nil || parsed.User.Username() == "" || strings.TrimPrefix(parsed.EscapedPath(), "/") == "" {
		return nil, ErrInvalid
	}
	password, hasPassword := parsed.User.Password()
	if !hasPassword {
		return nil, ErrInvalid
	}
	database := strings.TrimPrefix(parsed.Path, "/")
	if !databaseName.MatchString(database) || !databaseName.MatchString(parsed.User.Username()) {
		return nil, ErrInvalid
	}
	for _, value := range []string{parsed.Hostname(), parsed.Port(), parsed.User.Username(), password, database} {
		if strings.ContainsAny(value, "\x00\r\n") {
			return nil, ErrInvalid
		}
	}
	sslMode := parsed.Query().Get("sslmode")
	if sslMode == "" {
		sslMode = "require"
	}
	if !map[string]bool{"disable": true, "require": true, "verify-ca": true, "verify-full": true}[sslMode] {
		return nil, ErrInvalid
	}
	servicePath := filepath.Join(workspace, "pg_service.conf")
	service := fmt.Sprintf("[cloud_clicker]\nhost=%s\nport=%s\ndbname=%s\nuser=%s\nsslmode=%s\n",
		parsed.Hostname(), defaultPort(parsed.Port()), database, parsed.User.Username(), sslMode)
	if err := os.WriteFile(servicePath, []byte(service), 0o600); err != nil {
		return nil, err
	}
	passPath := filepath.Join(workspace, "pgpass")
	pass := strings.Join([]string{pgpassEscape(parsed.Hostname()), pgpassEscape(defaultPort(parsed.Port())), pgpassEscape(database), pgpassEscape(parsed.User.Username()), pgpassEscape(password)}, ":") + "\n"
	if err := os.WriteFile(passPath, []byte(pass), 0o600); err != nil {
		return nil, err
	}
	return []string{"PGSERVICEFILE=" + servicePath, "PGPASSFILE=" + passPath}, nil
}

func pgpassEscape(value string) string {
	return strings.NewReplacer("\\", "\\\\", ":", "\\:").Replace(value)
}

func defaultPort(value string) string {
	if value == "" {
		return "5432"
	}
	return value
}

func readSecretFile(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o022 != 0 ||
		(info.Mode().Perm()&0o044 != 0 && filepath.Dir(path) != "/run/secrets") {
		return "", ErrInvalid
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(io.LimitReader(file, 16<<10))
	data, err := io.ReadAll(reader)
	if err != nil || len(data) == 0 || len(data) >= 16<<10 {
		return "", ErrInvalid
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", ErrInvalid
	}
	return value, nil
}

type ObjectiveObservation struct {
	IncidentAt       time.Time `json:"incident_at"`
	RestoreStartedAt time.Time `json:"restore_started_at"`
	SmokePassedAt    time.Time `json:"smoke_passed_at"`
}

type ObjectiveMeasurement struct {
	RPOSeconds int64 `json:"rpo_seconds"`
	RTOSeconds int64 `json:"rto_seconds"`
}

func MeasureObjectives(header Header, observation ObjectiveObservation) (ObjectiveMeasurement, error) {
	if validateHeader(header) != nil || observation.IncidentAt.IsZero() || observation.RestoreStartedAt.IsZero() || observation.SmokePassedAt.IsZero() ||
		observation.IncidentAt.Before(header.CompletedAt) || observation.RestoreStartedAt.Before(observation.IncidentAt) || observation.SmokePassedAt.Before(observation.RestoreStartedAt) {
		return ObjectiveMeasurement{}, ErrInvalid
	}
	measurement := ObjectiveMeasurement{
		RPOSeconds: int64(observation.IncidentAt.Sub(header.CompletedAt) / time.Second),
		RTOSeconds: int64(observation.SmokePassedAt.Sub(observation.RestoreStartedAt) / time.Second),
	}
	if measurement.RPOSeconds > int64((6*time.Hour)/time.Second) || measurement.RTOSeconds > int64((4*time.Hour)/time.Second) {
		return ObjectiveMeasurement{}, ErrInvalid
	}
	return measurement, nil
}

type boundedWriter struct {
	Writer    io.Writer
	Remaining int64
}

func (writer *boundedWriter) Write(data []byte) (int, error) {
	original := len(data)
	if writer.Remaining > 0 {
		limit := int64(len(data))
		if limit > writer.Remaining {
			limit = writer.Remaining
		}
		_, _ = writer.Writer.Write(data[:limit])
		writer.Remaining -= limit
	}
	return original, nil
}
