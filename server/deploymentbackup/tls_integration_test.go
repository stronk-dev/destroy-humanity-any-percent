package deploymentbackup

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
	"filippo.io/age"
)

func backupTLSCertificate(t *testing.T, name string, parent *x509.Certificate, signer *ecdsa.PrivateKey, server bool) ([]byte, []byte, *x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: name}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	if parent == nil {
		certificate.IsCA = true
		certificate.BasicConstraintsValid = true
		certificate.KeyUsage |= x509.KeyUsageCertSign
		parent, signer = certificate, key
	} else if server {
		certificate.DNSNames = []string{name}
		certificate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	} else {
		certificate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	der, err := x509.CreateCertificate(rand.Reader, certificate, parent, &key.PublicKey, signer)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	encodedKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encodedKey}), parsed, key
}

// Exercise the actual service-file consumer, not only its generated text. The
// declared disposable Postgres superuser installs ephemeral fixtures through COPY;
// no host file or existing key is read, and SSL settings are reset on every exit.
func TestPostgresBackupPreservesTLSCertificateParametersIntegration(t *testing.T) {
	sourceURL, targetURL, adminURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_RESTORE_DATABASE_URL"), os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || targetURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	for _, raw := range []string{sourceURL, targetURL, adminURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Host != "postgres:5432" || u.Query().Get("sslmode") != "disable" {
			t.Fatal("TLS fixture requires the declared disposable Docker service")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := save.OpenPostgres(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	resetDatabase(t, adminURL, "cloud_clicker_source")
	resetDatabase(t, adminURL, "cloud_clicker_restore")
	var sslContext string
	if err := admin.QueryRowContext(ctx, "SELECT context FROM pg_settings WHERE name='ssl'").Scan(&sslContext); err != nil || sslContext != "sighup" {
		t.Fatal("TLS fixture requires reloadable SSL", err)
	}
	ca, _, issuer, issuerKey := backupTLSCertificate(t, "backup-test-ca", nil, nil, false)
	serverCert, serverKey, _, _ := backupTLSCertificate(t, "postgres", issuer, issuerKey, true)
	clientCert, clientKey, _, _ := backupTLSCertificate(t, "cloud_clicker", issuer, issuerKey, false)
	suffix := strings.ToLower(rand.Text())
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	serverFiles := map[string][]byte{"ssl_cert_file": serverCert, "ssl_key_file": serverKey, "ssl_ca_file": ca}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		for _, setting := range []string{"ssl", "ssl_cert_file", "ssl_key_file", "ssl_ca_file"} {
			if _, err := admin.ExecContext(cleanup, "ALTER SYSTEM RESET "+setting); err != nil {
				t.Error("reset ephemeral TLS setting", err)
			}
		}
		if _, err := admin.ExecContext(cleanup, "SELECT pg_reload_conf()"); err != nil {
			t.Error("reset TLS reload", err)
		}
	}()
	for setting, data := range serverFiles {
		path := "/tmp/backup-tls-" + suffix + "-" + setting + ".pem"
		// One PEM line per COPY text row preserves exact PEM bytes; embedded
		// newlines in a single COPY text field would instead become backslashes.
		// COPY's ordinary files follow the server umask (often0644). This
		// fixed fixture-local command creates the generated key privately from
		// its first byte; it accesses no host file or pre-existing secret.
		copySQL := "COPY (SELECT unnest(string_to_array(" + quote(strings.TrimSuffix(string(data), "\n")) + ", E'\\n'))) TO PROGRAM " + quote("umask 077; cat > "+path)
		if _, err := admin.ExecContext(ctx, copySQL); err != nil {
			t.Fatal("install generated server TLS fixture", err)
		}
		if _, err := admin.ExecContext(ctx, "ALTER SYSTEM SET "+setting+" = "+quote(path)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := admin.ExecContext(ctx, "ALTER SYSTEM SET ssl = 'on'"); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.ExecContext(ctx, "SELECT pg_reload_conf()"); err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	caPath := writeRegular(t, workspace, "custom ca.pem", ca)
	certPath := writeRegular(t, workspace, "custom client.pem", clientCert)
	keyPath := writeSecret(t, workspace, "custom client.key", string(clientKey))
	withTLS := func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		query := u.Query()
		query.Set("sslmode", "verify-full")
		query.Set("sslrootcert", caPath)
		query.Set("sslcert", certPath)
		query.Set("sslkey", keyPath)
		u.RawQuery = query.Encode()
		return u.String()
	}
	tlsSource, tlsTarget := withTLS(sourceURL), withTLS(targetURL)
	// Require a successful actual Go TLS connection before testing libpq. A
	// reload/setup error cannot masquerade as the missing-service-field regression.
	probe, err := sql.Open("pgx", tlsSource)
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Close()
	var goIdentity string
	until := time.Now().Add(10 * time.Second)
	for {
		err = probe.QueryRowContext(ctx, "SELECT ssl::text||'|'||COALESCE(client_dn,'missing') FROM pg_stat_ssl WHERE pid=pg_backend_pid()").Scan(&goIdentity)
		if !time.Now().Before(until) {
			t.Fatal("Go TLS setup deadline", err)
		}
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if goIdentity != "true|CN=cloud_clicker" && goIdentity != "true|/CN=cloud_clicker" {
		t.Fatal("Go inspection did not use the declared verified TLS/client certificate", goIdentity)
	}
	sourceSecret := writeSecret(t, workspace, "source-url", tlsSource)
	targetSecret := writeSecret(t, workspace, "target-url", tlsTarget)
	environment, err := databaseServiceEnvironment(sourceSecret, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var clientIdentity bytes.Buffer
	if err := (ExecRunner{}).Run(ctx, "psql", []string{"--dbname=service=cloud_clicker", "-Atqc", "SELECT ssl::text||'|'||COALESCE(client_dn,'missing') FROM pg_stat_ssl WHERE pid=pg_backend_pid()"}, environment, &clientIdentity); err != nil {
		t.Fatal("libpq failed with the same declared TLS inputs after Go connected", err)
	}
	if strings.TrimSpace(clientIdentity.String()) != goIdentity {
		t.Fatal("libpq TLS/client certificate differs from Go", clientIdentity.String())
	}
	if err := save.Migrate(ctx, probe); err != nil {
		t.Fatal(err)
	}
	seedBackupEpoch(t, probe)
	seedBackupPopulation(t, probe)
	var migration int
	if err := probe.QueryRowContext(ctx, "SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version").Scan(&migration); err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	before, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: sourceSecret})
	if err != nil {
		t.Fatal(err)
	}
	_, backup, err := CreatePostgresBackup(ctx, PostgresBackupInput{Directory: t.TempDir(), BackupID: "20261104T120000Z-000000000003", ServerID: "tls-test", StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(), DatabaseURLFile: sourceSecret, ReleaseManifest: writeRegular(t, workspace, "manifest.json", manifest), EpochDeclaration: writeRegular(t, workspace, "epoch.json", epoch)})
	if err != nil {
		t.Fatal("actual verified-TLS encrypted backup", err)
	}
	if _, err := RestorePostgresBackup(ctx, PostgresRestoreInput{BackupPath: backup, ExpectedManifestSHA256: digest(manifest), IdentityFile: writeSecret(t, workspace, "age-identity", identity.String()), TargetDatabaseURLFile: targetSecret}); err != nil {
		t.Fatal("actual verified-TLS encrypted restore", err)
	}
	after, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: targetSecret})
	if err != nil || CompareRecoveryIdentity(before, after) != nil {
		t.Fatal("TLS restored identity mismatch", err)
	}
	wrongCA, _, _, _ := backupTLSCertificate(t, "wrong-ca", nil, nil, false)
	u, _ := url.Parse(tlsSource)
	query := u.Query()
	query.Set("sslrootcert", writeRegular(t, workspace, "wrong-ca.pem", wrongCA))
	u.RawQuery = query.Encode()
	wrongEnvironment, err := databaseServiceEnvironment(writeSecret(t, workspace, "wrong-url", u.String()), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := (ExecRunner{}).Run(ctx, "psql", []string{"--dbname=service=cloud_clicker", "-Atqc", "SELECT 1"}, wrongEnvironment, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "certificate verify failed") {
		t.Fatal("wrong CA did not cause the certificate-verification refusal", err)
	}
	if ctx.Err() != nil {
		t.Fatal("TLS integration incomplete", ctx.Err())
	}
	t.Log("verified TLS and client identity match in Go/libpq; encrypted backup/restore identity matches; wrong CA refused")
}
