package typer

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/minigame"
)

func fixtureBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/typer-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func declarations() Declarations {
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return Declarations{CopyKeys: keys}
}

func mutateFixture(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(fixtureBytes(t), &value); err != nil {
		t.Fatal(err)
	}
	mutate(value)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTyperCatalogLoadsFixtureAndRejectsDefects(t *testing.T) {
	catalog, err := LoadCatalog(fixtureBytes(t), declarations())
	if err != nil || len(catalog.Prompts) < 12 {
		t.Fatalf("fixture must load with at least 12 prompts: %v", err)
	}
	prompt := func(value map[string]any) map[string]any { return value["prompts"].([]any)[0].(map[string]any) }
	cases := map[string]func(map[string]any){
		"unknown root key":     func(v map[string]any) { v["extra"] = 1 },
		"unknown policy key":   func(v map[string]any) { v["policy"].(map[string]any)["extra"] = 1 },
		"non-ascii prompt":     func(v map[string]any) { prompt(v)["text"] = "ls é" },
		"double space prompt":  func(v map[string]any) { prompt(v)["text"] = "ls  -la" },
		"trailing space":       func(v map[string]any) { prompt(v)["text"] = "ls " },
		"unknown era":          func(v map[string]any) { prompt(v)["era_id"] = "nowhere" },
		"unknown scene key":    func(v map[string]any) { prompt(v)["scene_copy_key"] = "typer.prompt.missing.scene" },
		"pool below run":       func(v map[string]any) { v["policy"].(map[string]any)["run_length"] = 99 },
		"unsorted prompts":     func(v map[string]any) { p := v["prompts"].([]any); p[0], p[1] = p[1], p[0] },
		"line bytes above cap": func(v map[string]any) { v["policy"].(map[string]any)["max_line_bytes"] = 1025 },
	}
	for name, mutate := range cases {
		if _, err := LoadCatalog(mutateFixture(t, mutate), declarations()); !errors.Is(err, ErrInvalidCatalog) {
			t.Fatalf("%s: expected rejection, got %v", name, err)
		}
	}
}

func TestTyperCreateRejectsScalingBelowPinnedClamp(t *testing.T) {
	content := mutateFixture(t, func(value map[string]any) {
		value["eras"].([]any)[0].(map[string]any)["min_tier"] = float64(0)
	})
	if _, err := LoadCatalog(content, declarations()); err != nil {
		t.Fatalf("tier-zero era is valid content and must reach scaling validation: %v", err)
	}
	_, err := NewTenant().Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 1,
		ScalingInputs: map[string]int64{ScalingDestination: 0}, Content: content,
		ContentHash: ContentHash(content), ContentSchemaVersion: SchemaVersion})
	if !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Fatalf("tier-zero scaling bypassed TT1 clamp_min=1: %v", err)
	}
	validContent := fixtureBytes(t)
	valid, err := NewTenant().Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 1,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: validContent,
		ContentHash: ContentHash(validContent), ContentSchemaVersion: SchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	if err := json.Unmarshal(valid, &snapshot); err != nil {
		t.Fatal(err)
	}
	snapshot["era_tier"] = float64(0)
	mutated, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewTenant().ValidateSnapshot(mutated); !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Fatalf("tier-zero snapshot bypassed TT1 clamp_min=1: %v", err)
	}
}

type harness struct {
	t        *testing.T
	content  []byte
	hash     string
	seed     uint64
	snapshot json.RawMessage
	revision int64
	result   *minigame.Result
}

func newHarness(t *testing.T, seed uint64) *harness {
	t.Helper()
	content := fixtureBytes(t)
	h := &harness{t: t, content: content, hash: ContentHash(content), seed: seed}
	snapshot, err := NewTenant().Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: seed,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: content, ContentHash: h.hash, ContentSchemaVersion: SchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	h.snapshot, h.revision = snapshot, 1
	return h
}

func (h *harness) apply(command string, serverTimeMS int64) error {
	h.t.Helper()
	output, err := NewTenant().Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: h.seed, Revision: h.revision, Snapshot: h.snapshot,
		Command: json.RawMessage(command), ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: h.content, ContentHash: h.hash,
		ContentSchemaVersion: SchemaVersion, ServerTimeMs: serverTimeMS})
	if err != nil {
		return err
	}
	h.snapshot, h.revision, h.result = output.Snapshot, h.revision+1, output.Result
	return nil
}

