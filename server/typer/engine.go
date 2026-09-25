package typer

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/minigame"
)

const (
	EngineRef          = "typer"
	RunSubstream       = "typer.run.v1"
	PromptSubstream    = "typer.prompts.v1"
	ScalingDestination = "typer.era_tier"

	PhaseReady    = "ready"
	PhaseTyping   = "typing"
	PhaseTerminal = "terminal"

	AssistTimed   = "timed"
	AssistUntimed = "untimed"

	OutcomeCompleted  = "completed"
	OutcomeEndedEarly = "ended_early"
	OutcomeTimedOut   = "timed_out"

	SubmissionCleared = "cleared"
	SubmissionMiss    = "miss"
)

var errorTaxonomy = []string{"illegal_phase", "invalid_assist_level", "invalid_text", "line_too_long"}

// resultFactKinds is the byte-sorted certified fact set (TT4.7).
var resultFactKinds = []string{"typer.assisted", "typer.clean_lines", "typer.elapsed_ms", "typer.lines_cleared", "typer.misses"}

// Submission fields are declared in byte order so the nested object is
// canonical (the platform re-canonicalizes every level).
type Submission struct {
	FirstMismatchIndex *int64 `json:"first_mismatch_index"`
	Outcome            string `json:"outcome"`
}

// Snapshot is typer.snapshot.v1: exactly eighteen keys (TT4.5).
type Snapshot struct {
	TyperContentHash    string      `json:"typer_content_hash"`
	TyperSchemaVersion  int         `json:"typer_schema_version"`
	Phase               string      `json:"phase"`
	EraTier             int64       `json:"era_tier"`
	AssistLevel         *string     `json:"assist_level"`
	PromptIndex         int64       `json:"prompt_index"`
	PromptsTotal        int64       `json:"prompts_total"`
	CurrentPromptID     *string     `json:"current_prompt_id"`
	CurrentPromptText   *string     `json:"current_prompt_text"`
	CurrentPromptMisses int64       `json:"current_prompt_misses"`
	LinesCleared        int64       `json:"lines_cleared"`
	CleanLines          int64       `json:"clean_lines"`
	Misses              int64       `json:"misses"`
	StartedServerMS     *int64      `json:"started_server_ms"`
	DeadlineServerMS    *int64      `json:"deadline_server_ms"`
	LastServerMS        *int64      `json:"last_server_ms"`
	LastSubmission      *Submission `json:"last_submission"`
	Revision            int64       `json:"revision"`
}

type Tenant struct{}

func NewTenant() Tenant { return Tenant{} }

func (Tenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: EngineRef, EngineVersion: EngineVersion,
		CommandSchema: "typer.command.v1", SnapshotSchema: "typer.snapshot.v1", ResultSchema: "minigame.result.v1",
		Modes: []minigame.Mode{minigame.ModeSolo}, ErrorTaxonomy: append([]string(nil), errorTaxonomy...),
		Destinations: map[string]minigame.DestinationClass{ScalingDestination: minigame.DestinationBreadth}}
}

func (Tenant) ValidateCommand(data json.RawMessage) error {
	_, rejection := decodeCommand(data)
	return rejection
}

func (Tenant) ValidateSnapshot(data json.RawMessage) error {
	_, err := decodeSnapshot(data)
	return err
}

// ValidateResult is content-free: it checks the closed outcome and fact sets
// and the structural bounds. run_length and hardcaps are enforced by the
// engine, which holds the content.
func (Tenant) ValidateResult(result *minigame.Result) error {
	if result == nil {
		return nil
	}
	if result.Outcome != OutcomeCompleted && result.Outcome != OutcomeEndedEarly && result.Outcome != OutcomeTimedOut ||
		result.RatingDelta != nil || len(result.ScoreFacts) != len(resultFactKinds) {
		return minigame.ErrInvalidTenant
	}
	values := map[string]int64{}
	for index, fact := range result.ScoreFacts {
		if fact.Kind != resultFactKinds[index] || fact.Value < 0 {
			return minigame.ErrInvalidTenant
		}
		values[fact.Kind] = fact.Value
	}
	if values["typer.assisted"] > 1 || values["typer.clean_lines"] > values["typer.lines_cleared"] {
		return minigame.ErrInvalidTenant
	}
	return nil
}

func (Tenant) Create(input minigame.CreateInput) (json.RawMessage, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo {
		return nil, minigame.ErrInvalidTenant
	}
	eraTier, ok := validScaling(input.ScalingInputs, catalog)
	if !ok {
		return nil, minigame.ErrInvalidTenant
	}
	return encodeSnapshot(Snapshot{TyperContentHash: input.ContentHash, TyperSchemaVersion: input.ContentSchemaVersion,
		Phase: PhaseReady, EraTier: eraTier, PromptsTotal: catalog.Policy.RunLength, Revision: 1})
}

