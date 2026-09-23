package deploymentbrowser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"cloud-clicker/server/deploymentconfig"

	"github.com/chromedp/cdproto/network"
	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const DefaultBrowserPath = "/ms-playwright/chromium-1234/chrome-linux64/chrome"
const phase0JourneyGuard = 2 * time.Hour
const phase0ActionGuard = 20000

var hashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type Config struct {
	Origin            string
	Output            string
	ManifestSHA256    string
	BrowserPath       string
	AllowLoopbackHTTP bool
	Now               func() time.Time
}

type Result struct {
	SchemaVersion        int       `json:"schema_version"`
	ManifestSHA256       string    `json:"manifest_sha256"`
	StartedAt            time.Time `json:"started_at"`
	CompletedAt          time.Time `json:"completed_at"`
	Surface              string    `json:"surface"`
	BootstrapCommitted   bool      `json:"bootstrap_committed"`
	CredentialsPresent   bool      `json:"credentials_present"`
	WebSocketObserved    bool      `json:"websocket_observed"`
	ManualIntentObserved bool      `json:"manual_intent_observed"`
	ManualIntentStatus   int64     `json:"manual_intent_status"`
	PageExceptionCount   int       `json:"page_exception_count"`
	GateCrossed          bool      `json:"gate_crossed"`
	RunEndObserved       bool      `json:"run_end_observed"`
	NextRunObserved      bool      `json:"next_run_observed"`
	InitialRunSeq        int       `json:"initial_run_seq"`
	FinalRunSeq          int       `json:"final_run_seq"`
	ActionCount          int       `json:"action_count"`
	IntentResponseCount  int       `json:"intent_response_count"`
	FailedIntentCount    int       `json:"failed_intent_count"`
	ObjectiveCompleted   bool      `json:"objective_completed"`
	GuardExhausted       bool      `json:"guard_exhausted"`
}

func Run(ctx context.Context, config Config) (Result, error) {
	if err := validateConfig(config); err != nil {
		return Result{}, err
	}
	started := config.Now().UTC()
	options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options, chromedp.ExecPath(config.BrowserPath), chromedp.NoSandbox, chromedp.Flag("disable-dev-shm-usage", true))
	allocator, cancelAllocator := chromedp.NewExecAllocator(ctx, options...)
	defer cancelAllocator()
	browserContext, cancelBrowser := chromedp.NewContext(allocator)
	defer cancelBrowser()

	var lock sync.Mutex
	websocketObserved := false
	websocketRequests := map[network.RequestID]bool{}
	pageExceptions := 0
	intentRequests := map[network.RequestID]bool{}
	intentObserved := false
	intentStatus := int64(0)
	intentResponseCount := 0
	failedIntentCount := 0
	chromedp.ListenTarget(browserContext, func(event any) {
		lock.Lock()
		defer lock.Unlock()
		switch value := event.(type) {
		case *network.EventWebSocketCreated:
			if strings.Contains(value.URL, "/connection/websocket") {
				websocketRequests[value.RequestID] = true
			}
		case *network.EventWebSocketHandshakeResponseReceived:
			if websocketRequests[value.RequestID] {
				websocketObserved = true
			}
		case *network.EventRequestWillBeSent:
			if value.Request.Method == "POST" && strings.HasSuffix(value.Request.URL, "/api/v1/intents") {
				intentRequests[value.RequestID] = true
			}
		case *network.EventResponseReceived:
			if intentRequests[value.RequestID] {
				intentObserved = true
				intentStatus = value.Response.Status
				intentResponseCount++
				if value.Response.Status != 200 {
					failedIntentCount++
				}
				delete(intentRequests, value.RequestID)
			}
		case *cdpruntime.EventExceptionThrown:
			pageExceptions++
		}
	})

	var storageJSON string
	err := chromedp.Run(browserContext,
		network.Enable(),
		chromedp.Navigate(config.Origin),
		chromedp.WaitVisible(`//button[normalize-space()="BEGIN ATTEMPT"]`, chromedp.BySearch),
		chromedp.Click(`//button[normalize-space()="BEGIN ATTEMPT"]`, chromedp.BySearch),
		chromedp.WaitVisible(`main[data-surface="desk"]`, chromedp.ByQuery),
		chromedp.Poll(`document.body.textContent.match(/You are visitor #[0-9]+/) !== null`, nil, chromedp.WithPollingInterval(100*time.Millisecond)),
		chromedp.Evaluate(`JSON.stringify({bootstrap_committed:localStorage.getItem("cloud-clicker.bootstrap-key.v1")===null,credentials_present:localStorage.getItem("cloud-clicker.credentials.v1")!==null})`, &storageJSON),
		chromedp.Click(`//button[normalize-space()="Fix Computer"]`, chromedp.BySearch),
		chromedp.Poll(`document.querySelector("main")?.getAttribute("aria-busy")==="false"`, nil, chromedp.WithPollingInterval(100*time.Millisecond)),
		chromedp.Poll(`globalThis.performance.getEntriesByType("resource").some((entry)=>entry.name.endsWith("/api/v1/intents"))`, nil, chromedp.WithPollingInterval(100*time.Millisecond)),
	)
	if err != nil {
		return Result{}, err
	}
	var storage struct {
		BootstrapCommitted bool `json:"bootstrap_committed"`
		CredentialsPresent bool `json:"credentials_present"`
	}
	if json.Unmarshal([]byte(storageJSON), &storage) != nil {
		return Result{}, errors.New("invalid browser storage observation")
	}
	lock.Lock()
	manualObserved, manualStatus := intentObserved, intentStatus
	lock.Unlock()
	initialRunSeq, err := readRunSeq(browserContext)
	if err != nil || initialRunSeq != 1 {
		return Result{}, errors.New("invalid initial browser run")
	}
	journey, err := drivePhase0(browserContext)
	if err != nil {
		return Result{}, err
	}
	finalRunSeq, err := readRunSeq(browserContext)
	if err != nil || finalRunSeq != initialRunSeq+1 {
		return Result{}, errors.New("browser did not reach the next run")
	}
	lock.Lock()
	result := Result{SchemaVersion: 2, ManifestSHA256: config.ManifestSHA256, StartedAt: started, CompletedAt: config.Now().UTC(),
		Surface: "run_2_desk", BootstrapCommitted: storage.BootstrapCommitted, CredentialsPresent: storage.CredentialsPresent,
		WebSocketObserved: websocketObserved, ManualIntentObserved: manualObserved, ManualIntentStatus: manualStatus,
		PageExceptionCount: pageExceptions, GateCrossed: journey.gateCrossed, RunEndObserved: journey.runEndObserved,
		NextRunObserved: journey.nextRunObserved, InitialRunSeq: initialRunSeq, FinalRunSeq: finalRunSeq,
		ActionCount: journey.actionCount, IntentResponseCount: intentResponseCount, FailedIntentCount: failedIntentCount,
		ObjectiveCompleted: true}
	lock.Unlock()
	if err := ValidateResult(result); err != nil {
		return Result{}, err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return Result{}, err
	}
	if err := writeResult(config.Output, append(encoded, '\n')); err != nil {
		return Result{}, err
	}
	return result, nil
}

