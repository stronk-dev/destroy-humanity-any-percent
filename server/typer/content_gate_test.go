package typer

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

var updateTyperCorpus = flag.Bool("update-typer-corpus", false, "regenerate the checked-in Typer content corpus")

const corpusPath = "../../testdata/typer/content-gate-v1.json"

type corpusStep struct {
	Command      json.RawMessage `json:"command"`
	ServerTimeMS int64           `json:"server_time_ms"`
	Expect       string          `json:"expect"`
}

type corpusScenario struct {
	Name             string           `json:"name"`
	Seed             string           `json:"seed"`
	EraTier          int64            `json:"era_tier"`
	Steps            []corpusStep     `json:"steps"`
	ExpectedTerminal json.RawMessage  `json:"expected_terminal"`
	ExpectedResult   *minigame.Result `json:"expected_result"`
	CoversPrompts    []string         `json:"covers_prompts"`
}

type contentCorpus struct {
	Version          int              `json:"version"`
	TyperContentHash string           `json:"typer_content_hash"`
	TransitionBudget int              `json:"transition_budget"`
	Scenarios        []corpusScenario `json:"scenarios"`
}

// scenarioBuilder drives the real engine and records each step with its
// observed outcome, so the corpus is generated, never hand-written.
type scenarioBuilder struct {
	h        *harness
	scenario corpusScenario
	covers   map[string]bool
}

func newScenario(t *testing.T, name string, seed uint64) *scenarioBuilder {
	return &scenarioBuilder{h: newHarness(t, seed), scenario: corpusScenario{Name: name, Seed: strconv.FormatUint(seed, 10), EraTier: 1}, covers: map[string]bool{}}
}

func (b *scenarioBuilder) step(command string, at int64) string {
	b.h.t.Helper()
	if state := b.h.state(); state.CurrentPromptID != nil {
		b.covers[*state.CurrentPromptID] = true
	}
	err := b.h.apply(command, at)
	expect := "applied"
	if err != nil {
		expect = rejectionCode(err)
		if expect == "" {
			b.h.t.Fatalf("%s: non-rejection error %v", b.scenario.Name, err)
		}
	}
	b.scenario.Steps = append(b.scenario.Steps, corpusStep{Command: json.RawMessage(command), ServerTimeMS: at, Expect: expect})
	return expect
}

func (b *scenarioBuilder) finish() corpusScenario {
	b.h.t.Helper()
	if b.h.result == nil {
		b.h.t.Fatalf("%s: scenario did not reach a terminal result", b.scenario.Name)
	}
	b.scenario.ExpectedTerminal, b.scenario.ExpectedResult = b.h.snapshot, b.h.result
	b.scenario.CoversPrompts = []string{}
	for id := range b.covers {
		b.scenario.CoversPrompts = append(b.scenario.CoversPrompts, id)
	}
	sort.Strings(b.scenario.CoversPrompts)
	return b.scenario
}

// smart rewrites a prompt with the OD-7 look-alikes a mobile keyboard would
// insert; each normalizes back to the exact prompt.
func smart(text string) string {
	quote := 0
	var out strings.Builder
	for index := 0; index < len(text); index++ {
		switch {
		case text[index] == '\'':
			out.WriteString([]string{"‘", "’"}[quote%2])
			quote++
		case text[index] == '"':
			out.WriteString([]string{"“", "”"}[quote%2])
			quote++
		case strings.HasPrefix(text[index:], "--"):
			out.WriteString("—")
			index++
		case text[index] == ' ' && index == strings.IndexByte(text, ' '):
			out.WriteString(" ")
		default:
			out.WriteByte(text[index])
		}
	}
	return out.String()
}

