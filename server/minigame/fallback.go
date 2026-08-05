package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
)

var ErrInvalidFallbackPolicy = errors.New("invalid minigame fallback policy")

type FallbackKind string

const (
	FallbackSolo       FallbackKind = "solo"
	FallbackBot        FallbackKind = "bot"
	FallbackNPCPartner FallbackKind = "npc_partner"
)

type PolicyIdentity struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

type FallbackPolicy struct {
	Kind             FallbackKind
	BotRef           *PolicyIdentity
	NPCProfile       *PolicyIdentity
	RateReductionPPM int64
}

type fallbackSoloWire struct {
	Kind FallbackKind `json:"kind"`
}

type fallbackBotWire struct {
	Kind             FallbackKind    `json:"kind"`
	BotRef           botIdentityWire `json:"bot_ref"`
	RateReductionPPM int64           `json:"rate_reduction_ppm"`
}

type fallbackNPCWire struct {
	Kind             FallbackKind    `json:"kind"`
	NPCProfile       npcIdentityWire `json:"npc_profile"`
	RateReductionPPM int64           `json:"rate_reduction_ppm"`
}

type botIdentityWire struct {
	PolicyID string `json:"policy_id"`
	Version  string `json:"version"`
}

type npcIdentityWire struct {
	ProfileID string `json:"profile_id"`
	Version   string `json:"version"`
}

func LoadFallbackPolicy(data []byte) (FallbackPolicy, error) {
	var raw map[string]json.RawMessage
	var kind FallbackKind
	if !uniqueJSONKeys(data) || json.Unmarshal(data, &raw) != nil || raw == nil || json.Unmarshal(raw["kind"], &kind) != nil {
		return FallbackPolicy{}, ErrInvalidFallbackPolicy
	}
	switch kind {
	case FallbackSolo:
		if !hasExactJSONKeys(data, "kind") {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		var wire fallbackSoloWire
		if decodeExact(data, &wire) != nil {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		return FallbackPolicy{Kind: FallbackSolo}, nil
	case FallbackBot:
		if !hasExactJSONKeys(data, "bot_ref", "kind", "rate_reduction_ppm") {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		var wire fallbackBotWire
		if decodeExact(data, &wire) != nil || !validPolicyIdentity(wire.BotRef.PolicyID, wire.BotRef.Version) || !validReduction(wire.RateReductionPPM) {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		return FallbackPolicy{Kind: FallbackBot, BotRef: &PolicyIdentity{ID: wire.BotRef.PolicyID, Version: wire.BotRef.Version}, RateReductionPPM: wire.RateReductionPPM}, nil
	case FallbackNPCPartner:
		if !hasExactJSONKeys(data, "kind", "npc_profile", "rate_reduction_ppm") {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		var wire fallbackNPCWire
		if decodeExact(data, &wire) != nil || !validPolicyIdentity(wire.NPCProfile.ProfileID, wire.NPCProfile.Version) || !validReduction(wire.RateReductionPPM) {
			return FallbackPolicy{}, ErrInvalidFallbackPolicy
		}
		return FallbackPolicy{Kind: FallbackNPCPartner, NPCProfile: &PolicyIdentity{ID: wire.NPCProfile.ProfileID, Version: wire.NPCProfile.Version}, RateReductionPPM: wire.RateReductionPPM}, nil
	default:
		return FallbackPolicy{}, ErrInvalidFallbackPolicy
	}
}

func validPolicyIdentity(id, version string) bool {
	return mechanicalPattern.MatchString(id) && versionPattern.MatchString(version)
}

func validReduction(value int64) bool {
	return value >= 0 && value <= 1_000_000
}

func hasExactJSONKeys(data []byte, expected ...string) bool {
	var raw map[string]json.RawMessage
	if json.Unmarshal(data, &raw) != nil || raw == nil || len(raw) != len(expected) {
		return false
	}
	actual := make([]string, 0, len(raw))
	for key := range raw {
		actual = append(actual, key)
	}
	sort.Strings(actual)
	sort.Strings(expected)
	for index := range actual {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}

func decodeExact(data []byte, destination any) error {
	if !uniqueJSONKeys(data) {
		return ErrInvalidFallbackPolicy
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return ErrInvalidFallbackPolicy
	}
	return nil
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
	var trailing any
	return errors.Is(decoder.Decode(&trailing), io.EOF)
}
