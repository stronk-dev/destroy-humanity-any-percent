// Package deploymentrehearsal owns the machine-readable R-006 evidence
// boundary. It validates observed results; it does not turn component tests or
// an operator-authored JSON file into deployment proof.
package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid deployment rehearsal evidence")

const (
	SchemaVersion = 1
	MaximumRPO    = 6 * time.Hour
	MaximumRTO    = 4 * time.Hour
)

type Evidence struct {
	SchemaVersion          int          `json:"schema_version"`
	RunID                  string       `json:"run_id"`
	ManifestSHA256         string       `json:"manifest_sha256"`
	PreviousManifestSHA256 string       `json:"previous_manifest_sha256"`
	ReleaseVersion         string       `json:"release_version"`
	PreviousReleaseVersion string       `json:"previous_release_version"`
	StartedAt              time.Time    `json:"started_at"`
	CompletedAt            time.Time    `json:"completed_at"`
	Host                   Host         `json:"host"`
	Tools                  []Tool       `json:"tools"`
	Steps                  []Step       `json:"steps"`
	Populations            []Population `json:"populations"`
	Objectives             Objectives   `json:"objectives"`
	Artifacts              []Artifact   `json:"artifacts"`
	Exclusions             []string     `json:"exclusions"`
	ObjectiveCompleted     bool         `json:"objective_completed"`
	GuardExhausted         bool         `json:"guard_exhausted"`
}

type Host struct {
	OS                        string `json:"os"`
	Architecture              string `json:"architecture"`
	Distribution              string `json:"distribution"`
	Kernel                    string `json:"kernel"`
	DockerEngine              string `json:"docker_engine"`
	DockerCompose             string `json:"docker_compose"`
	CleanStart                bool   `json:"clean_start"`
	SourceCheckoutAbsent      bool   `json:"source_checkout_absent"`
	ProviderCredentialsAbsent bool   `json:"provider_credentials_absent"`
}

