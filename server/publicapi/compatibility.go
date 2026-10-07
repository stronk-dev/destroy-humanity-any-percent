package publicapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type compatibilityDocument struct {
	GeneratorVersion string                     `json:"generator_version"`
	Operations       []compatibilityOperation   `json:"operations"`
	Schemas          map[string]json.RawMessage `json:"schemas"`
}

type compatibilityOperation struct {
	ID         string   `json:"id"`
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	Surface    Surface  `json:"surface"`
	Auth       AuthMode `json:"auth"`
	Request    string   `json:"request,omitempty"`
	Parameters []string `json:"parameters"`
	Query      []string `json:"query,omitempty"`
	Responses  []string `json:"responses"`
}

type compatibilityMode uint8

const (
	compatibilityRequest compatibilityMode = 1 << iota
	compatibilityResponse
)

// CheckCompatibilityPin enforces the additive-only v1 law against a committed
// pin. New operations are allowed. Existing method/path/auth, path parameters,
// query parameters (exactly — stricter than C2 requires), request identity,
// and statuses remain; request unions/enums do not grow;
// response enums/unions may widen and response objects may add optional fields.
func CheckCompatibilityPin(prior []byte, current *Registry) error {
	var old compatibilityDocument
	if json.Unmarshal(prior, &old) != nil || old.GeneratorVersion != GeneratorVersion || old.Schemas == nil {
		return fmt.Errorf("%w: invalid compatibility pin", ErrInvalidOperation)
	}
	encoded, err := CanonicalOperationPins(current)
	if err != nil {
		return err
	}
	var next compatibilityDocument
	if json.Unmarshal(encoded, &next) != nil {
		return ErrInvalidOperation
	}
	nextOperations := map[string]compatibilityOperation{}
	for _, operation := range next.Operations {
		nextOperations[operation.ID] = operation
	}
	uses := map[string]compatibilityMode{}
	for _, before := range old.Operations {
		after, ok := nextOperations[before.ID]
		if !ok || before.Method != after.Method || before.Path != after.Path || before.Surface != after.Surface ||
			before.Auth != after.Auth || before.Request != after.Request || !equalStrings(before.Parameters, after.Parameters) || !equalStrings(before.Query, after.Query) ||
			!containsAll(after.Responses, before.Responses) {
			return fmt.Errorf("%w: incompatible operation %s", ErrInvalidOperation, before.ID)
		}
		if before.Request != "" {
			uses[before.Request] |= compatibilityRequest
		}
		for _, response := range before.Responses {
			parts := strings.Split(response, ":")
			if len(parts) == 4 && parts[1] == string(ResponseSchema) && parts[3] != "" {
				uses[parts[3]] |= compatibilityResponse
			}
		}
	}
	for name, mode := range uses {
		if err := compareNamedCompatibility(name, mode, old.Schemas, next.Schemas, map[string]bool{}); err != nil {
			return fmt.Errorf("%w: schema %s: %v", ErrInvalidOperation, name, err)
		}
	}
	return nil
}

func compareNamedCompatibility(name string, mode compatibilityMode, oldDefinitions, nextDefinitions map[string]json.RawMessage, stack map[string]bool) error {
	oldRaw, oldOK := oldDefinitions[name]
	nextRaw, nextOK := nextDefinitions[name]
	if !oldOK || !nextOK || stack[name] {
		if oldOK && nextOK && stack[name] {
			return nil
		}
		return ErrInvalidSchema
	}
	var oldSchema, nextSchema map[string]any
	// Descriptor bounds are int64. Decoding them through float64 can erase a
	// one-unit narrowing above 2^53, including at either signed int64 limit.
	oldDecoder := json.NewDecoder(bytes.NewReader(oldRaw))
	nextDecoder := json.NewDecoder(bytes.NewReader(nextRaw))
	oldDecoder.UseNumber()
	nextDecoder.UseNumber()
	if oldDecoder.Decode(&oldSchema) != nil || nextDecoder.Decode(&nextSchema) != nil {
		return ErrInvalidSchema
	}
	stack[name] = true
	err := compareCompatibilitySchema(oldSchema, nextSchema, mode, oldDefinitions, nextDefinitions, stack)
	delete(stack, name)
	return err
}