func (h *harness) state() Snapshot {
	h.t.Helper()
	value, err := decodeSnapshot(h.snapshot)
	if err != nil {
		h.t.Fatal(err)
	}
	return value
}

func (h *harness) currentText() string { return *h.state().CurrentPromptText }

func submit(text string) string {
	encoded, _ := json.Marshal(map[string]string{"kind": "submit_line", "text": text})
	return string(encoded)
}

func rejectionCode(err error) string {
	var rejection *minigame.Rejection
	if errors.As(err, &rejection) {
		return rejection.Code
	}
	return ""
}

func fact(result *minigame.Result, kind string) int64 {
	for _, value := range result.ScoreFacts {
		if value.Kind == kind {
			return value.Value
		}
	}
	return -1
}

func TestTyperDeadlineBoundary(t *testing.T) {
	for _, offset := range []int64{0, 1} {
		h := newHarness(t, 7)
		if err := h.apply(`{"assist_level":"timed","kind":"begin"}`, 1_000); err != nil {
			t.Fatal(err)
		}
		deadline := *h.state().DeadlineServerMS
		if deadline != 1_000+120_000 {
			t.Fatalf("deadline %d", deadline)
		}
		if err := h.apply(submit(h.currentText()), deadline+offset); err != nil {
			t.Fatal(err)
		}
		state := h.state()
		if offset == 0 && (state.LinesCleared != 1 || state.Phase != PhaseTyping) {
			t.Fatalf("submit exactly at the deadline must be scored: %+v", state)
		}
		if offset == 1 && (state.LinesCleared != 0 || h.result == nil || h.result.Outcome != OutcomeTimedOut || fact(h.result, "typer.elapsed_ms") != 120_000) {
			t.Fatalf("submit one ms late must end timed_out unscored: %+v %+v", state, h.result)
		}
	}
}

func TestTyperClockNeverRewinds(t *testing.T) {
	h := newHarness(t, 3)
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 50_000); err != nil {
		t.Fatal(err)
	}
	if err := h.apply(submit("wrong"), 10_000); err != nil {
		t.Fatal(err)
	}
	if last := *h.state().LastServerMS; last != 50_000 {
		t.Fatalf("backwards stamp rewound last_server_ms to %d", last)
	}
	if err := h.apply(`{"kind":"end_run"}`, 20_000); err != nil {
		t.Fatal(err)
	}
	if elapsed := fact(h.result, "typer.elapsed_ms"); elapsed != 0 || NewTenant().ValidateResult(h.result) != nil {
		t.Fatalf("elapsed %d must be non-negative and valid", elapsed)
	}
	negative := *h.result
	negative.ScoreFacts = append([]minigame.ScoreFact(nil), h.result.ScoreFacts...)
	negative.ScoreFacts[2].Value = -1
	if NewTenant().ValidateResult(&negative) == nil {
		t.Fatal("ValidateResult must reject a negative elapsed fact")
	}
}

func TestTyperComparisonAndNormalization(t *testing.T) {
	cases := []struct {
		submitted, target string
		cleared           bool
		index             int64
	}{
		{"ls -la", "ls -la", true, 0},
		{"LS -la", "ls -la", false, 0},
		{"ls -l", "ls -la", false, 5},
		{"ls -lah", "ls -la", false, 6},
		{"echo ‘hello world’ > index.html", "echo 'hello world' > index.html", true, 0},
		{"grep “404” access_log", "grep \"404\" access_log", true, 0},
		{"df -h", "df -h", true, 0},
		{"./configure —prefix=/usr/local", "./configure --prefix=/usr/local", true, 0},
		{"./configure –prefix=/usr/local", "./configure --prefix=/usr/local", false, 12},
		{"\uff4c\uff53 -la", "ls -la", false, 0},
	}
	for _, testCase := range cases {
		normalized := Normalize(testCase.submitted)
		if (normalized == testCase.target) != testCase.cleared {
			t.Fatalf("%q vs %q: cleared mismatch", testCase.submitted, testCase.target)
		}
		if !testCase.cleared && FirstMismatchIndex(normalized, testCase.target) != testCase.index {
			t.Fatalf("%q: mismatch index %d want %d", testCase.submitted, FirstMismatchIndex(normalized, testCase.target), testCase.index)
		}
	}
}

