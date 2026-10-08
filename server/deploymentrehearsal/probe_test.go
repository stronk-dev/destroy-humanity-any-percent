package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentconfig"
	"cloud-clicker/server/operations"
	"cloud-clicker/server/releasepackage"
)

func TestBundleMutationProbesRequirePreparedMutationAndGateRejection(t *testing.T) {
	for name, mutation := range bundleMutations {
		t.Run(name, func(t *testing.T) {
			request, original := probeFixture(t, name, mutation)
			calls := 0
			outcome, err := runBundleMutationProbe(request, mutation, func(root string) error {
				calls++
				path := filepath.Join(root, filepath.FromSlash(mutation.path))
				if calls == 1 {
					// Unmutated baseline: the candidate must validate first.
					if baseline, err := os.ReadFile(path); err != nil || string(baseline) != string(original) {
						t.Fatalf("baseline validation saw a mutated candidate: %v", err)
					}
					return nil
				}
				if mutation.mutate == nil {
					if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("removal fixture remained: %v", err)
					}
				} else {
					changed, err := os.ReadFile(path)
					if err != nil || string(changed) == string(original) {
						t.Fatalf("rewrite fixture not changed: %v", err)
					}
				}
				if mutation.path != releasepackage.ReleaseManifestPath {
					data, err := os.ReadFile(filepath.Join(root, releasepackage.ReleaseManifestPath))
					if err != nil {
						t.Fatal(err)
					}
					var manifest releasepackage.ReleaseManifest
					if err := json.Unmarshal(data, &manifest); err != nil {
						t.Fatalf("mutation manifest was not rebound: %v", err)
					}
					if got := artifactHash(manifest.Artifacts, "sentinel.txt"); got != hashBytes([]byte("unchanged sentinel\n")) {
						t.Fatalf("untouched artifact not rebound: got %q", got)
					}
					want := ""
					if mutation.mutate != nil {
						changed, err := os.ReadFile(path)
						if err != nil {
							t.Fatal(err)
						}
						want = hashBytes(changed)
					}
					if got := artifactHash(manifest.Artifacts, mutation.path); got != want {
						t.Fatalf("mutation artifact hash not rebound: got %q want %q", got, want)
					}
					if name == "changed_sbom" && manifest.Images[0].SBOMSHA256 != want {
						t.Fatalf("mutated SBOM hash not rebound in image identity: got %q want %q", manifest.Images[0].SBOMSHA256, want)
					}
				}
				return errors.New("release gate rejected exact mutation")
			})
			if err != nil || outcome != ProbeRejected || calls != 2 {
				t.Fatalf("prepared rejected fixture outcome=%d calls=%d err=%v", outcome, calls, err)
			}
			unchanged, err := os.ReadFile(filepath.Join(request.CandidateBundle, filepath.FromSlash(mutation.path)))
			if err != nil || string(unchanged) != string(original) {
				t.Fatalf("probe modified retained candidate: %v", err)
			}

			outcome, err = runBundleMutationProbe(request, mutation, func(string) error { return nil })
			if err != nil || outcome != ProbeAccepted {
				t.Fatalf("gate acceptance masqueraded as rejection: outcome=%d err=%v", outcome, err)
			}
		})
	}
}

func TestBundleMutationProbeSetupFailureCannotCountAsRejection(t *testing.T) {
	mutation := bundleMutations["removed_catalog"]
	request, _ := probeFixture(t, "removed_catalog", mutation)
	if err := os.Remove(filepath.Join(request.CandidateBundle, filepath.FromSlash(mutation.path))); err != nil {
		t.Fatal(err)
	}
	calls := 0
	outcome, err := runBundleMutationProbe(request, mutation, func(string) error {
		// Only the baseline check may run; the mutated gate must never be
		// reached when the mutation itself could not be prepared.
		if calls++; calls > 1 {
			t.Fatal("validator reached after fixture setup failure")
		}
		return nil
	})
	if err == nil || outcome == ProbeRejected {
		t.Fatalf("setup failure satisfied negative row: outcome=%d err=%v", outcome, err)
	}
}

