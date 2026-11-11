package evidencejson

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
)

// DecodeFlatRecord requires every JSON-tagged field of the record exactly once.
// A missing or null observation must not become a successful zero value.
func DecodeFlatRecord[T any](data []byte, record *T) bool {
	return DecodeRecord(data, record)
}

// DecodeRecord checks required observations recursively before Go can fold key
// case, overwrite duplicates, or supply zero values for missing/null members.
// An omitempty field may be absent; when present it must still be unambiguous.
func DecodeRecord[T any](data []byte, record *T) bool {
	return validShape(data, reflect.TypeFor[T]()) && json.Unmarshal(data, record) == nil
}

func validShape(data []byte, shape reflect.Type) bool {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return false
	}
	if shape.Kind() == reflect.Pointer {
		return validShape(data, shape.Elem())
	}
	// JSON scalar owners such as time.Time validate their own representation.
	if reflect.PointerTo(shape).Implements(reflect.TypeFor[json.Unmarshaler]()) {
		return true
	}
	switch shape.Kind() {
	case reflect.Struct:
		return validObject(data, shape)
	case reflect.Slice, reflect.Array:
		var elements []json.RawMessage
		if json.Unmarshal(data, &elements) != nil || shape.Kind() == reflect.Array && len(elements) != shape.Len() {
			return false
		}
		for _, element := range elements {
			if !validShape(element, shape.Elem()) {
				return false
			}
		}
	case reflect.Map, reflect.Interface:
		// Evidence DTOs have a declared shape, not open-ended observations.
		return false
	}
	return true
}

func validObject(data []byte, shape reflect.Type) bool {
	fields := make(map[string]reflect.Type, shape.NumField())
	required := make(map[string]bool, shape.NumField())
	for field := range shape.Fields() {
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" || field.PkgPath != "" || fields[name] != nil {
			return false
		}
		fields[name] = field.Type
		if !strings.Contains(","+options+",", ",omitempty,") {
			required[name] = true
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	for decoder.More() {
		member, err := decoder.Token()
		key, _ := member.(string)
		field, ok := fields[key]
		if err != nil || !ok {
			return false
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil || !validShape(value, field) {
			return false
		}
		delete(fields, key)
		delete(required, key)
	}
	end, err := decoder.Token()
	return err == nil && end == json.Delim('}') && len(required) == 0
}