func TestTyperRejectionsMutateNothing(t *testing.T) {
	h := newHarness(t, 11)
	if code := rejectionCode(h.apply(submit("ls"), 1)); code != "illegal_phase" {
		t.Fatalf("submit before begin: %q", code)
	}
	if code := rejectionCode(h.apply(`{"assist_level":"fast","kind":"begin"}`, 1)); code != "invalid_assist_level" {
		t.Fatalf("bad assist: %q", code)
	}
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	before, revision := string(h.snapshot), h.revision
	for text, code := range map[string]string{"ls\u0007": "invalid_text", "ls\u007f": "invalid_text", strings.Repeat("a", 257): "line_too_long"} {
		if got := rejectionCode(h.apply(submit(text), 99)); got != code {
			t.Fatalf("%q: got %q want %q", text, got, code)
		}
	}
	for _, command := range []string{`{"kind":"submit_line","text":"ls","server_ms":5}`, `{"kind":"submit_line","text":"ls","score":5}`,
		`{"kind":"end_run","elapsed_ms":1}`, `{"kind":"submit_line","prompt_id":"ls_la","text":"ls"}`} {
		if h.apply(command, 99) == nil {
			t.Fatalf("authority field accepted: %s", command)
		}
	}
	if string(h.snapshot) != before || h.revision != revision {
		t.Fatal("a rejected command mutated the snapshot or advanced the revision")
	}
}

func TestTyperSubmitLineDistinguishesValidReplacementRuneFromMalformedEncoding(t *testing.T) {
	for _, data := range [][]byte{
		[]byte(`{"kind":"submit_line","text":"�"}`),
		[]byte(`{"kind":"submit_line","text":"\ufffd"}`),
		[]byte(`{"kind":"submit_line","text":"\ud83d\ude42"}`),
		[]byte(`{"kind":"submit_line","text":"\\ud800"}`),
	} {
		if _, err := decodeCommand(data); err != nil {
			t.Fatalf("valid Unicode line %s rejected: %v", data, err)
		}
	}
	malformed := append([]byte(`{"kind":"submit_line","text":"`), 0xff)
	malformed = append(malformed, []byte(`"}`)...)
	for _, data := range [][]byte{
		malformed,
		[]byte(`{"kind":"submit_line","text":"\ud800"}`),
		[]byte(`{"kind":"submit_line","text":"\udc00"}`),
		[]byte(`{"kind":"submit_line","text":"\ud800\u0041"}`),
	} {
		_, err := decodeCommand(data)
		if code := rejectionCode(err); code != "invalid_text" {
			t.Fatalf("malformed Unicode line %s: got %q, want invalid_text", data, code)
		}
	}
}

func TestTyperCleanLineAccounting(t *testing.T) {
	h := newHarness(t, 5)
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	if err := h.apply(submit("nope"), 2); err != nil {
		t.Fatal(err)
	}
	if err := h.apply(submit(h.currentText()), 3); err != nil {
		t.Fatal(err)
	}
	if state := h.state(); state.LinesCleared != 1 || state.CleanLines != 0 || state.Misses != 1 || state.CurrentPromptMisses != 0 {
		t.Fatalf("miss-then-clear accounting: %+v", state)
	}
	if err := h.apply(submit(h.currentText()), 4); err != nil {
		t.Fatal(err)
	}
	if state := h.state(); state.LinesCleared != 2 || state.CleanLines != 1 {
		t.Fatalf("clean clear accounting: %+v", state)
	}
}

func TestTyperSnapshotNeverContainsAFuturePrompt(t *testing.T) {
	h := newHarness(t, 9)
	catalog, _ := LoadCatalog(h.content, declarations())
	order := PromptOrder(catalog, 9, 1)
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	for _, future := range order[1:] {
		if strings.Contains(string(h.snapshot), `"`+future.PromptID+`"`) {
			t.Fatalf("snapshot leaks future prompt %s", future.PromptID)
		}
	}
}

