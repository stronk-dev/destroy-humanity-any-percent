package evidencejson

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// DecodeFlatRecord requires every JSON-tagged field of the record exactly once.
// A missing or null observation must not become a successful zero value.
func DecodeFlatRecord[T any](data []byte, record *T) bool {
	shape := reflect.TypeFor[T]()
	remaining := make(map[string]struct{}, shape.NumField())
	for field := range shape.Fields() {
		remaining[field.Tag.Get("json")] = struct{}{}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	for decoder.More() {
		member, err := decoder.Token()
		key, _ := member.(string)
		if _, ok := remaining[key]; err != nil || !ok {
			return false
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(value, []byte("null")) {
			return false
		}
		delete(remaining, key)
	}
	end, err := decoder.Token()
	return err == nil && end == json.Delim('}') && len(remaining) == 0 && json.Unmarshal(data, record) == nil
}
