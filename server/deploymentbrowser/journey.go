package deploymentbrowser

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	cdpruntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

type pageState struct {
	Surface         string `json:"surface"`
	GatePresent     bool   `json:"gate_present"`
	GateEnabled     bool   `json:"gate_enabled"`
	WindDownEnabled bool   `json:"wind_down_enabled"`
	BuyMaxEnabled   bool   `json:"buy_max_enabled"`
	ManualEnabled   bool   `json:"manual_enabled"`
	DeclineEnabled  bool   `json:"decline_enabled"`
	ContinueEnabled bool   `json:"continue_enabled"`
}

type journeyResult struct {
	gateCrossed     bool
	runEndObserved  bool
	nextRunObserved bool
	actionCount     int
}

type browserAction string

// A purchase on every newly affordable snapshot would spend all cash forever
// and starve the gate requirement. This is driver pacing, not game balance.
const purchaseSpacing = 15 * time.Second

const (
	actionWait     browserAction = "wait"
	actionGate     browserAction = "gate"
	actionWindDown browserAction = "wind_down"
	actionBuyMax   browserAction = "buy_max"
	actionManual   browserAction = "manual"
	actionDecline  browserAction = "decline"
	actionContinue browserAction = "continue"
)

// This policy uses only enabled player controls. It cannot create a run by
// writing client state, calling the intent API directly or changing the clock.
func chooseAction(state pageState, gateCrossed bool) browserAction {
	switch state.Surface {
	case "offer_sheet":
		if state.DeclineEnabled {
			return actionDecline
		}
	case "run_end":
		if gateCrossed && state.ContinueEnabled {
			return actionContinue
		}
	case "desk":
		if !gateCrossed && state.GateEnabled {
			return actionGate
		}
		if gateCrossed && state.WindDownEnabled {
			return actionWindDown
		}
		if !gateCrossed && state.BuyMaxEnabled {
			return actionBuyMax
		}
		if !gateCrossed && state.ManualEnabled {
			return actionManual
		}
	}
	return actionWait
}

func drivePhase0(ctx context.Context) (journeyResult, error) {
	var result journeyResult
	deadline := time.Now().Add(phase0JourneyGuard)
	gateRequested := false
	nextPurchaseAt := time.Time{}
	for time.Now().Before(deadline) && result.actionCount < phase0ActionGuard {
		if err := ctx.Err(); err != nil {
			return journeyResult{}, err
		}
		state, err := observePage(ctx)
		if err != nil {
			return journeyResult{}, err
		}
		if gateRequested && (state.Surface == "run_end" ||
			state.Surface == "desk" && !state.GatePresent && state.WindDownEnabled) {
			result.gateCrossed = true
		}
		if state.Surface == "run_end" && result.gateCrossed {
			result.runEndObserved = true
		}
		if time.Now().Before(nextPurchaseAt) {
			state.BuyMaxEnabled = false
		}
		action := chooseAction(state, result.gateCrossed)
		if action == actionWait {
			select {
			case <-ctx.Done():
				return journeyResult{}, ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
			continue
		}
		if err := clickAction(ctx, action); err != nil {
			return journeyResult{}, err
		}
		result.actionCount++
		if action == actionGate {
			gateRequested = true
		}
		if action == actionBuyMax {
			nextPurchaseAt = time.Now().Add(purchaseSpacing)
		}
		if action == actionContinue {
			if err := chromedp.Run(ctx, chromedp.WaitVisible(`main[data-surface="desk"]`, chromedp.ByQuery)); err != nil {
				return journeyResult{}, err
			}
			result.nextRunObserved = true
			return result, nil
		}
		select {
		case <-ctx.Done():
			return journeyResult{}, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
		if err := chromedp.Run(ctx, chromedp.Poll(`document.querySelector("main")?.getAttribute("aria-busy")==="false"`, nil,
			chromedp.WithPollingInterval(100*time.Millisecond))); err != nil {
			return journeyResult{}, err
		}
	}
	return journeyResult{}, errors.New("phase-0 browser objective guard exhausted")
}

func observePage(ctx context.Context) (pageState, error) {
	const observation = `JSON.stringify((()=>{
    const main=document.querySelector("main");
    const button=(label)=>[...document.querySelectorAll("button")].find((item)=>item.textContent?.trim()===label);
    const enabled=(label)=>{const item=button(label);return !!item&&!item.disabled&&!!item.getClientRects().length};
    return {surface:main?.getAttribute("data-surface")??"",gate_present:!!button("Move Into the Garage"),
      gate_enabled:enabled("Move Into the Garage"),wind_down_enabled:enabled("Wind Down Company"),
      buy_max_enabled:enabled("Buy Max"),manual_enabled:enabled("Fix Computer"),
      decline_enabled:enabled("Decline"),continue_enabled:enabled("Start the Next Company")};
  })())`
	var encoded string
	if err := chromedp.Run(ctx, chromedp.Evaluate(observation, &encoded)); err != nil {
		return pageState{}, err
	}
	var state pageState
	if json.Unmarshal([]byte(encoded), &state) != nil || state.Surface == "" {
		return pageState{}, errors.New("invalid browser page observation")
	}
	return state, nil
}

func clickAction(ctx context.Context, action browserAction) error {
	selector := ""
	switch action {
	case actionGate:
		selector = `//button[normalize-space()="Move Into the Garage" and not(@disabled)]`
	case actionWindDown:
		selector = `//button[normalize-space()="Wind Down Company" and not(@disabled)]`
	case actionBuyMax:
		selector = `(//button[normalize-space()="Buy Max" and not(@disabled)])[last()]`
	case actionManual:
		selector = `//button[normalize-space()="Fix Computer" and not(@disabled)]`
	case actionDecline:
		selector = `//button[normalize-space()="Decline" and not(@disabled)]`
	case actionContinue:
		selector = `//button[normalize-space()="Start the Next Company" and not(@disabled)]`
	default:
		return errors.New("unknown browser action")
	}
	return chromedp.Run(ctx, chromedp.Click(selector, chromedp.BySearch))
}

func readRunSeq(ctx context.Context) (int, error) {
	const snapshot = `(async()=>{
    const stored=localStorage.getItem("cloud-clicker.credentials.v1");
    if(!stored)return 0;
    const credentials=JSON.parse(stored);
    const response=await fetch("/api/v1/founder/state",{headers:{Authorization:"Bearer "+credentials.accessToken}});
    if(!response.ok)return 0;
    const body=await response.json();
    return body?.run?.run_seq??0;
  })()`
	var seq int
	if err := chromedp.Run(ctx, chromedp.Evaluate(snapshot, &seq,
		func(params *cdpruntime.EvaluateParams) *cdpruntime.EvaluateParams {
			return params.WithAwaitPromise(true)
		})); err != nil {
		return 0, err
	}
	return seq, nil
}
