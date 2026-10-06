package save

import (
	"bytes"
	"encoding/json"
)

// R5's purchase event has eight required, non-null, exact-case fields. Token
// inspection also rejects duplicate keys, which struct/map decoding would lose.
// This applies only to this event; other payload contracts remain unchanged.
func validReputationPurchaseEventFields(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') {
		return false
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen[key] {
			return false
		}
		switch key {
		case "node_id", "node_kind", "cost", "reputation_level", "reputation_spent_before", "reputation_spent_after", "unlock_ppm_after", "source":
		default:
			return false
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return false
		}
		seen[key] = true
	}
	last, err := decoder.Token()
	return err == nil && last == json.Delim('}') && len(seen) == 8 && ensureJSONEnd(decoder) == nil
}
