package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"cloud-clicker/server/deploymentrehearsal"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage")
	}
	switch os.Args[1] {
	case "validate":
		validated, err := runValidate(os.Args[2:])
		if err != nil {
			fail("invalid_evidence")
		}
		fmt.Printf("R-006 evidence valid: run=%s manifest=%s\n", validated.RunID, validated.ManifestSHA256)
	case "validate-build":
		validated, err := runValidateBuild(os.Args[2:])
		if err != nil {
			fail("invalid_evidence")
		}
		fmt.Printf("release build valid: role=%s manifest=%s\n", validated.Role, validated.ManifestSHA256)
	case "run-plan":
		if err := runPlan(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal plan passed")
	case "probe":
		outcome, err := runProbe(os.Args[2:])
		if err != nil {
			failCode("invalid_probe", 2)
		}
		if outcome == deploymentrehearsal.ProbeRejected {
			failCode("fixture_rejected", 1)
		}
		fmt.Println("deployment rehearsal probe passed")
	default:
		fail("usage")
	}
}

func runProbe(args []string) (deploymentrehearsal.ProbeOutcome, error) {
	set := flag.NewFlagSet("probe", flag.ContinueOnError)
	population := set.String("population", "", "exact R-006 population")
	candidate := set.String("candidate-bundle", "", "absolute candidate bundle path")
	previous := set.String("previous-bundle", "", "absolute previous bundle path")
	work := set.String("work-directory", "", "absolute private rehearsal work directory")
	if set.Parse(args) != nil || set.NArg() != 0 {
		return deploymentrehearsal.ProbeAccepted, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.RunProbe(deploymentrehearsal.ProbeRequest{Population: *population,
		CandidateBundle: *candidate, PreviousBundle: *previous, WorkDirectory: *work})
}

func runPlan(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("run-plan", flag.ContinueOnError)
	planPath := set.String("plan", "", "reviewed exact-command rehearsal plan")
	output := set.String("output", "", "empty result directory")
	if set.Parse(args) != nil || set.NArg() != 0 || *planPath == "" || *output == "" {
		return deploymentrehearsal.ErrInvalid
	}
	plan, err := deploymentrehearsal.LoadExecutionPlan(*planPath)
	if err != nil {
		return err
	}
	return deploymentrehearsal.ExecutePlan(ctx, plan, *output, time.Now)
}

func runValidate(args []string) (deploymentrehearsal.Evidence, error) {
	set := flag.NewFlagSet("validate", flag.ContinueOnError)
	evidence := set.String("evidence", "", "exact R-006 evidence JSON")
	if set.Parse(args) != nil || set.NArg() != 0 || *evidence == "" {
		return deploymentrehearsal.Evidence{}, deploymentrehearsal.ErrInvalid
	}
	validated, err := deploymentrehearsal.Load(*evidence)
	if err != nil {
		return deploymentrehearsal.Evidence{}, err
	}
	return validated, nil
}

func runValidateBuild(args []string) (deploymentrehearsal.BuildRecord, error) {
	set := flag.NewFlagSet("validate-build", flag.ContinueOnError)
	record := set.String("record", "", "exact release build record JSON")
	if set.Parse(args) != nil || set.NArg() != 0 || *record == "" {
		return deploymentrehearsal.BuildRecord{}, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.LoadBuildRecord(*record)
}

func fail(class string) {
	if class != "usage" && class != "invalid_input" && class != "invalid_evidence" {
		class = "operation_failed"
	}
	fmt.Fprintf(os.Stderr, "deployment rehearsal failed: error_class=%s\n", class)
	os.Exit(1)
}

func failCode(class string, code int) {
	fmt.Fprintf(os.Stderr, "deployment rehearsal failed: error_class=%s\n", class)
	os.Exit(code)
}