func TestTyperSnapshotRejectsMalformedState(t *testing.T) {
	h := newHarness(t, 42)
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"negative misses":        func(row map[string]any) { row["misses"] = -1 },
		"negative clean lines":   func(row map[string]any) { row["clean_lines"] = -1 },
		"negative prompt misses": func(row map[string]any) { row["current_prompt_misses"] = -1 },
		"unknown assist level":   func(row map[string]any) { row["assist_level"] = "custom" },
		"malformed submission": func(row map[string]any) {
			row["last_submission"] = map[string]any{"outcome": "miss", "first_mismatch_index": nil}
		},
	} {
		var row map[string]any
		if err := json.Unmarshal(h.snapshot, &row); err != nil {
			t.Fatal(err)
		}
		mutate(row)
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if err := NewTenant().ValidateSnapshot(encoded); err == nil {
			t.Fatalf("%s: Go tenant admitted malformed snapshot", name)
		}
	}
	var overCap map[string]any
	if err := json.Unmarshal(h.snapshot, &overCap); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(h.content, declarations())
	if err != nil {
		t.Fatal(err)
	}
	overCap["misses"] = float64(catalog.Policy.MissesHardcap + 1)
	encoded, err := json.Marshal(overCap)
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewTenant().Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: h.seed, Revision: h.revision,
		Snapshot: encoded, Command: json.RawMessage(`{"kind":"end_run"}`), ScalingInputs: map[string]int64{ScalingDestination: 1},
		Content: h.content, ContentHash: h.hash, ContentSchemaVersion: SchemaVersion, ServerTimeMs: 2})
	if !errors.Is(err, minigame.ErrTenantDivergence) {
		t.Fatalf("above-cap misses should diverge against pinned content: %v", err)
	}
}

func TestTyperTypingSnapshotRequiresRemainingPrompt(t *testing.T) {
	h := newHarness(t, 42)
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	for h.state().PromptIndex < h.state().PromptsTotal-1 {
		if err := h.apply(submit(h.currentText()), h.revision); err != nil {
			t.Fatal(err)
		}
	}
	last := h.state()
	if err := NewTenant().ValidateSnapshot(h.snapshot); err != nil {
		t.Fatalf("valid final prompt refused: %v", err)
	}
	registry, err := minigame.NewTenantRegistry(NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	invalid := last
	invalid.PromptIndex, invalid.LinesCleared = last.PromptsTotal, last.PromptsTotal
	encoded, err := encodeSnapshot(invalid)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewTenant().ValidateSnapshot(encoded); !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Errorf("typing at prompts_total must be refused: %v", err)
	}
	input := minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: h.seed, Revision: h.revision,
		Snapshot: encoded, Command: json.RawMessage(`{"kind":"end_run"}`),
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: h.content,
		ContentHash: h.hash, ContentSchemaVersion: SchemaVersion, ServerTimeMs: h.revision}
	for name, apply := range map[string]func(minigame.ApplyInput) (minigame.ApplyOutput, error){
		"tenant": NewTenant().Apply,
		"registry": func(value minigame.ApplyInput) (minigame.ApplyOutput, error) {
			return registry.Apply(EngineRef, EngineVersion, value)
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if failure := recover(); failure != nil {
					t.Fatalf("malformed typing snapshot panicked instead of refusing: %v", failure)
				}
			}()
			before := append([]byte(nil), input.Snapshot...)
			output, err := apply(input)
			if !errors.Is(err, minigame.ErrTenantDivergence) || len(output.Snapshot) != 0 || output.Result != nil {
				t.Fatalf("malformed snapshot must diverge without output: %+v %v", output, err)
			}
			if !bytes.Equal(input.Snapshot, before) {
				t.Fatal("refusal mutated the input snapshot")
			}
		})
	}
	// Genuine final prompt still completes through the registry; no hardcoded
	// provisional run length or rejection of every typing state can satisfy this.
	input.Snapshot, input.Command = h.snapshot, json.RawMessage(submit(h.currentText()))
	output, err := registry.Apply(EngineRef, EngineVersion, input)
	if err != nil || output.Result == nil || output.Result.Outcome != OutcomeCompleted ||
		fact(output.Result, "typer.lines_cleared") != last.PromptsTotal ||
		fact(output.Result, "typer.clean_lines") != last.PromptsTotal {
		t.Fatalf("valid final prompt did not complete: %+v %v", output, err)
	}
	if err := NewTenant().ValidateSnapshot(output.Snapshot); err != nil {
		t.Fatalf("valid terminal output refused: %v", err)
	}
}

