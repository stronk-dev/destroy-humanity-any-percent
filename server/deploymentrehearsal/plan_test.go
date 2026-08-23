package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"sort"
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
	if err != nil || len(entries) != len(RequiredPopulations) {
		t.Fatalf("result population=%d err=%v", len(entries), err)
	}
	if err := ExecutePlan(context.Background(), plan, output, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("driver overlaid prior results: %v", err)
	}
}

func TestExecutionPlanRejectsVacuousUnsafeAndGuardedCommands(t *testing.T) {
	for name, mutate := range map[string]func(*ExecutionPlan){
		"missing check": func(value *ExecutionPlan) { value.Checks = value.Checks[:len(value.Checks)-1] },
		"wrong order": func(value *ExecutionPlan) {
			value.Checks[0], value.Checks[1] = value.Checks[1], value.Checks[0]
		},
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
	names := make([]string, 0, len(RequiredPopulations))
	for name := range RequiredPopulations {
		names = append(names, name)
	}
	sort.Strings(names)
	checks := make([]PlannedCheck, len(names))
	for index, name := range names {
		kind := RequiredPopulations[name]
		mode, expected := "pass", 0
		if kind == "negative" {
			mode, expected = "fail", 1
		}
		stepIndex := index * len(RequiredSteps) / len(names)
		checks[index] = PlannedCheck{Name: name, Kind: kind, Step: RequiredSteps[stepIndex],
			Command: helperCommand(mode), ExpectedExit: expected, TimeoutSeconds: 10}
	}
	return ExecutionPlan{SchemaVersion: 1, RunID: "r006-plan-001", ManifestSHA256: hashForBuild("a"),
		PreviousManifestSHA256: hashForBuild("b"), Checks: checks}
}

func helperCommand(mode string) []string {
	return []string{os.Args[0], "-test.run=^TestPlanHelperProcess$", "--", mode}
}
