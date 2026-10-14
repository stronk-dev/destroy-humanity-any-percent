package economy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Resource shape is identical across the supported catalog versions. Inspect
// original row/cap names before case-insensitive struct decoding can alias them.
func validateResourceShapes(data []byte) error {
	var source struct {
		Resources []json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		return catalogError("decode", err)
	}
	for index, data := range source.Resources {
		var row map[string]json.RawMessage
		if json.Unmarshal(data, &row) != nil || !resourceShapeKeys(row, "id", "scope", "numeric_kind", "initial", "minimum", "hardcap") {
			return catalogError(fmt.Sprintf("resources[%d]", index), errors.New("resource requires exact row keys"))
		}
		if !bytes.Equal(bytes.TrimSpace(row["hardcap"]), []byte("null")) {
			var cap map[string]json.RawMessage
			if json.Unmarshal(row["hardcap"], &cap) != nil || !resourceShapeKeys(cap, "amount", "reason_key") {
				return catalogError(fmt.Sprintf("resources[%d].hardcap", index), errors.New("hardcap requires exact object keys or null"))
			}
		}
	}
	return nil
}

func resourceShapeKeys(object map[string]json.RawMessage, expected ...string) bool {
	if len(object) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, present := object[key]; !present {
			return false
		}
	}
	return true
}