func compareCompatibilitySchema(oldSchema, nextSchema map[string]any, mode compatibilityMode,
	oldDefinitions, nextDefinitions map[string]json.RawMessage, stack map[string]bool,
) error {
	oldRef, oldHasRef := oldSchema["$ref"].(string)
	nextRef, nextHasRef := nextSchema["$ref"].(string)
	if oldHasRef || nextHasRef {
		if !oldHasRef || !nextHasRef || oldRef != nextRef {
			return ErrInvalidSchema
		}
		return compareNamedCompatibility(strings.TrimPrefix(oldRef, "#/components/schemas/"), mode, oldDefinitions, nextDefinitions, stack)
	}
	oldType, _ := oldSchema["type"].(string)
	nextType, _ := nextSchema["type"].(string)
	if oldType != nextType || stringValue(oldSchema["format"]) != stringValue(nextSchema["format"]) {
		return ErrInvalidSchema
	}
	if !compatibleBounds(oldSchema, nextSchema) || !compatibleEnum(oldSchema, nextSchema, mode) {
		return ErrInvalidSchema
	}
	if oldOne, ok := oldSchema["oneOf"].([]any); ok {
		nextOne, nextOK := nextSchema["oneOf"].([]any)
		if !nextOK || len(nextOne) < len(oldOne) || mode&compatibilityRequest != 0 && len(nextOne) != len(oldOne) {
			return ErrInvalidSchema
		}
		for index, oldArm := range oldOne {
			oldMap, oldOK := oldArm.(map[string]any)
			nextMap, nextOK := nextOne[index].(map[string]any)
			if !oldOK || !nextOK {
				return ErrInvalidSchema
			}
			if err := compareCompatibilitySchema(oldMap, nextMap, mode, oldDefinitions, nextDefinitions, stack); err != nil {
				return fmt.Errorf("oneOf[%d]: %w", index, err)
			}
		}
		return nil
	}
	if oldType == "array" {
		oldItems, oldOK := oldSchema["items"].(map[string]any)
		nextItems, nextOK := nextSchema["items"].(map[string]any)
		if !oldOK || !nextOK {
			return ErrInvalidSchema
		}
		return compareCompatibilitySchema(oldItems, nextItems, mode, oldDefinitions, nextDefinitions, stack)
	}
	if oldType != "object" {
		return nil
	}
	oldProperties, oldOK := oldSchema["properties"].(map[string]any)
	nextProperties, nextOK := nextSchema["properties"].(map[string]any)
	if !oldOK || !nextOK {
		return ErrInvalidSchema
	}
	oldRequired := stringSet(oldSchema["required"])
	nextRequired := stringSet(nextSchema["required"])
	// C2 forbids optional-to-required promotion on either stability surface,
	// whether this object is a request, a response, or shared by both.
	for name := range nextRequired {
		if !oldRequired[name] {
			return fmt.Errorf("%w: field %s became required", ErrInvalidSchema, name)
		}
	}
	for name, oldProperty := range oldProperties {
		nextProperty, ok := nextProperties[name]
		oldMap, oldOK := oldProperty.(map[string]any)
		nextMap, nextOK := nextProperty.(map[string]any)
		if !ok || !oldOK || !nextOK {
			return fmt.Errorf("%w: field %s removed or invalid", ErrInvalidSchema, name)
		}
		if err := compareCompatibilitySchema(oldMap, nextMap, mode, oldDefinitions, nextDefinitions, stack); err != nil {
			return fmt.Errorf("field %s: %w", name, err)
		}
	}
	return nil
}

func compatibleBounds(oldSchema, nextSchema map[string]any) bool {
	for _, bound := range []string{"minimum", "maximum"} {
		before, oldExists := oldSchema[bound]
		after, nextExists := nextSchema[bound]
		var oldValue, nextValue int64
		for index, item := range []struct {
			value  any
			exists bool
		}{{before, oldExists}, {after, nextExists}} {
			if !item.exists {
				continue
			}
			number, ok := item.value.(json.Number)
			if !ok {
				return false
			}
			value, err := number.Int64()
			if err != nil {
				return false
			}
			if index == 0 {
				oldValue = value
			} else {
				nextValue = value
			}
		}
		if nextExists && (!oldExists || bound == "minimum" && nextValue > oldValue || bound == "maximum" && nextValue < oldValue) {
			return false
		}
	}
	return true
}

func compatibleEnum(oldSchema, nextSchema map[string]any, mode compatibilityMode) bool {
	oldValues, oldOK := oldSchema["enum"].([]any)
	nextValues, nextOK := nextSchema["enum"].([]any)
	if !oldOK && !nextOK {
		return true
	}
	if oldOK != nextOK || mode&compatibilityRequest != 0 && len(oldValues) != len(nextValues) {
		return false
	}
	want := map[string]bool{}
	for _, value := range nextValues {
		want[stringValue(value)] = true
	}
	for _, value := range oldValues {
		if !want[stringValue(value)] {
			return false
		}
	}
	return true
}

func containsAll(values, required []string) bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	for _, value := range required {
		if !set[value] {
			return false
		}
	}
	return true
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func stringSet(value any) map[string]bool {
	result := map[string]bool{}
	values, _ := value.([]any)
	for _, item := range values {
		result[stringValue(item)] = true
	}
	return result
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