func (Tenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo {
		return minigame.ApplyOutput{}, minigame.ErrInvalidTenant
	}
	eraTier, ok := validScaling(input.ScalingInputs, catalog)
	if !ok {
		return minigame.ApplyOutput{}, minigame.ErrInvalidTenant
	}
	snapshot, err := decodeSnapshot(input.Snapshot)
	if err != nil || snapshot.Revision != input.Revision || snapshot.TyperContentHash != input.ContentHash ||
		snapshot.TyperSchemaVersion != input.ContentSchemaVersion || snapshot.EraTier != eraTier ||
		validateSnapshotAgainstCatalog(snapshot, catalog, input.Seed) != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	command, rejection := decodeCommand(input.Command)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	result, rejection := transition(&snapshot, command, catalog, input.Seed, input.ServerTimeMs)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	snapshot.Revision = input.Revision + 1
	encoded, err := encodeSnapshot(snapshot)
	if err != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	return minigame.ApplyOutput{Snapshot: encoded, Result: result}, nil
}

func transition(snapshot *Snapshot, command command, catalog *Catalog, seed uint64, serverTimeMS int64) (*minigame.Result, error) {
	switch command.Kind {
	case "begin":
		if snapshot.Phase != PhaseReady {
			return nil, reject("illegal_phase", "begin requires ready phase")
		}
		t := advanceClock(snapshot, serverTimeMS)
		assist := command.AssistLevel
		snapshot.AssistLevel = &assist
		snapshot.StartedServerMS = int64Pointer(t)
		if assist == AssistTimed {
			snapshot.DeadlineServerMS = int64Pointer(t + catalog.Policy.TimedBudgetMS)
		}
		snapshot.Phase = PhaseTyping
		setCurrentPrompt(snapshot, PromptOrder(catalog, seed, snapshot.EraTier)[0])
		return nil, nil
	case "submit_line":
		if snapshot.Phase != PhaseTyping {
			return nil, reject("illegal_phase", "submit_line requires typing phase")
		}
		if int64(len(command.Text)) > catalog.Policy.MaxLineBytes {
			return nil, reject("line_too_long", "submitted line exceeds max_line_bytes")
		}
		t := advanceClock(snapshot, serverTimeMS)
		if snapshot.DeadlineServerMS != nil && t > *snapshot.DeadlineServerMS {
			return terminate(snapshot, catalog, OutcomeTimedOut), nil
		}
		normalized := Normalize(command.Text)
		target := *snapshot.CurrentPromptText
		if normalized == target {
			snapshot.LinesCleared++
			if snapshot.CurrentPromptMisses == 0 {
				snapshot.CleanLines++
			}
			snapshot.PromptIndex++
			snapshot.CurrentPromptMisses = 0
			snapshot.LastSubmission = &Submission{Outcome: SubmissionCleared}
			if snapshot.PromptIndex == catalog.Policy.RunLength {
				return terminate(snapshot, catalog, OutcomeCompleted), nil
			}
			setCurrentPrompt(snapshot, PromptOrder(catalog, seed, snapshot.EraTier)[snapshot.PromptIndex])
			return nil, nil
		}
		index := FirstMismatchIndex(normalized, target)
		snapshot.Misses = saturatingIncrement(snapshot.Misses, catalog.Policy.MissesHardcap)
		snapshot.CurrentPromptMisses = saturatingIncrement(snapshot.CurrentPromptMisses, catalog.Policy.MissesHardcap)
		snapshot.LastSubmission = &Submission{Outcome: SubmissionMiss, FirstMismatchIndex: &index}
		return nil, nil
	case "end_run":
		if snapshot.Phase != PhaseReady && snapshot.Phase != PhaseTyping {
			return nil, reject("illegal_phase", "end_run requires a non-terminal phase")
		}
		advanceClock(snapshot, serverTimeMS)
		return terminate(snapshot, catalog, OutcomeEndedEarly), nil
	default:
		return nil, reject("illegal_phase", "unknown command kind")
	}
}

// advanceClock applies t = max(sample, last_server_ms): a clock step
// backwards never rewinds the run.
func advanceClock(snapshot *Snapshot, sample int64) int64 {
	t := sample
	if snapshot.LastServerMS != nil && *snapshot.LastServerMS > t {
		t = *snapshot.LastServerMS
	}
	snapshot.LastServerMS = int64Pointer(t)
	return t
}

func setCurrentPrompt(snapshot *Snapshot, prompt Prompt) {
	id, text := prompt.PromptID, prompt.Text
	snapshot.CurrentPromptID, snapshot.CurrentPromptText = &id, &text
}