func TestTyperSnapshotRequiresNumericValuesAndExactFeedback(t *testing.T) {
	h := newHarness(t, 42)
	type state struct {
		name     string
		snapshot json.RawMessage
		revision int64
	}
	states := []state{{PhaseReady, h.snapshot, h.revision}}
	if err := h.apply(`{"assist_level":"untimed","kind":"begin"}`, 1); err != nil {
		t.Fatal(err)
	}
	if err := h.apply(submit(h.currentText()), 2); err != nil {
		t.Fatal(err)
	}
	states = append(states, state{PhaseTyping, h.snapshot, h.revision})
	if err := h.apply(`{"kind":"end_run"}`, 3); err != nil {
		t.Fatal(err)
	}
	states = append(states, state{PhaseTerminal, h.snapshot, h.revision})
	registry, err := minigame.NewTenantRegistry(NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	apply := map[string]func(minigame.ApplyInput) (minigame.ApplyOutput, error){
		"tenant": NewTenant().Apply,
		"registry": func(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
			return registry.Apply(EngineRef, EngineVersion, input)
		},
	}
	for _, valid := range states {
		t.Run(valid.name, func(t *testing.T) {
			input := minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: h.seed, Revision: valid.revision,
				Snapshot: valid.snapshot, Command: json.RawMessage(`{"kind":"end_run"}`),
				ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: h.content,
				ContentHash: h.hash, ContentSchemaVersion: SchemaVersion, ServerTimeMs: 10}
			if err := NewTenant().ValidateSnapshot(valid.snapshot); err != nil {
				t.Fatalf("valid %s snapshot refused: %v", valid.name, err)
			}
			for path, run := range apply {
				output, err := run(input)
				if valid.name == PhaseTerminal {
					if rejectionCode(err) != "illegal_phase" {
						t.Fatalf("%s valid terminal must reach its phase gate: %v", path, err)
					}
				} else if err != nil || output.Result == nil || output.Result.Outcome != OutcomeEndedEarly {
					t.Fatalf("%s valid %s did not end: %+v %v", path, valid.name, output, err)
				}
			}
			fields := []string{"typer_schema_version", "era_tier", "prompt_index", "prompts_total", "current_prompt_misses", "lines_cleared", "clean_lines", "misses", "revision"}
			if valid.name != PhaseReady {
				fields = append(fields, "missing feedback index")
			}
			for _, field := range fields {
				t.Run(field, func(t *testing.T) {
					var row map[string]json.RawMessage
					if err := json.Unmarshal(valid.snapshot, &row); err != nil {
						t.Fatal(err)
					}
					if field == "missing feedback index" {
						row["last_submission"] = json.RawMessage(`{"outcome":"cleared"}`)
					} else {
						row[field] = json.RawMessage(`null`)
					}
					encoded, err := json.Marshal(row)
					if err != nil {
						t.Fatal(err)
					}
					if err := NewTenant().ValidateSnapshot(encoded); !errors.Is(err, minigame.ErrInvalidTenant) {
						t.Errorf("malformed %s admitted: %v", field, err)
					}
					invalid := input
					invalid.Snapshot = encoded
					before := append([]byte(nil), encoded...)
					for path, run := range apply {
						output, err := run(invalid)
						if !errors.Is(err, minigame.ErrTenantDivergence) || len(output.Snapshot) != 0 || output.Result != nil {
							t.Errorf("%s malformed %s must refuse without output: %+v %v", path, field, output, err)
						}
						if !bytes.Equal(invalid.Snapshot, before) {
							t.Fatalf("%s refusal mutated the snapshot", path)
						}
					}
				})
			}
		})
	}
}

func TestTyperIsolation(t *testing.T) {
	for _, path := range []string{"engine.go", "catalog.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"time.Now", "\"time\"", "company.", "economy", "ledger", "production"} {
			if strings.Contains(string(source), forbidden) {
				t.Fatalf("%s references %q: the engine must be pure and economy-free", path, forbidden)
			}
		}
	}
}

func TestTyperSnapshotsAreRegistryCanonical(t *testing.T) {
	registry, err := minigame.NewTenantRegistry(NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	content := fixtureBytes(t)
	input := minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 21, ScalingInputs: map[string]int64{ScalingDestination: 1},
		Content: content, ContentHash: ContentHash(content), ContentSchemaVersion: SchemaVersion}
	snapshot, err := registry.Create(EngineRef, EngineVersion, input)
	if err != nil {
		t.Fatal(err)
	}
	revision := int64(1)
	for _, command := range []string{`{"assist_level":"untimed","kind":"begin"}`, submit("miss"), `{"kind":"end_run"}`} {
		output, err := registry.Apply(EngineRef, EngineVersion, minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: 21, Revision: revision,
			Snapshot: snapshot, Command: json.RawMessage(command), ScalingInputs: input.ScalingInputs, Content: content,
			ContentHash: input.ContentHash, ContentSchemaVersion: SchemaVersion, ServerTimeMs: 10 * revision})
		if err != nil {
			t.Fatalf("%s: registry rejected tenant output: %v", command, err)
		}
		snapshot, revision = output.Snapshot, revision+1
	}
}
