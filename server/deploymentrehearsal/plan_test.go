package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExecutionPlanRunsEveryExactPositiveAndNegativeCheck(t *testing.T) {
	plan := validExecutionPlan()
	if err := ValidateExecutionPlan(plan); err != nil {
		t.Fatal(err)
	}
	nowValue := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	now := func() time.Time {
		value := nowValue
		nowValue = nowValue.Add(time.Millisecond)
		return value
	}
	output := t.TempDir()
	if err := ExecutePlan(context.Background(), plan, output, now); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(output)
	if err != nil || len(entries) != len(planPopulations()) {
		t.Fatalf("result population=%d err=%v", len(entries), err)
	}
	if err := ExecutePlan(context.Background(), plan, output, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("driver overlaid prior results: %v", err)
	}
}

func TestRequiredPlanChecksFollowTheLifecycleDependencyOrder(t *testing.T) {
	want := strings.Fields(`
		source_checkout_present
		missing_or_malformed_secret
		duplicate_key_id_or_value
		invalid_origin_or_proxy_depth
		clean_linux_amd64_bundle_only_install
		phase0_browser_flow_through_caddy
		non_clean_restore_target
		truncated_or_corrupt_backup
		wrong_age_identity
		wrong_release_manifest
		empty_database_backup_restore
		interrupted_backup_writer
		populated_database_identity_restore
		rpo_or_rto_above_bound
		rpo_within_six_hours
		rto_within_four_hours
		gameserver_restart_during_admitted_work
		wrong_epoch_or_artifact_set
		bounded_drain_and_restart
		missing_previous_image_or_backup
		irreversible_or_down_migration
		exact_previous_release_rollback
		current_previous_key_overlap
		public_metrics_route
		health_only_alert_receiver
		severed_alert_rule_or_counter
		early_journal_eviction
		incomplete_or_guarded_observation
		private_metrics_and_alert_delivery
		fourteen_day_journal_budget
		removed_catalog
		removed_client
		removed_license
		removed_config
		removed_helper
		changed_image_digest
		changed_runtime_config_digest
		changed_sbom
		seeded_source_secret
		seeded_image_secret
		provider_off_operation
		six_image_sbom_license_provenance
	`)
	got := make([]string, len(requiredPlanChecks))
	for index, check := range requiredPlanChecks {
		got[index] = check.name
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("lifecycle plan order changed:\n%s", strings.Join(got, "\n"))
	}
}

func TestExecutionPlanRejectsVacuousUnsafeAndGuardedCommands(t *testing.T) {
	for name, mutate := range map[string]func(*ExecutionPlan){
		"missing check": func(value *ExecutionPlan) { value.Checks = value.Checks[:len(value.Checks)-1] },
		"wrong order": func(value *ExecutionPlan) {
			value.Checks[0], value.Checks[1] = value.Checks[1], value.Checks[0]
		},
		"wrong step": func(value *ExecutionPlan) { value.Checks[0].Step = "candidate_install" },
		"duplicate":  func(value *ExecutionPlan) { value.Checks[1] = value.Checks[0] },
		"negative exits zero": func(value *ExecutionPlan) {
			for index := range value.Checks {
				if value.Checks[index].Kind == "negative" {
					value.Checks[index].ExpectedExit = 0
					return
				}
			}
		},
		"shell":    func(value *ExecutionPlan) { value.Checks[0].Command = []string{"sh", "-c", "true"} },
		"secret":   func(value *ExecutionPlan) { value.Checks[0].Command = []string{"probe", "password=private"} },
		"no guard": func(value *ExecutionPlan) { value.Checks[0].TimeoutSeconds = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			plan := validExecutionPlan()
			mutate(&plan)
			if err := ValidateExecutionPlan(plan); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unsafe/vacuous plan accepted: %v", err)
			}
		})
	}

	plan := validExecutionPlan()
	plan.Checks[0].Command = helperCommand("overflow")
	nowValue := time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)
	now := func() time.Time { value := nowValue; nowValue = nowValue.Add(time.Millisecond); return value }
	if err := ExecutePlan(context.Background(), plan, t.TempDir(), now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("overflow guard accepted: %v", err)
	}
}

func TestPlanHelperProcess(t *testing.T) {
	separator := -1
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator < 0 || separator+1 >= len(os.Args) {
		return
	}
	switch os.Args[separator+1] {
	case "pass":
		return
	case "fail":
		os.Exit(1)
	case "overflow":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", maximumCommandOutput+1))
		return
	}
}

func validExecutionPlan() ExecutionPlan {
	populations := planPopulations()
	checks := make([]PlannedCheck, len(requiredPlanChecks))
	for index, required := range requiredPlanChecks {
		kind := populations[required.name]
		mode, expected := "pass", 0
		if kind == "negative" {
			mode, expected = "fail", 1
		}
		checks[index] = PlannedCheck{Name: required.name, Kind: kind, Step: required.step,
			Command: helperCommand(mode), ExpectedExit: expected, TimeoutSeconds: 10}
	}
	return ExecutionPlan{SchemaVersion: 1, RunID: "r006-plan-001", ManifestSHA256: hashForBuild("a"),
		PreviousManifestSHA256: hashForBuild("b"), Checks: checks}
}

func helperCommand(mode string) []string {
	return []string{os.Args[0], "-test.run=^TestPlanHelperProcess$", "--", mode}
}
