package deploymentrehearsal

import (
	"context"
	"errors"
	"testing"

	"cloud-clicker/server/deploymentrelease"
)

type restartFake struct {
	evidence   deploymentrelease.DrainEvidence
	exitCode   string
	startErr   error
	killed     bool
	restarted  bool
	smokeCalls int
}

func (fake *restartFake) KillGameserver(context.Context, deploymentrelease.Bundle) (deploymentrelease.DrainEvidence, string, error) {
	fake.killed = true
	return fake.evidence, fake.exitCode, nil
}
func (fake *restartFake) Start(context.Context, deploymentrelease.Bundle) error {
	fake.restarted = true
	return fake.startErr
}
func (fake *restartFake) AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error {
	fake.smokeCalls++
	return nil
}

func TestRestartDuringAdmittedWorkRequiresARejectedDrainAndRestoredService(t *testing.T) {
	valid := deploymentrelease.DrainEvidence{ReadinessDown: true, CourtesyFrame: true, IntentsRefused: true, AdmittedComplete: true,
		JobsFlushed: true, SocketsClosed: true, WithinBound: true}
	run := func(t *testing.T, fake *restartFake, installed bool) (ProbeOutcome, error) {
		t.Helper()
		fixture := newLifecycleFixture(t)
		if installed {
			fixture.installCandidate(t)
		}
		return probeRestartDuringAdmittedWork(context.Background(), fixture.config, restartDependencies{loadBundle: fixture.dependencies.loadBundle,
			runtime: func(ScenarioConfig) restartRuntime { return fake }})
	}
	killed := &restartFake{exitCode: "137"}
	if outcome, err := run(t, killed, true); err != nil || outcome != ProbeRejected || !killed.restarted || killed.smokeCalls != 1 {
		t.Fatalf("SIGKILL restart outcome=%d err=%v fake=%+v", outcome, err, killed)
	}
	if outcome, err := run(t, &restartFake{evidence: valid, exitCode: "0"}, true); err != nil || outcome != ProbeAccepted {
		t.Fatalf("graceful drain counted as rejected: outcome=%d err=%v", outcome, err)
	}
	if outcome, err := run(t, &restartFake{exitCode: "137", startErr: errors.New("stack did not return")}, true); err == nil || outcome == ProbeRejected {
		t.Fatalf("unrestored host counted as rejection: outcome=%d err=%v", outcome, err)
	}
	notInstalled := &restartFake{exitCode: "137"}
	if outcome, err := run(t, notInstalled, false); err == nil || outcome == ProbeRejected || notInstalled.killed {
		t.Fatalf("restart probe ran without an installed candidate: outcome=%d err=%v", outcome, err)
	}
}