type Tool struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type Step struct {
	Name         string    `json:"name"`
	CommandClass string    `json:"command_class"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	InputSHA256  string    `json:"input_sha256"`
	OutputSHA256 string    `json:"output_sha256"`
	Result       string    `json:"result"`
}

type Population struct {
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	Result         string `json:"result"`
	SeveringCaught bool   `json:"severing_caught"`
	EvidenceSHA256 string `json:"evidence_sha256"`
}

type Objectives struct {
	IncidentAt            time.Time `json:"incident_at"`
	NewestValidBackupAt   time.Time `json:"newest_valid_backup_at"`
	RestoreStartedAt      time.Time `json:"restore_started_at"`
	AuthenticatedSmokeAt  time.Time `json:"authenticated_smoke_at"`
	RPOSeconds            int64     `json:"rpo_seconds"`
	RTOSeconds            int64     `json:"rto_seconds"`
	RestoredIdentityMatch bool      `json:"restored_identity_match"`
}

type Artifact struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

var RequiredSteps = []string{
	"host_preflight",
	"candidate_install",
	"browser_phase0",
	"empty_backup_restore",
	"populated_backup_restore",
	"incident_recovery",
	"candidate_release",
	"previous_release_rollback",
	"key_rotation",
	"operations_alerts_retention",
	"provider_off_supply_chain",
}

var requiredCommandClass = map[string]string{
	"host_preflight":              "host_observation",
	"candidate_install":           "deployment_install",
	"browser_phase0":              "browser_workflow",
	"empty_backup_restore":        "backup_restore",
	"populated_backup_restore":    "backup_restore",
	"incident_recovery":           "recovery_objective",
	"candidate_release":           "deployment_release",
	"previous_release_rollback":   "deployment_rollback",
	"key_rotation":                "credential_rotation",
	"operations_alerts_retention": "operations_rehearsal",
	"provider_off_supply_chain":   "supply_chain_rehearsal",
}

var RequiredPopulations = map[string]string{
	"clean_linux_amd64_bundle_only_install":   "positive",
	"phase0_browser_flow_through_caddy":       "positive",
	"empty_database_backup_restore":           "positive",
	"populated_database_identity_restore":     "positive",
	"rpo_within_six_hours":                    "positive",
	"rto_within_four_hours":                   "positive",
	"bounded_drain_and_restart":               "positive",
	"exact_previous_release_rollback":         "positive",
	"current_previous_key_overlap":            "positive",
	"private_metrics_and_alert_delivery":      "positive",
	"fourteen_day_journal_budget":             "positive",
	"provider_off_operation":                  "positive",
	"six_image_sbom_license_provenance":       "positive",
	"removed_catalog":                         "negative",
	"removed_client":                          "negative",
	"removed_license":                         "negative",
	"removed_config":                          "negative",
	"removed_helper":                          "negative",
	"changed_image_digest":                    "negative",
	"changed_runtime_config_digest":           "negative",
	"changed_sbom":                            "negative",
	"source_checkout_present":                 "negative",
	"missing_or_malformed_secret":             "negative",
	"duplicate_key_id_or_value":               "negative",
	"invalid_origin_or_proxy_depth":           "negative",
	"seeded_source_secret":                    "negative",
	"seeded_image_secret":                     "negative",
	"truncated_or_corrupt_backup":             "negative",
	"wrong_age_identity":                      "negative",
	"wrong_release_manifest":                  "negative",
	"non_clean_restore_target":                "negative",
	"interrupted_backup_writer":               "negative",
	"gameserver_restart_during_admitted_work": "negative",
	"wrong_epoch_or_artifact_set":             "negative",
	"missing_previous_image_or_backup":        "negative",
	"irreversible_or_down_migration":          "negative",
	"public_metrics_route":                    "negative",
	"health_only_alert_receiver":              "negative",
	"severed_alert_rule_or_counter":           "negative",
	"early_journal_eviction":                  "negative",
	"incomplete_or_guarded_observation":       "negative",
	"rpo_or_rto_above_bound":                  "negative",
	"forged_successful_evidence":              "negative",
}

var RequiredArtifacts = []string{
	"candidate_manifest",
	"previous_manifest",
	"release_ledger",
	"rotation_ledger",
	"backup_header",
	"browser_result",
	"alert_delivery",
	"journal_observation",
	"rehearsal_plan",
	"secret_scan",
	"supply_chain",
}

var RequiredExclusions = []string{
	"public_hosting",
	"availability_sla",
	"multi_node",
	"linux_arm64",
	"account_retention",
	"sunset_covenant",
}

var (
	identifierPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
	versionPattern    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
	hashPattern       = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

func Load(path string) (Evidence, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Evidence{}, err
	}
	return Decode(data)
}

func Decode(data []byte) (Evidence, error) {
	if containsForbiddenEvidenceKey(data) {
		return Evidence{}, ErrInvalid
	}
	var evidence Evidence
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&evidence) != nil || decoder.Decode(&struct{}{}) != io.EOF || Validate(evidence) != nil {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
}

func Validate(evidence Evidence) error {
	if evidence.SchemaVersion != SchemaVersion || !identifierPattern.MatchString(evidence.RunID) ||
		!hashPattern.MatchString(evidence.ManifestSHA256) || !hashPattern.MatchString(evidence.PreviousManifestSHA256) ||
		evidence.ManifestSHA256 == evidence.PreviousManifestSHA256 || !versionPattern.MatchString(evidence.ReleaseVersion) ||
		!versionPattern.MatchString(evidence.PreviousReleaseVersion) || evidence.ReleaseVersion == evidence.PreviousReleaseVersion ||
		evidence.StartedAt.IsZero() || !evidence.CompletedAt.After(evidence.StartedAt) || !evidence.ObjectiveCompleted || evidence.GuardExhausted {
		return ErrInvalid
	}
	if err := validateHost(evidence.Host); err != nil {
		return err
	}
	if err := validateTools(evidence.Tools); err != nil {
		return err
	}
	if err := validateSteps(evidence.Steps, evidence.StartedAt, evidence.CompletedAt); err != nil {
		return err
	}
	if err := validatePopulations(evidence.Populations); err != nil {
		return err
	}
	if err := validateObjectives(evidence.Objectives, evidence.StartedAt, evidence.CompletedAt); err != nil {
		return err
	}
	if err := requireNamedHashes(evidence.Artifacts, RequiredArtifacts); err != nil {
		return err
	}
	if !sameStringSet(evidence.Exclusions, RequiredExclusions) {
		return ErrInvalid
	}
	return nil
}

func validateHost(host Host) error {
	if host.OS != "linux" || host.Architecture != "amd64" || !identifierPattern.MatchString(host.Distribution) ||
		host.Kernel == "" || len(host.Kernel) > 128 || host.DockerEngine == "" || len(host.DockerEngine) > 128 ||
		host.DockerCompose == "" || len(host.DockerCompose) > 128 || !host.CleanStart || !host.SourceCheckoutAbsent ||
		!host.ProviderCredentialsAbsent {
		return ErrInvalid
	}
	return nil
}

func validateTools(tools []Tool) error {
	required := []string{"deployment-rehearsal", "deployment-release", "browser-driver"}
	artifacts := make([]Artifact, len(tools))
	for index, tool := range tools {
		artifacts[index] = Artifact{Name: tool.Name, SHA256: tool.SHA256}
	}
	return requireNamedHashes(artifacts, required)
}

func validateSteps(steps []Step, runStart, runEnd time.Time) error {
	if len(steps) != len(RequiredSteps) {
		return ErrInvalid
	}
	for index, name := range RequiredSteps {
		step := steps[index]
		if step.Name != name || step.CommandClass != requiredCommandClass[name] || step.Result != "passed" ||
			step.StartedAt.Before(runStart) || step.CompletedAt.After(runEnd) || !step.CompletedAt.After(step.StartedAt) ||
			!hashPattern.MatchString(step.InputSHA256) || !hashPattern.MatchString(step.OutputSHA256) ||
			index > 0 && step.StartedAt.Before(steps[index-1].CompletedAt) {
			return ErrInvalid
		}
	}
	return nil
}

func validatePopulations(populations []Population) error {
	if len(populations) != len(RequiredPopulations) {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for _, population := range populations {
		kind, ok := RequiredPopulations[population.Name]
		if !ok || seen[population.Name] || population.Kind != kind || population.Result != "passed" ||
			population.SeveringCaught != (kind == "negative") || !hashPattern.MatchString(population.EvidenceSHA256) {
			return ErrInvalid
		}
		seen[population.Name] = true
	}
	return nil
}

func validateObjectives(objectives Objectives, runStart, runEnd time.Time) error {
	if objectives.IncidentAt.Before(runStart) || objectives.IncidentAt.After(runEnd) ||
		objectives.NewestValidBackupAt.After(objectives.IncidentAt) || objectives.RestoreStartedAt.Before(objectives.IncidentAt) ||
		objectives.AuthenticatedSmokeAt.Before(objectives.RestoreStartedAt) || objectives.AuthenticatedSmokeAt.After(runEnd) ||
		!objectives.RestoredIdentityMatch {
		return ErrInvalid
	}
	rpo := objectives.IncidentAt.Sub(objectives.NewestValidBackupAt)
	rto := objectives.AuthenticatedSmokeAt.Sub(objectives.RestoreStartedAt)
	if rpo < 0 || rto < 0 || rpo > MaximumRPO || rto > MaximumRTO || objectives.RPOSeconds != int64(rpo/time.Second) ||
		objectives.RTOSeconds != int64(rto/time.Second) {
		return ErrInvalid
	}
	return nil
}

func requireNamedHashes(artifacts []Artifact, required []string) error {
	if len(artifacts) != len(required) {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for _, artifact := range artifacts {
		if !slices.Contains(required, artifact.Name) || seen[artifact.Name] || !hashPattern.MatchString(artifact.SHA256) {
			return ErrInvalid
		}
		seen[artifact.Name] = true
	}
	return nil
}

func sameStringSet(actual, required []string) bool {
	if len(actual) != len(required) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range actual {
		if !slices.Contains(required, value) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func containsForbiddenEvidenceKey(data []byte) bool {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return false
	}
	forbidden := map[string]bool{
		"account_id": true, "founder_id": true, "raw_ip": true, "database_url": true,
		"password": true, "recovery_code": true, "secret_value": true, "token": true,
	}
	var walk func(any) bool
	walk = func(current any) bool {
		switch typed := current.(type) {
		case map[string]any:
			for key, child := range typed {
				if forbidden[strings.ToLower(key)] || walk(child) {
					return true
				}
			}
		case []any:
			for _, child := range typed {
				if walk(child) {
					return true
				}
			}
		}
		return false
	}
	return walk(value)
}

func Explain(err error) string {
	if err == nil {
		return "valid"
	}
	return fmt.Sprintf("%v", err)
}
