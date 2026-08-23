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
	default:
		fail("usage")
	}
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