func generateCorpus(t *testing.T) contentCorpus {
	content := fixtureBytes(t)
	catalog, err := LoadCatalog(content, declarations())
	if err != nil {
		t.Fatal(err)
	}
	corpus := contentCorpus{Version: 1, TyperContentHash: ContentHash(content)}
	covered := map[string]bool{}
	// Completed timed runs with smart punctuation until every prompt is dealt.
	for seed := uint64(1); len(covered) < len(catalog.Prompts); seed++ {
		if seed > 200 {
			t.Fatal("no seed set covers every prompt")
		}
		order := PromptOrder(catalog, seed, 1)
		fresh := false
		for _, prompt := range order {
			fresh = fresh || !covered[prompt.PromptID]
		}
		if !fresh {
			continue
		}
		b := newScenario(t, "completed_timed_seed_"+strconv.FormatUint(seed, 10), seed)
		b.step(`{"assist_level":"timed","kind":"begin"}`, 1_000)
		for index := range order {
			b.step(submit(smart(b.h.currentText())), 1_000+int64(index+1)*5_000)
		}
		scenario := b.finish()
		for _, id := range scenario.CoversPrompts {
			covered[id] = true
		}
		corpus.Scenarios = append(corpus.Scenarios, scenario)
	}
	// Untimed with misses, a case miss, every rejection code, and a backwards clock.
	b := newScenario(t, "completed_untimed_misses_and_rejections", 42)
	b.step(submit("ls"), 10)
	b.step(`{"assist_level":"sometimes","kind":"begin"}`, 10)
	b.step(`{"assist_level":"untimed","kind":"begin"}`, 100_000)
	b.step(submit("ls\u0007"), 100_001)
	b.step(submit(strings.Repeat("x", 257)), 100_002)
	b.step(submit(strings.ToUpper(b.h.currentText())), 90_000)
	b.step(submit(b.h.currentText()+"x"), 100_010)
	// A fullwidth look-alike stays a miss: no Unicode normalization (TT4.4).
	current := b.h.currentText()
	b.step(submit(string(rune(current[0])+0xfee0)+current[1:]), 100_011)
	for b.h.state().Phase == PhaseTyping {
		b.step(submit(b.h.currentText()), 100_100+int64(b.h.state().PromptIndex))
	}
	b.step(`{"kind":"end_run"}`, 200_000)
	corpus.Scenarios = append(corpus.Scenarios, b.finish())
	// Deadline exact (scored) then one ms late (timed_out, unscored).
	b = newScenario(t, "timed_out_after_exact_deadline", 77)
	b.step(`{"assist_level":"timed","kind":"begin"}`, 5_000)
	b.step(submit(b.h.currentText()), 125_000)
	b.step(submit(b.h.currentText()), 125_001)
	corpus.Scenarios = append(corpus.Scenarios, b.finish())
	// Ended early before begin: all-zero facts.
	b = newScenario(t, "ended_early_before_begin", 13)
	b.step(`{"kind":"end_run"}`, 1)
	corpus.Scenarios = append(corpus.Scenarios, b.finish())
	for _, scenario := range corpus.Scenarios {
		for _, step := range scenario.Steps {
			if step.Expect == "applied" {
				corpus.TransitionBudget++
			}
		}
	}
	return corpus
}

func TestTyperContentGate(t *testing.T) {
	corpus := generateCorpus(t)
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateTyperCorpus {
		if err := os.MkdirAll("../../testdata/typer", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(corpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	existing, err := os.ReadFile(corpusPath)
	if err != nil || !bytes.Equal(existing, encoded) {
		t.Fatalf("checked-in Typer corpus is stale; run make typer-corpus (%v)", err)
	}
	outcomes, modes, codes := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, scenario := range corpus.Scenarios {
		outcomes[scenario.ExpectedResult.Outcome] = true
		for _, step := range scenario.Steps {
			codes[step.Expect] = true
			if strings.Contains(string(step.Command), `"timed"`) {
				modes["timed"] = true
			}
			if strings.Contains(string(step.Command), `"untimed"`) {
				modes["untimed"] = true
			}
		}
	}
	if len(outcomes) != 3 || len(modes) != 2 {
		t.Fatalf("corpus must cover all outcomes and both modes: %v %v", outcomes, modes)
	}
	for _, code := range errorTaxonomy {
		if !codes[code] {
			t.Fatalf("corpus misses rejection %s", code)
		}
	}
}
