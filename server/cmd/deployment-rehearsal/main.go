package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"cloud-clicker/server/deploymentbrowser"
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
			failCode("fixture_rejected", deploymentrehearsal.ProbeRejectedExit)
		}
		fmt.Println("deployment rehearsal probe passed")
	case "supply-chain":
		if _, err := runSupplyChain(os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal supply chain passed")
	case "forge-proof":
		if _, err := runForgeProof(os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal forgery proof passed")
	case "observe-host":
		if _, err := runObserveHost(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal host observation passed")
	case "install-candidate":
		if err := runInstallCandidate(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal candidate install passed")
	case "restart-admitted-work":
		outcome, err := runScenarioProbe(context.Background(), "restart-admitted-work", os.Args[2:], deploymentrehearsal.ProbeRestartDuringAdmittedWork)
		if err != nil {
			failCode("invalid_probe", 2)
		}
		if outcome == deploymentrehearsal.ProbeRejected {
			failCode("fixture_rejected", deploymentrehearsal.ProbeRejectedExit)
		}
		fmt.Println("deployment rehearsal restart was accepted as a drain")
	case "restore-non-clean":
		outcome, err := runNonCleanRestore(context.Background(), os.Args[2:])
		if err != nil {
			failCode("invalid_probe", 2)
		}
		if outcome == deploymentrehearsal.ProbeRejected {
			failCode("fixture_rejected", deploymentrehearsal.ProbeRejectedExit)
		}
		fmt.Println("deployment rehearsal non-clean restore was accepted")
	case "lifecycle-release":
		if err := runScenarioStep(context.Background(), "lifecycle-release", os.Args[2:], deploymentrehearsal.ReleaseLifecycle); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal lifecycle release passed")
	case "lifecycle-rollback":
		if err := runScenarioStep(context.Background(), "lifecycle-rollback", os.Args[2:], deploymentrehearsal.RollbackLifecycle); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal lifecycle rollback passed")
	case "run-browser":
		if _, err := runBrowser(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal product browser passed")
	case "recover-empty":
		if _, err := runEmptyRecovery(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal empty recovery passed")
	case "recover-populated":
		if _, err := runPopulatedRecovery(context.Background(), os.Args[2:]); err != nil {
			fail("invalid_evidence")
		}
		fmt.Println("deployment rehearsal populated recovery passed")
	default:
		fail("usage")
	}
}

func runEmptyRecovery(ctx context.Context, args []string) (deploymentrehearsal.RecoveryCheckpoint, error) {
	config, err := recoveryConfig(args)
	if err != nil {
		return deploymentrehearsal.RecoveryCheckpoint{}, err
	}
	return deploymentrehearsal.RunEmptyRecovery(ctx, config)
}

func runPopulatedRecovery(ctx context.Context, args []string) (deploymentrehearsal.ObjectiveObservation, error) {
	config, err := recoveryConfig(args)
	if err != nil {
		return deploymentrehearsal.ObjectiveObservation{}, err
	}
	return deploymentrehearsal.RunPopulatedRecovery(ctx, config)
}

func recoveryConfig(args []string) (deploymentrehearsal.ScenarioConfig, error) {
	set := flag.NewFlagSet("recovery", flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.ScenarioConfig{}, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.LoadScenarioConfig(*configPath)
}

func runBrowser(ctx context.Context, args []string) (deploymentbrowser.Result, error) {
	set := flag.NewFlagSet("run-browser", flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentbrowser.Result{}, deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return deploymentbrowser.Result{}, err
	}
	return deploymentrehearsal.RunBrowser(ctx, config)
}

func runScenarioProbe(ctx context.Context, name string, args []string, probe func(context.Context, deploymentrehearsal.ScenarioConfig) (deploymentrehearsal.ProbeOutcome, error)) (deploymentrehearsal.ProbeOutcome, error) {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.ProbeAccepted, deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return deploymentrehearsal.ProbeAccepted, err
	}
	return probe(ctx, config)
}

func runNonCleanRestore(ctx context.Context, args []string) (deploymentrehearsal.ProbeOutcome, error) {
	set := flag.NewFlagSet("restore-non-clean", flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.ProbeAccepted, deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return deploymentrehearsal.ProbeAccepted, err
	}
	return deploymentrehearsal.ProbeNonCleanRestore(ctx, config)
}

func runScenarioStep(ctx context.Context, name string, args []string, step func(context.Context, deploymentrehearsal.ScenarioConfig) error) error {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return err
	}
	return step(ctx, config)
}

func runInstallCandidate(ctx context.Context, args []string) error {
	set := flag.NewFlagSet("install-candidate", flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return err
	}
	return deploymentrehearsal.InstallCandidate(ctx, config)
}

func runObserveHost(ctx context.Context, args []string) (deploymentrehearsal.HostObservation, error) {
	set := flag.NewFlagSet("observe-host", flag.ContinueOnError)
	configPath := set.String("config", "", "private runtime scenario input")
	if set.Parse(args) != nil || set.NArg() != 0 || *configPath == "" {
		return deploymentrehearsal.HostObservation{}, deploymentrehearsal.ErrInvalid
	}
	config, err := deploymentrehearsal.LoadScenarioConfig(*configPath)
	if err != nil {
		return deploymentrehearsal.HostObservation{}, err
	}
	return deploymentrehearsal.ObserveHost(ctx, config)
}

func runForgeProof(args []string) (deploymentrehearsal.ForgeryProof, error) {
	set := flag.NewFlagSet("forge-proof", flag.ContinueOnError)
	base := set.String("base-evidence", "", "validated 42-population base dossier")
	plan := set.String("plan", "", "reviewed exact-command rehearsal plan")
	results := set.String("results", "", "exclusive per-population result directory")
	artifacts := set.String("artifacts", "", "exclusive retained run-artifact directory")
	candidateBundle := set.String("candidate-bundle", "", "exact candidate release bundle")
	output := set.String("output", "", "exclusive forgery-proof result")
	if set.Parse(args) != nil || set.NArg() != 0 {
		return deploymentrehearsal.ForgeryProof{}, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.ProduceForgeryProof(deploymentrehearsal.ForgeryRequest{BaseEvidence: *base, Plan: *plan,
		Results: *results, Artifacts: *artifacts, CandidateBundle: *candidateBundle, Output: *output, Now: time.Now})
}

func runSupplyChain(args []string) (deploymentrehearsal.SupplyChainResult, error) {
	set := flag.NewFlagSet("supply-chain", flag.ContinueOnError)
	candidateBundle := set.String("candidate-bundle", "", "exact candidate bundle")
	previousBundle := set.String("previous-bundle", "", "exact previous bundle")
	candidateBuild := set.String("candidate-build", "", "candidate independent-build record")
	previousBuild := set.String("previous-build", "", "previous independent-build record")
	secretScan := set.String("secret-scan", "", "structured source/image scan")
	output := set.String("output", "", "exclusive supply-chain result")
	if set.Parse(args) != nil || set.NArg() != 0 {
		return deploymentrehearsal.SupplyChainResult{}, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.ObserveSupplyChain(deploymentrehearsal.SupplyChainRequest{CandidateBundle: *candidateBundle,
		PreviousBundle: *previousBundle, CandidateBuild: *candidateBuild, PreviousBuild: *previousBuild,
		SecretScan: *secretScan, Output: *output, Now: time.Now})
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
	plan := set.String("plan", "", "reviewed exact-command rehearsal plan")
	results := set.String("results", "", "exclusive per-population result directory")
	artifacts := set.String("artifacts", "", "exclusive retained run-artifact directory")
	candidateBundle := set.String("candidate-bundle", "", "exact candidate release bundle")
	seal := set.String("seal", "", "exclusive base-evidence and forgery-proof directory")
	if set.Parse(args) != nil || set.NArg() != 0 || *evidence == "" || *plan == "" || *results == "" || *artifacts == "" || *candidateBundle == "" || *seal == "" {
		return deploymentrehearsal.Evidence{}, deploymentrehearsal.ErrInvalid
	}
	return deploymentrehearsal.LoadAndValidateRun(*evidence, *plan, *results, *artifacts, *candidateBundle, *seal)
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
