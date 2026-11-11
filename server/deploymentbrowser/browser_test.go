package deploymentbrowser

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
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
	withoutFailedCount := bytes.Replace(data, []byte(`"failed_intent_count":0,`), nil, 1)
	if bytes.Equal(withoutFailedCount, data) {
		t.Fatal("fixture lacks the failure count")
	}
	for name, mutated := range map[string][]byte{
		"unknown":              append(append([]byte(nil), data[:len(data)-1]...), []byte(`,"summary":"passed"}`)...),
		"trailing":             append(append([]byte(nil), data...), []byte(` {}`)...),
		"invalid":              []byte(`{"schema_version":1}`),
		"missing failed count": withoutFailedCount,
		"null page exceptions": bytes.Replace(data, []byte(`"page_exception_count":0`), []byte(`"page_exception_count":null`), 1),
		"null guard status":    bytes.Replace(data, []byte(`"guard_exhausted":false`), []byte(`"guard_exhausted":null`), 1),
		"case alias":           bytes.Replace(data, []byte(`"guard_exhausted"`), []byte(`"Guard_Exhausted"`), 1),
		"overwritten failure":  append([]byte(`{"failed_intent_\u0063ount":1,`), data[1:]...),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeResult(mutated); err == nil {
				t.Fatal("invalid browser evidence accepted")
			}
		})
	}
	escaped := bytes.Replace(data, []byte(`"guard_exhausted"`), []byte(`"guard_\u0065xhausted"`), 1)
	if _, err := DecodeResult(escaped); err != nil {
		t.Fatalf("valid escaped member rejected: %v", err)
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

func TestJourneyBuysFromVisibleEnabledRows(t *testing.T) {
	browser := os.Getenv("CLOUD_CLICKER_TEST_CHROMIUM")
	if browser == "" {
		t.Skip("set CLOUD_CLICKER_TEST_CHROMIUM to execute the browser driver regression")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	allocator, cancelAllocator := chromedp.NewExecAllocator(ctx, append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browser), chromedp.NoSandbox,
	)...)
	defer cancelAllocator()
	page, cancelPage := chromedp.NewContext(allocator)
	defer cancelPage()

	const html = `<main data-surface="desk">
		<button disabled>Buy Max</button>
		<button id="earlier" onclick="document.body.dataset.clicked=this.id">Buy Max</button>
		<button id="eligible" onclick="document.body.dataset.clicked=this.id">Buy Max</button>
		<button aria-disabled="true" onclick="document.body.dataset.clicked='pending'">Buy Max</button>
		<button id="hidden" hidden onclick="document.body.dataset.clicked=this.id">Buy Max</button>
		<button>Fix Computer</button>
	</main>`
	if err := chromedp.Run(page, chromedp.Navigate("data:text/html;base64,"+base64.StdEncoding.EncodeToString([]byte(html)))); err != nil {
		t.Fatal(err)
	}
	state, err := observePage(page)
	if err != nil || chooseAction(state, false) != actionBuyMax {
		t.Fatalf("visible eligible purchase was missed: state=%+v err=%v", state, err)
	}
	if err := clickAction(page, actionBuyMax); err != nil {
		t.Fatal(err)
	}
	var clicked string
	if err := chromedp.Run(page, chromedp.Evaluate("document.body.dataset.clicked", &clicked)); err != nil || clicked != "eligible" {
		t.Fatalf("clicked=%q err=%v", clicked, err)
	}
	if err := chromedp.Run(page,
		chromedp.SetAttributeValue("#earlier", "disabled", "", chromedp.ByQuery),
		chromedp.SetAttributeValue("#eligible", "disabled", "", chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}
	state, err = observePage(page)
	if err != nil || chooseAction(state, false) != actionManual {
		t.Fatalf("hidden or disabled purchase counted as actionable: state=%+v err=%v", state, err)
	}
}