func TestProbeRejectsUnsupportedPopulationAndUnsafePaths(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "candidate")
	previous := filepath.Join(root, "previous")
	work := filepath.Join(root, "work")
	for _, path := range []string{candidate, previous, work} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	request := ProbeRequest{Population: "not_registered", CandidateBundle: candidate, PreviousBundle: previous, WorkDirectory: work}
	if outcome, err := RunProbe(request); !errors.Is(err, ErrInvalid) || outcome == ProbeRejected {
		t.Fatalf("unsupported population became rejection: outcome=%d err=%v", outcome, err)
	}
	request.WorkDirectory = filepath.Join(candidate, "work")
	if err := os.Mkdir(request.WorkDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if outcome, err := RunProbe(request); !errors.Is(err, ErrInvalid) || outcome == ProbeRejected {
		t.Fatalf("nested work path accepted: outcome=%d err=%v", outcome, err)
	}
}

func TestConfigNegativeProbesExerciseEveryMatrixAndDiscriminate(t *testing.T) {
	for _, population := range []string{"missing_or_malformed_secret", "duplicate_key_id_or_value", "invalid_origin_or_proxy_depth"} {
		t.Run(population, func(t *testing.T) {
			outcome, err := runConfigNegativeProbe(population, deploymentConfigLoader)
			if err != nil || outcome != ProbeRejected {
				t.Fatalf("config matrix outcome=%d err=%v", outcome, err)
			}

			calls := 0
			outcome, err = runConfigNegativeProbe(population, func(environment []string, readFile deploymentconfig.ReadFile) (deploymentconfig.Config, error) {
				calls++
				if calls == 2 {
					return deploymentconfig.Config{}, nil
				}
				return deploymentConfigLoader(environment, readFile)
			})
			if err != nil || outcome != ProbeAccepted {
				t.Fatalf("accepted severing did not fail probe: outcome=%d err=%v", outcome, err)
			}
		})
	}
}