func terminate(snapshot *Snapshot, catalog *Catalog, outcome string) *minigame.Result {
	snapshot.Phase = PhaseTerminal
	snapshot.CurrentPromptID, snapshot.CurrentPromptText = nil, nil
	return &minigame.Result{Outcome: outcome, RatingDelta: nil, ScoreFacts: []minigame.ScoreFact{
		{Kind: "typer.assisted", Value: assistedFact(snapshot)},
		{Kind: "typer.clean_lines", Value: snapshot.CleanLines},
		{Kind: "typer.elapsed_ms", Value: elapsedFact(snapshot, catalog)},
		{Kind: "typer.lines_cleared", Value: snapshot.LinesCleared},
		{Kind: "typer.misses", Value: snapshot.Misses},
	}}
}

func assistedFact(snapshot *Snapshot) int64 {
	if snapshot.AssistLevel != nil && *snapshot.AssistLevel == AssistUntimed {
		return 1
	}
	return 0
}

func elapsedFact(snapshot *Snapshot, catalog *Catalog) int64 {
	if snapshot.StartedServerMS == nil || snapshot.LastServerMS == nil {
		return 0
	}
	end := *snapshot.LastServerMS
	if snapshot.DeadlineServerMS != nil && *snapshot.DeadlineServerMS < end {
		end = *snapshot.DeadlineServerMS
	}
	elapsed := end - *snapshot.StartedServerMS
	if elapsed > catalog.Policy.ElapsedHardcapMS {
		elapsed = catalog.Policy.ElapsedHardcapMS
	}
	return elapsed
}

func saturatingIncrement(value, hardcap int64) int64 {
	if value >= hardcap {
		return hardcap
	}
	return value + 1
}

// Normalize applies OD-7's closed smart-punctuation map and nothing else.
func Normalize(text string) string {
	return strings.NewReplacer("‘", "'", "’", "'", "“", "\"", "”", "\"", " ", " ", "—", "--").Replace(text)
}

// FirstMismatchIndex is the first differing byte offset, or the shorter
// length when one is a prefix of the other.
func FirstMismatchIndex(submitted, target string) int64 {
	limit := len(submitted)
	if len(target) < limit {
		limit = len(target)
	}
	for index := 0; index < limit; index++ {
		if submitted[index] != target[index] {
			return int64(index)
		}
	}
	return int64(limit)
}

// PromptOrder is TT4.2: a downward Fisher–Yates over the eligible pool with
// the typer.prompts.v1 substream of the typer.run.v1 run seed.
func PromptOrder(catalog *Catalog, seed uint64, eraTier int64) []Prompt {
	pool := catalog.Pool(eraTier)
	runSeed := determinism.Substream(seed, RunSubstream).Next()
	random := determinism.Substream(runSeed, PromptSubstream)
	for index := len(pool) - 1; index > 0; index-- {
		swap := int(random.Bound(uint64(index + 1)))
		pool[index], pool[swap] = pool[swap], pool[index]
	}
	return pool[:catalog.Policy.RunLength]
}

type command struct {
	Kind        string
	AssistLevel string
	Text        string
}

func decodeCommand(data []byte) (command, error) {
	if !uniqueJSONKeys(data) {
		return command{}, reject("illegal_phase", "command keys are not unique")
	}
	var header struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(data, &header) != nil {
		return command{}, reject("illegal_phase", "command is not an object")
	}
	switch header.Kind {
	case "begin":
		var wire struct {
			Kind        string `json:"kind"`
			AssistLevel string `json:"assist_level"`
		}
		if !hasExactJSONKeys(data, "kind", "assist_level") || strictDecode(data, &wire) != nil ||
			wire.AssistLevel != AssistTimed && wire.AssistLevel != AssistUntimed {
			return command{}, reject("invalid_assist_level", "begin requires assist_level timed or untimed")
		}
		return command{Kind: header.Kind, AssistLevel: wire.AssistLevel}, nil
	case "submit_line":
		var wire struct {
			Kind string `json:"kind"`
			Text string `json:"text"`
		}
		if !hasExactJSONKeys(data, "kind", "text") || strictDecode(data, &wire) != nil {
			return command{}, reject("invalid_text", "submit_line schema mismatch")
		}
		if !validSubmittedText(wire.Text) {
			return command{}, reject("invalid_text", "submitted line contains invalid characters")
		}
		return command{Kind: header.Kind, Text: wire.Text}, nil
	case "end_run":
		if !hasExactJSONKeys(data, "kind") {
			return command{}, reject("illegal_phase", "end_run schema mismatch")
		}
		return command{Kind: header.Kind}, nil
	default:
		return command{}, reject("illegal_phase", "unknown command kind")
	}
}

