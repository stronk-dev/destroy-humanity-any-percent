package deploymentbrowser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrowserResultRequiresEveryObservedUserBoundary(t *testing.T) {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	valid := validJourneyResult(start)
	if err := ValidateResult(valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*Result){
		"bootstrap retained":   func(value *Result) { value.BootstrapCommitted = false },
		"no credentials":       func(value *Result) { value.CredentialsPresent = false },
		"no websocket":         func(value *Result) { value.WebSocketObserved = false },
		"no intent":            func(value *Result) { value.ManualIntentObserved = false },
		"intent failed":        func(value *Result) { value.ManualIntentStatus = 500 },
		"page exception":       func(value *Result) { value.PageExceptionCount = 1 },
		"gate absent":          func(value *Result) { value.GateCrossed = false },
		"terminal absent":      func(value *Result) { value.RunEndObserved = false },
		"next run absent":      func(value *Result) { value.NextRunObserved = false },
		"wrong run":            func(value *Result) { value.FinalRunSeq = 1 },
		"no actions":           func(value *Result) { value.ActionCount = 0 },
		"no intent responses":  func(value *Result) { value.IntentResponseCount = 0 },
		"failed intent":        func(value *Result) { value.FailedIntentCount = 1 },
		"elapsed guard":        func(value *Result) { value.CompletedAt = start.Add(phase0JourneyGuard + time.Second) },
		"objective incomplete": func(value *Result) { value.ObjectiveCompleted = false },
		"guard exhausted":      func(value *Result) { value.GuardExhausted = true },
	} {
		t.Run(name, func(t *testing.T) {
			value := valid
			mutate(&value)
			if err := ValidateResult(value); err == nil {
				t.Fatal("invalid browser result accepted")
			}
		})
	}
}

func TestBrowserConfigAndResultOutputFailClosed(t *testing.T) {
	config := Config{Origin: "https://play.example.test", Output: filepath.Join(t.TempDir(), "result.json"),
		ManifestSHA256: "sha256:" + strings.Repeat("a", 64), BrowserPath: DefaultBrowserPath, Now: time.Now}
	if err := validateConfig(config); err != nil {
		t.Fatal(err)
	}
	config.Origin = "http://play.example.test"
	if err := validateConfig(config); err == nil {
		t.Fatal("insecure production origin accepted")
	}
	if err := writeResult(config.Output, []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	if err := writeResult(config.Output, []byte("{}\n")); err == nil {
		t.Fatal("browser result overwritten")
	}
	info, err := os.Stat(config.Output)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("result mode=%v err=%v", info.Mode(), err)
	}
}

func TestDecodeBrowserResultRejectsUnknownTrailingAndInvalidEvidence(t *testing.T) {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	valid := validJourneyResult(start)
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeResult(data); err != nil {
		t.Fatal(err)
	}
	for name, mutated := range map[string][]byte{
		"unknown":  append(append([]byte(nil), data[:len(data)-1]...), []byte(`,"summary":"passed"}`)...),
		"trailing": append(append([]byte(nil), data...), []byte(` {}`)...),
		"invalid":  []byte(`{"schema_version":1}`),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeResult(mutated); err == nil {
				t.Fatal("invalid browser evidence accepted")
			}
		})
	}
}

func validJourneyResult(start time.Time) Result {
	return Result{SchemaVersion: 2, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), StartedAt: start,
		CompletedAt: start.Add(time.Second), Surface: "run_2_desk", BootstrapCommitted: true, CredentialsPresent: true,
		WebSocketObserved: true, ManualIntentObserved: true, ManualIntentStatus: 200, GateCrossed: true,
		RunEndObserved: true, NextRunObserved: true, InitialRunSeq: 1, FinalRunSeq: 2,
		ActionCount: 3, IntentResponseCount: 3, ObjectiveCompleted: true}
}

func TestJourneySelectsOnlyEnabledPlayerControls(t *testing.T) {
	for _, sample := range []struct {
		name    string
		state   pageState
		crossed bool
		want    browserAction
	}{
		{name: "offer interruption", state: pageState{Surface: "offer_sheet", DeclineEnabled: true}, want: actionDecline},
		{name: "disabled offer", state: pageState{Surface: "offer_sheet"}, want: actionWait},
		{name: "gate before purchase", state: pageState{Surface: "desk", GateEnabled: true, BuyMaxEnabled: true}, want: actionGate},
		{name: "buy before manual", state: pageState{Surface: "desk", BuyMaxEnabled: true, ManualEnabled: true}, want: actionBuyMax},
		{name: "manual", state: pageState{Surface: "desk", ManualEnabled: true}, want: actionManual},
		{name: "exit after gate", state: pageState{Surface: "desk", WindDownEnabled: true, BuyMaxEnabled: true}, crossed: true, want: actionWindDown},
		{name: "do not exit before gate", state: pageState{Surface: "desk", WindDownEnabled: true}, want: actionWait},
		{name: "terminal continuation", state: pageState{Surface: "run_end", ContinueEnabled: true}, crossed: true, want: actionContinue},
		{name: "no blind continuation", state: pageState{Surface: "run_end", ContinueEnabled: true}, want: actionWait},
	} {
		t.Run(sample.name, func(t *testing.T) {
			if got := chooseAction(sample.state, sample.crossed); got != sample.want {
				t.Fatalf("action=%s want=%s", got, sample.want)
			}
		})
	}
}