func TestPublicMetricsProbeRebindsArtifactBeforeSemanticRejection(t *testing.T) {
	request, _ := probeFixture(t, "public_metrics_route", bundleMutation{path: "Caddyfile"})
	manifest := releasepackage.ReleaseManifest{Artifacts: []releasepackage.File{{Path: "Caddyfile", SHA256: hashBytes([]byte("fixture"))}}}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(request.CandidateBundle, releasepackage.ReleaseManifestPath), append(encoded, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	outcome, err := runPreparedBundleProbe(request, preparePublicMetricsRoute, func(root string) error {
		calls++
		caddy, readErr := os.ReadFile(filepath.Join(root, "Caddyfile"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		manifestData, readErr := os.ReadFile(filepath.Join(root, releasepackage.ReleaseManifestPath))
		var changed releasepackage.ReleaseManifest
		if readErr != nil || json.Unmarshal(manifestData, &changed) != nil || changed.Artifacts[0].SHA256 != hashBytes(caddy) {
			t.Fatalf("mutated artifact was not rebound: %v", readErr)
		}
		if calls == 1 {
			if string(caddy) != "fixture" {
				t.Fatal("baseline already contains the injected route")
			}
			return nil
		}
		if !strings.Contains(string(caddy), "/metrics") {
			t.Fatal("public route fixture missing")
		}
		return errors.New("semantic Caddy gate rejected public metrics")
	})
	if err != nil || outcome != ProbeRejected || calls != 2 {
		t.Fatalf("public metrics outcome=%d err=%v", outcome, err)
	}
	original, err := os.ReadFile(filepath.Join(request.CandidateBundle, "Caddyfile"))
	if err != nil || string(original) != "fixture" {
		t.Fatalf("retained Caddyfile changed: %q err=%v", original, err)
	}
	outcome, err = runPreparedBundleProbe(request, preparePublicMetricsRoute, func(string) error { return nil })
	if err != nil || outcome != ProbeAccepted {
		t.Fatalf("accepted public route counted as a successful negative: outcome=%d err=%v", outcome, err)
	}
	assertProbeWorkEmpty(t, request.WorkDirectory)
}

func TestSeededSecretProbesRequireAnActualScannerFinding(t *testing.T) {
	for _, population := range []string{"seeded_source_secret", "seeded_image_secret"} {
		t.Run(population, func(t *testing.T) {
			request := validProbeDirectories(t, population)
			calls := 0
			outcome, err := runSecretNegativeProbe(request, func(findings []releasepackage.SecretFinding) error {
				calls++
				if calls == 1 && len(findings) != 0 {
					t.Fatal("secret gate did not receive the clean scanner baseline first")
				}
				if calls == 2 && (len(findings) != 1 || findings[0].Rule != "seeded-fixture") {
					t.Fatal("secret gate did not receive the actual seeded finding")
				}
				return releasepackage.RequireNoSecrets(findings)
			})
			if err != nil || outcome != ProbeRejected || calls != 2 {
				t.Fatalf("secret fixture outcome=%d err=%v", outcome, err)
			}
			outcome, err = runSecretNegativeProbe(request, func([]releasepackage.SecretFinding) error { return nil })
			if err != nil || outcome != ProbeAccepted {
				t.Fatalf("sleeping secret gate satisfied row: outcome=%d err=%v", outcome, err)
			}
		})
	}
}

func TestReleaseNegativeProbesCannotCountAlwaysRejectingGates(t *testing.T) {
	for _, population := range []string{"public_metrics_route", "seeded_source_secret", "seeded_image_secret"} {
		t.Run(population, func(t *testing.T) {
			request := validProbeDirectories(t, population)
			unrelated := errors.New("baseline gate is broken")
			var outcome ProbeOutcome
			var err error
			if population == "public_metrics_route" {
				prepared := false
				outcome, err = runPreparedBundleProbe(request, func(string) error {
					prepared = true
					return nil
				}, func(string) error { return unrelated })
				if prepared {
					t.Error("negative prepared before the healthy baseline passed")
				}
			} else {
				outcome, err = runSecretNegativeProbe(request, func([]releasepackage.SecretFinding) error { return unrelated })
			}
			if !errors.Is(err, unrelated) || outcome == ProbeRejected {
				t.Fatalf("always-rejecting gate counted as release proof: outcome=%d err=%v", outcome, err)
			}
			assertProbeWorkEmpty(t, request.WorkDirectory)
		})
	}
}

func TestTypedHostAlertJournalAndObjectiveProbesDiscriminate(t *testing.T) {
	for _, population := range []string{"source_checkout_present", "health_only_alert_receiver", "severed_alert_rule_or_counter",
		"early_journal_eviction", "incomplete_or_guarded_observation", "rpo_or_rto_above_bound"} {
		t.Run(population, func(t *testing.T) {
			request := validProbeDirectories(t, population)
			outcome, err := RunProbe(request)
			if err != nil || outcome != ProbeRejected {
				t.Fatalf("typed probe outcome=%d err=%v", outcome, err)
			}
		})
	}

	if outcome, err := runHostNegativeProbe(func(Host) error { return nil }); err != nil || outcome != ProbeAccepted {
		t.Fatalf("sleeping host gate satisfied negative: outcome=%d err=%v", outcome, err)
	}
	alertCalls := 0
	if outcome, err := runAlertNegativeProbe("health_only_alert_receiver", func(value operations.AlertDeliveryObservation) error {
		alertCalls++
		if alertCalls == 2 {
			return nil
		}
		return operations.ValidateAlertDeliveryObservation(value)
	}); err != nil || outcome != ProbeAccepted {
		t.Fatalf("sleeping alert gate satisfied negative: outcome=%d err=%v", outcome, err)
	}
	journalCalls := 0
	if outcome, err := runJournalNegativeProbe("early_journal_eviction", func(value operations.JournalObservation) error {
		journalCalls++
		if journalCalls == 2 {
			return nil
		}
		return operations.ValidateJournalObservation(value)
	}); err != nil || outcome != ProbeAccepted {
		t.Fatalf("sleeping journal gate satisfied negative: outcome=%d err=%v", outcome, err)
	}
	objectiveCalls := 0
	if outcome, err := runObjectiveNegativeProbe(func(value Objectives, start, end time.Time) error {
		objectiveCalls++
		if objectiveCalls == 2 {
			return nil
		}
		return validateObjectives(value, start, end)
	}); err != nil || outcome != ProbeAccepted {
		t.Fatalf("sleeping objective gate satisfied negative: outcome=%d err=%v", outcome, err)
	}
}

func TestBackupEnvelopeProbesDiscriminate(t *testing.T) {
	for _, population := range []string{"truncated_or_corrupt_backup", "wrong_age_identity", "wrong_release_manifest", "interrupted_backup_writer"} {
		t.Run(population, func(t *testing.T) {
			request := validProbeDirectories(t, population)
			outcome, err := RunProbe(request)
			if err != nil || outcome != ProbeRejected {
				t.Fatalf("backup probe outcome=%d err=%v", outcome, err)
			}
		})
	}

	request := validProbeDirectories(t, "wrong_release_manifest")
	restoreCalls := 0
	outcome, err := runBackupRestoreNegativeProbe(request, func(path, manifest string, identity age.Identity, output io.Writer) (deploymentbackup.Header, error) {
		restoreCalls++
		if restoreCalls == 2 {
			return deploymentbackup.Header{}, nil
		}
		return deploymentbackup.Restore(path, manifest, identity, output)
	})
	if err != nil || outcome != ProbeAccepted {
		t.Fatalf("sleeping restore gate satisfied negative: outcome=%d err=%v", outcome, err)
	}

	request = validProbeDirectories(t, "interrupted_backup_writer")
	createCalls := 0
	outcome, err = runInterruptedBackupProbe(request, func(input deploymentbackup.CreateInput) (deploymentbackup.Header, string, error) {
		createCalls++
		if createCalls == 2 {
			// A defective writer discards the read failure and commits the prefix.
			data, _ := io.ReadAll(input.Dump)
			input.Dump = strings.NewReader(string(data))
		}
		return deploymentbackup.Create(input)
	})
	if err != nil || outcome != ProbeAccepted || createCalls != 2 {
		t.Fatalf("sleeping create gate satisfied negative: outcome=%d err=%v", outcome, err)
	}
}

func TestInterruptedBackupProbeCannotCountUnrelatedFailures(t *testing.T) {
	for _, kind := range []string{"always_rejects", "rejects_before_read", "different_error_after_read"} {
		t.Run(kind, func(t *testing.T) {
			request := validProbeDirectories(t, "interrupted_backup_writer")
			calls := 0
			unrelated := errors.New("unrelated backup failure")
			outcome, err := runInterruptedBackupProbe(request, func(input deploymentbackup.CreateInput) (deploymentbackup.Header, string, error) {
				calls++
				if kind != "always_rejects" && calls == 1 {
					return deploymentbackup.Create(input)
				}
				if kind == "different_error_after_read" {
					_, _ = io.Copy(io.Discard, input.Dump)
				}
				return deploymentbackup.Header{}, "", unrelated
			})
			if err == nil || outcome == ProbeRejected {
				t.Fatalf("unrelated failure counted as interrupted-write proof: outcome=%d err=%v calls=%d", outcome, err, calls)
			}
			entries, readErr := os.ReadDir(request.WorkDirectory)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("probe leaked its temporary population: entries=%v err=%v", entries, readErr)
			}
		})
	}
}

func TestInterruptedBackupProbeRequiresRestorableBaselineAndPreservesIt(t *testing.T) {
	for _, kind := range []string{"no_envelope", "wrong_payload", "wrapped_interruption", "partial_file", "changed_baseline"} {
		t.Run(kind, func(t *testing.T) {
			request := validProbeDirectories(t, "interrupted_backup_writer")
			calls := 0
			var baselinePath string
			outcome, err := runInterruptedBackupProbe(request, func(input deploymentbackup.CreateInput) (deploymentbackup.Header, string, error) {
				calls++
				if calls == 1 {
					if kind == "no_envelope" {
						return deploymentbackup.Header{}, "", nil
					}
					if kind == "wrong_payload" {
						input.Dump = strings.NewReader("wrong baseline bytes")
					}
					header, path, err := deploymentbackup.Create(input)
					baselinePath = path
					return header, path, err
				}
				_, readErr := io.Copy(io.Discard, input.Dump)
				if readErr == nil {
					t.Fatal("negative fixture did not deliver its interruption")
				}
				if kind == "partial_file" {
					if err := os.WriteFile(filepath.Join(input.Directory, ".leftover.tmp"), []byte("partial"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if kind == "changed_baseline" {
					if err := os.WriteFile(baselinePath, []byte("corrupted"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				return deploymentbackup.Header{}, "", errors.Join(errors.New("writer context"), readErr)
			})
			if kind == "wrapped_interruption" {
				if err != nil || outcome != ProbeRejected || calls != 2 {
					t.Fatalf("propagated interruption was not recognized: outcome=%d err=%v calls=%d", outcome, err, calls)
				}
			} else if err == nil || outcome == ProbeRejected {
				t.Fatalf("invalid baseline/cleanup counted: outcome=%d err=%v calls=%d", outcome, err, calls)
			}
			entries, readErr := os.ReadDir(request.WorkDirectory)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("probe leaked its temporary population: entries=%v err=%v", entries, readErr)
			}
		})
	}
}

func deploymentConfigLoader(environment []string, readFile deploymentconfig.ReadFile) (deploymentconfig.Config, error) {
	return deploymentconfig.Load(environment, readFile)
}

func probeFixture(t *testing.T, name string, mutation bundleMutation) (ProbeRequest, []byte) {
	t.Helper()
	request := validProbeDirectories(t, name)
	candidate := request.CandidateBundle
	original := []byte("fixture")
	if mutation.mutate != nil {
		if name == "wrong_epoch_or_artifact_set" {
			original = []byte("{\"current_epoch_id\": 8}\n")
		} else if name == "changed_sbom" {
			original = []byte("{\"name\":\"sha256-" + strings.Repeat("a", 64) + "\"}\n")
		} else {
			original = []byte("{\"reference\": \"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\", \"runtime_config_sha256\": \"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\"}\n")
		}
	}
	path := filepath.Join(candidate, filepath.FromSlash(mutation.path))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if mutation.path != releasepackage.ReleaseManifestPath {
		if err := os.WriteFile(filepath.Join(candidate, "sentinel.txt"), []byte("unchanged sentinel\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		manifest := []byte(`{"images":[{"name":"caddy","sbom_path":"sbom/caddy.spdx.json"}]}` + "\n")
		if err := os.WriteFile(filepath.Join(candidate, releasepackage.ReleaseManifestPath), manifest, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return request, original
}

func validProbeDirectories(t *testing.T, population string) ProbeRequest {
	t.Helper()
	root := t.TempDir()
	candidate := filepath.Join(root, "candidate")
	previous := filepath.Join(root, "previous")
	work := filepath.Join(root, "work")
	for _, path := range []string{candidate, previous, work} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return ProbeRequest{Population: population, CandidateBundle: candidate, PreviousBundle: previous, WorkDirectory: work}
}
