package deploymentbrowser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrowserResultRequiresEveryObservedUserBoundary(t *testing.T) {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	valid := Result{SchemaVersion: 1, ManifestSHA256: "sha256:" + strings.Repeat("a", 64), StartedAt: start,
		CompletedAt: start.Add(time.Second), Surface: "desk", BootstrapCommitted: true, CredentialsPresent: true,
		WebSocketObserved: true, ManualIntentObserved: true, ManualIntentStatus: 200, ObjectiveCompleted: true}
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