func ValidateResult(result Result) error {
	if result.SchemaVersion != 2 || !hashPattern.MatchString(result.ManifestSHA256) || result.StartedAt.IsZero() ||
		!result.CompletedAt.After(result.StartedAt) || result.CompletedAt.Sub(result.StartedAt) > phase0JourneyGuard ||
		result.Surface != "run_2_desk" || !result.BootstrapCommitted ||
		!result.CredentialsPresent || !result.WebSocketObserved || !result.ManualIntentObserved || result.ManualIntentStatus != 200 ||
		!result.GateCrossed || !result.RunEndObserved || !result.NextRunObserved || result.InitialRunSeq != 1 || result.FinalRunSeq != 2 ||
		result.ActionCount < 2 || result.ActionCount > phase0ActionGuard || result.IntentResponseCount < 2 || result.FailedIntentCount != 0 ||
		result.PageExceptionCount != 0 || !result.ObjectiveCompleted || result.GuardExhausted {
		return errors.New("invalid browser result")
	}
	return nil
}

func DecodeResult(data []byte) (Result, error) {
	var result Result
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateResult(result) != nil {
		return Result{}, errors.New("invalid browser result")
	}
	return result, nil
}

func validateConfig(config Config) error {
	validOrigin := deploymentconfig.ValidProductionOrigin(config.Origin)
	if config.AllowLoopbackHTTP {
		validOrigin = config.Origin == "http://127.0.0.1" || strings.HasPrefix(config.Origin, "http://127.0.0.1:") ||
			strings.HasPrefix(config.Origin, "http://host.docker.internal:")
	}
	if !validOrigin || !filepath.IsAbs(config.Output) || filepath.Clean(config.Output) != config.Output ||
		!hashPattern.MatchString(config.ManifestSHA256) || config.BrowserPath == "" || !filepath.IsAbs(config.BrowserPath) || config.Now == nil {
		return errors.New("invalid browser configuration")
	}
	return nil
}

func writeResult(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	return errors.Join(err, file.Close())
}
