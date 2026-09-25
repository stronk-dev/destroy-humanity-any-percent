// Package typer owns the immutable content and pure transition engine for
// Terminal Typer (rfc/minigame-terminal-typer.md TT3/TT4).
package typer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const (
	SchemaVersion = 1
	EngineVersion = "1.0.0"
	// EraTierMin/EraTierMax are the TT1 scaling row's clamp: the content
	// loader proves every reachable era_tier deals a full run.
	EraTierMin = 1
	EraTierMax = 9
)

var (
	ErrInvalidCatalog = errors.New("invalid Typer catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type Policy struct {
	RunLength          int64  `json:"run_length"`
	TimedBudgetMS      int64  `json:"timed_budget_ms"`
	MaxLineBytes       int64  `json:"max_line_bytes"`
	ElapsedHardcapMS   int64  `json:"elapsed_hardcap_ms"`
	MissesHardcap      int64  `json:"misses_hardcap"`
	RunLengthReasonKey string `json:"run_length_reason_key"`
	LineBytesReasonKey string `json:"line_bytes_reason_key"`
	ElapsedReasonKey   string `json:"elapsed_reason_key"`
	MissesReasonKey    string `json:"misses_reason_key"`
}

type Era struct {
	EraID   string `json:"era_id"`
	MinTier int64  `json:"min_tier"`
	CopyKey string `json:"copy_key"`
}

type Prompt struct {
	PromptID     string `json:"prompt_id"`
	EraID        string `json:"era_id"`
	Text         string `json:"text"`
	SceneCopyKey string `json:"scene_copy_key"`
}

type Catalog struct {
	SchemaVersion int      `json:"schema_version"`
	Policy        Policy   `json:"policy"`
	Eras          []Era    `json:"eras"`
	Prompts       []Prompt `json:"prompts"`
	eraMinTier    map[string]int64
	promptByID    map[string]Prompt
}

type Declarations struct {
	CopyKeys map[string]struct{}
}

const maxSafeInteger = 9_007_199_254_740_991

func LoadCatalog(data []byte, declarations Declarations) (*Catalog, error) {
	if len(declarations.CopyKeys) == 0 || !uniqueJSONKeys(data) ||
		!hasExactJSONKeys(data, "schema_version", "policy", "eras", "prompts") {
		return nil, ErrInvalidCatalog
	}
	var raw struct {
		Policy  json.RawMessage   `json:"policy"`
		Eras    []json.RawMessage `json:"eras"`
		Prompts []json.RawMessage `json:"prompts"`
	}
	if json.Unmarshal(data, &raw) != nil || !hasExactJSONKeys(raw.Policy, "run_length", "timed_budget_ms", "max_line_bytes",
		"elapsed_hardcap_ms", "misses_hardcap", "run_length_reason_key", "line_bytes_reason_key", "elapsed_reason_key", "misses_reason_key") {
		return nil, ErrInvalidCatalog
	}
	for _, row := range raw.Eras {
		if !hasExactJSONKeys(row, "era_id", "min_tier", "copy_key") {
			return nil, ErrInvalidCatalog
		}
	}
	for _, row := range raw.Prompts {
		if !hasExactJSONKeys(row, "prompt_id", "era_id", "text", "scene_copy_key") {
			return nil, ErrInvalidCatalog
		}
	}
	var catalog Catalog
	if strictDecode(data, &catalog) != nil || catalog.SchemaVersion != SchemaVersion || !validPolicy(catalog.Policy, declarations.CopyKeys) ||
		len(catalog.Eras) == 0 || len(catalog.Prompts) == 0 {
		return nil, ErrInvalidCatalog
	}
	catalog.eraMinTier = map[string]int64{}
	seenTiers := map[int64]bool{}
	for index, era := range catalog.Eras {
		if !idPattern.MatchString(era.EraID) || era.MinTier < 0 || era.MinTier > EraTierMax || !hasKey(declarations.CopyKeys, era.CopyKey) ||
			seenTiers[era.MinTier] || index > 0 && catalog.Eras[index-1].MinTier >= era.MinTier {
			return nil, ErrInvalidCatalog
		}
		if _, duplicate := catalog.eraMinTier[era.EraID]; duplicate {
			return nil, ErrInvalidCatalog
		}
		seenTiers[era.MinTier] = true
		catalog.eraMinTier[era.EraID] = era.MinTier
	}
	catalog.promptByID = map[string]Prompt{}
	for index, prompt := range catalog.Prompts {
		if !idPattern.MatchString(prompt.PromptID) || index > 0 && catalog.Prompts[index-1].PromptID >= prompt.PromptID ||
			!hasKey(declarations.CopyKeys, prompt.SceneCopyKey) || !validPromptText(prompt.Text, catalog.Policy.MaxLineBytes) {
			return nil, ErrInvalidCatalog
		}
		if _, ok := catalog.eraMinTier[prompt.EraID]; !ok {
			return nil, ErrInvalidCatalog
		}
		catalog.promptByID[prompt.PromptID] = prompt
	}
	// Reachability: every reachable era_tier deals a full run.
	for tier := int64(EraTierMin); tier <= EraTierMax; tier++ {
		if int64(len(catalog.Pool(tier))) < catalog.Policy.RunLength {
			return nil, ErrInvalidCatalog
		}
	}
	return &catalog, nil
}

func validPolicy(policy Policy, keys map[string]struct{}) bool {
	return policy.RunLength >= 1 && policy.RunLength <= maxSafeInteger && policy.TimedBudgetMS >= 1 && policy.TimedBudgetMS <= maxSafeInteger &&
		policy.MaxLineBytes >= 1 && policy.MaxLineBytes <= 1024 && policy.ElapsedHardcapMS >= 1 && policy.ElapsedHardcapMS <= maxSafeInteger &&
		policy.MissesHardcap >= 1 && policy.MissesHardcap <= maxSafeInteger &&
		hasKey(keys, policy.RunLengthReasonKey) && hasKey(keys, policy.LineBytesReasonKey) &&
		hasKey(keys, policy.ElapsedReasonKey) && hasKey(keys, policy.MissesReasonKey)
}

// validPromptText: 1..max bytes of printable ASCII, no leading/trailing space,
// no two consecutive spaces.
func validPromptText(text string, maxBytes int64) bool {
	if len(text) == 0 || int64(len(text)) > maxBytes || text[0] == ' ' || text[len(text)-1] == ' ' {
		return false
	}
	for index := 0; index < len(text); index++ {
		if text[index] < 0x20 || text[index] > 0x7e || text[index] == ' ' && index > 0 && text[index-1] == ' ' {
			return false
		}
	}
	return true
}

func hasKey(keys map[string]struct{}, key string) bool {
	if !idPattern.MatchString(key) {
		return false
	}
	_, ok := keys[key]
	return ok
}

// Pool is every prompt whose era has min_tier <= eraTier (cumulative, OD-6),
// in byte-sorted prompt_id order.
func (catalog *Catalog) Pool(eraTier int64) []Prompt {
	pool := []Prompt{}
	for _, prompt := range catalog.Prompts {
		if catalog.eraMinTier[prompt.EraID] <= eraTier {
			pool = append(pool, prompt)
		}
	}
	return pool
}

func (catalog *Catalog) Prompt(id string) (Prompt, bool) {
	prompt, ok := catalog.promptByID[id]
	return prompt, ok
}

func ContentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func strictDecode(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return ErrInvalidCatalog
	}
	return nil
}

func hasExactJSONKeys(data []byte, expected ...string) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil || object == nil || len(object) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func uniqueJSONKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var readValue func() bool
	readValue = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, compound := token.(json.Delim)
		if !compound {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				key, ok := keyToken.(string)
				if keyErr != nil || !ok {
					return false
				}
				if _, duplicate := seen[key]; duplicate {
					return false
				}
				seen[key] = struct{}{}
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim('}')
		case '[':
			for decoder.More() {
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim(']')
		default:
			return false
		}
	}
	if !readValue() {
		return false
	}
	_, err := decoder.Token()
	return errors.Is(err, io.EOF)
}