// validSubmittedText is TT4.4 step 1: valid UTF-8 with no C0 control
// characters or U+007F.
func validSubmittedText(text string) bool {
	if !utf8.ValidString(text) {
		return false
	}
	for _, character := range text {
		if character < 0x20 || character == 0x7f || character == utf8.RuneError {
			return false
		}
	}
	return true
}

func decodeSnapshot(data []byte) (Snapshot, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "typer_content_hash", "typer_schema_version", "phase", "era_tier", "assist_level",
		"prompt_index", "prompts_total", "current_prompt_id", "current_prompt_text", "current_prompt_misses", "lines_cleared", "clean_lines",
		"misses", "started_server_ms", "deadline_server_ms", "last_server_ms", "last_submission", "revision") {
		return Snapshot{}, minigame.ErrInvalidTenant
	}
	var value Snapshot
	if strictDecode(data, &value) != nil || value.TyperSchemaVersion != SchemaVersion ||
		!strings.HasPrefix(value.TyperContentHash, "sha256:") || len(value.TyperContentHash) != 71 ||
		value.Phase != PhaseReady && value.Phase != PhaseTyping && value.Phase != PhaseTerminal ||
		value.EraTier < 0 || value.EraTier > EraTierMax || value.Revision < 1 || value.PromptsTotal < 1 ||
		value.PromptIndex < 0 || value.PromptIndex > value.PromptsTotal || value.CurrentPromptMisses < 0 ||
		value.LinesCleared != value.PromptIndex || value.CleanLines < 0 || value.CleanLines > value.LinesCleared || value.Misses < 0 ||
		value.AssistLevel != nil && *value.AssistLevel != AssistTimed && *value.AssistLevel != AssistUntimed ||
		(value.CurrentPromptID == nil) != (value.CurrentPromptText == nil) ||
		(value.Phase == PhaseTyping) != (value.CurrentPromptID != nil) ||
		(value.Phase == PhaseReady) != (value.AssistLevel == nil && value.StartedServerMS == nil && value.LastServerMS == nil && value.PromptIndex == 0) && value.Phase != PhaseTerminal ||
		(value.AssistLevel == nil) != (value.StartedServerMS == nil) ||
		(value.DeadlineServerMS != nil) != (value.AssistLevel != nil && *value.AssistLevel == AssistTimed) {
		return Snapshot{}, minigame.ErrInvalidTenant
	}
	if value.LastSubmission != nil && (!hasExactSubmissionShape(value.LastSubmission)) {
		return Snapshot{}, minigame.ErrInvalidTenant
	}
	return value, nil
}

func hasExactSubmissionShape(value *Submission) bool {
	switch value.Outcome {
	case SubmissionCleared:
		return value.FirstMismatchIndex == nil
	case SubmissionMiss:
		return value.FirstMismatchIndex != nil && *value.FirstMismatchIndex >= 0
	default:
		return false
	}
}

func validateSnapshotAgainstCatalog(value Snapshot, catalog *Catalog, seed uint64) error {
	if value.PromptsTotal != catalog.Policy.RunLength || value.Misses > catalog.Policy.MissesHardcap ||
		value.CurrentPromptMisses > catalog.Policy.MissesHardcap {
		return minigame.ErrInvalidTenant
	}
	if value.Phase == PhaseTyping {
		expected := PromptOrder(catalog, seed, value.EraTier)[value.PromptIndex]
		if *value.CurrentPromptID != expected.PromptID || *value.CurrentPromptText != expected.Text {
			return minigame.ErrInvalidTenant
		}
	}
	return nil
}

func encodeSnapshot(snapshot Snapshot) (json.RawMessage, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var canonical map[string]json.RawMessage
	if json.Unmarshal(encoded, &canonical) != nil {
		return nil, minigame.ErrTenantDivergence
	}
	return json.Marshal(canonical)
}

func catalogForInput(data []byte, hash string, schemaVersion int) (*Catalog, error) {
	if len(data) == 0 || hash != ContentHash(data) || schemaVersion != SchemaVersion {
		return nil, ErrInvalidCatalog
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return LoadCatalog(data, Declarations{CopyKeys: keys})
}

func validScaling(values map[string]int64, catalog *Catalog) (int64, bool) {
	if len(values) != 1 {
		return 0, false
	}
	tier, ok := values[ScalingDestination]
	if !ok || tier < 0 || tier > EraTierMax || int64(len(catalog.Pool(tier))) < catalog.Policy.RunLength {
		return 0, false
	}
	return tier, true
}

func int64Pointer(value int64) *int64 { return &value }

func reject(code, detail string) *minigame.Rejection {
	return &minigame.Rejection{Code: code, Detail: detail}
}
