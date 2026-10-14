package publicapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
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
	// C2/C18 allow response widening. A formerly single closed object may gain
	// terminal alternatives while preserving its original arm and all old bytes.
	// oneOf needs more than a superset: an overlapping new arm could make old
	// responses match twice and become invalid. Accept only demonstrably disjoint
	// closed object arms here; unknown/ref/nested-union cases stay conservative.
	if _, alreadyUnion := oldSchema["oneOf"]; !alreadyUnion {
		if alternatives, widened := nextSchema["oneOf"].([]any); widened {
			if mode&compatibilityRequest != 0 || len(alternatives) < 2 {
				return ErrInvalidSchema
			}
			preserved, ok := alternatives[0].(map[string]any)
			if !ok {
				return ErrInvalidSchema
			}
			if err := compareCompatibilitySchema(oldSchema, preserved, mode, oldDefinitions, nextDefinitions, stack); err != nil {
				return err
			}
			for _, alternative := range alternatives[1:] {
				other, ok := alternative.(map[string]any)
				if !ok || !disjointClosedCompatibilityObjects(preserved, other) {
					return ErrInvalidSchema
				}
			}
			return nil
		}
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
		if err := preserveCompatibilityStringUnion(oldOne, nextOne, oldDefinitions, nextDefinitions); err != nil {
			return err
		}
		if err := preserveCompatibilityIntegerUnion(oldOne, nextOne, oldDefinitions, nextDefinitions); err != nil {
			return err
		}
		// oneOf is exclusive, not an ordinary set union. An appended arm
		// must not add a second match for any previously admitted response.
		for index, extra := range nextOne[len(oldOne):] {
			extraMap, ok := extra.(map[string]any)
			if !ok {
				return ErrInvalidSchema
			}
			for _, oldArm := range oldOne {
				oldMap, ok := oldArm.(map[string]any)
				if !ok || !disjointCompatibilitySchemas(oldMap, extraMap, oldDefinitions, nextDefinitions, 0) {
					return fmt.Errorf("oneOf[%d] overlaps or cannot be proved disjoint from an old response arm: %w", len(oldOne)+index, ErrInvalidSchema)
				}
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

// Existing string arms can grow only by explicit enum members: format/type
// changes and enum removal are already rejected above. Check both old and new
// literals so a newly added member cannot steal an old unbounded string value.
func preserveCompatibilityStringUnion(oldArms, nextArms []any, oldDefinitions, nextDefinitions map[string]json.RawMessage) error {
	values := map[string]bool{}
	var collect func(map[string]any, map[string]json.RawMessage, int)
	collect = func(schema map[string]any, definitions map[string]json.RawMessage, depth int) {
		if depth > maximumRuntimeSchemaDepth {
			return
		}
		resolved, ok := resolveCompatibilitySchema(schema, definitions)
		if !ok {
			return
		}
		if alternatives, ok := resolved["oneOf"].([]any); ok {
			for _, alternative := range alternatives {
				if arm, ok := alternative.(map[string]any); ok {
					collect(arm, definitions, depth+1)
				}
			}
		} else if stringValue(resolved["type"]) == "string" {
			if members, ok := resolved["enum"].([]any); ok {
				for _, member := range members {
					if text, ok := member.(string); ok {
						values[text] = true
					}
				}
			}
		}
	}
	for side, arms := range [][]any{oldArms, nextArms} {
		definitions := oldDefinitions
		if side == 1 {
			definitions = nextDefinitions
		}
		for _, arm := range arms {
			if schema, ok := arm.(map[string]any); ok {
				collect(schema, definitions, 0)
			}
		}
	}
	ordered := make([]string, 0, len(values))
	for value := range values {
		ordered = append(ordered, value)
	}
	sort.Strings(ordered)
	for _, value := range ordered {
		if compatibilityStringUnionMatches(oldArms, value, oldDefinitions, 0) == 1 &&
			compatibilityStringUnionMatches(nextArms, value, nextDefinitions, 0) != 1 {
			return fmt.Errorf("%w: oneOf enum growth invalidates old string %q", ErrInvalidSchema, value)
		}
	}
	return nil
}

func compatibilityStringUnionMatches(arms []any, value string, definitions map[string]json.RawMessage, depth int) int {
	if depth > maximumRuntimeSchemaDepth {
		return 0
	}
	matches := 0
	for _, arm := range arms {
		schema, ok := arm.(map[string]any)
		if !ok {
			continue
		}
		schema, ok = resolveCompatibilitySchema(schema, definitions)
		if !ok {
			continue
		}
		if nested, ok := schema["oneOf"].([]any); ok {
			if compatibilityStringUnionMatches(nested, value, definitions, depth+1) == 1 {
				matches++
			}
			continue
		}
		if stringValue(schema["type"]) != "string" || !matchesFormat(stringValue(schema["format"]), value) {
			continue
		}
		if members, restricted := schema["enum"].([]any); restricted {
			for _, member := range members {
				if text, ok := member.(string); ok && text == value {
					matches++
					break
				}
			}
		} else {
			matches++
		}
	}
	return matches
}

func resolveCompatibilitySchema(schema map[string]any, definitions map[string]json.RawMessage) (map[string]any, bool) {
	seen := map[string]bool{}
	for {
		ref, ok := schema["$ref"].(string)
		if !ok {
			return schema, schema != nil
		}
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(ref, prefix) || seen[ref] {
			return nil, false
		}
		seen[ref] = true
		raw, ok := definitions[strings.TrimPrefix(ref, prefix)]
		if !ok {
			return nil, false
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var resolved map[string]any
		if decoder.Decode(&resolved) != nil || resolved == nil {
			return nil, false
		}
		schema = resolved
	}
}

// A sufficient, bounded disjointness proof over the existing C17 grammar.
// Arrays can both admit [], and unknown string-format intersections are not
// assumed disjoint. References use each side's own definitions and exact bounds.
func disjointCompatibilitySchemas(left, right map[string]any, leftDefinitions, rightDefinitions map[string]json.RawMessage, depth int) bool {
	if depth > maximumRuntimeSchemaDepth {
		return false
	}
	left, leftOK := resolveCompatibilitySchema(left, leftDefinitions)
	right, rightOK := resolveCompatibilitySchema(right, rightDefinitions)
	if !leftOK || !rightOK {
		return false
	}
	for side, schema := range []map[string]any{left, right} {
		if alternatives, ok := schema["oneOf"].([]any); ok {
			if len(alternatives) < 2 {
				return false
			}
			for _, alternative := range alternatives {
				arm, ok := alternative.(map[string]any)
				if !ok || side == 0 && !disjointCompatibilitySchemas(arm, right, leftDefinitions, rightDefinitions, depth+1) ||
					side == 1 && !disjointCompatibilitySchemas(left, arm, leftDefinitions, rightDefinitions, depth+1) {
					return false
				}
			}
			return true
		}
	}
	leftType, rightType := stringValue(left["type"]), stringValue(right["type"])
	if leftType == "" || rightType == "" {
		return false
	}
	if leftType != rightType {
		return true
	}
	if leftType == "string" {
		leftEnum, leftOK := left["enum"].([]any)
		rightEnum, rightOK := right["enum"].([]any)
		if !leftOK || !rightOK || len(leftEnum) == 0 || len(rightEnum) == 0 {
			return false
		}
		for _, before := range leftEnum {
			for _, after := range rightEnum {
				if before == after {
					return false
				}
			}
		}
		return true
	}
	if leftType == "integer" {
		for _, pair := range [][2]map[string]any{{left, right}, {right, left}} {
			maximum, maxOK := pair[0]["maximum"].(json.Number)
			minimum, minOK := pair[1]["minimum"].(json.Number)
			if maxOK && minOK {
				maxValue, maxErr := maximum.Int64()
				minValue, minErr := minimum.Int64()
				if maxErr == nil && minErr == nil && maxValue < minValue {
					return true
				}
			}
		}
	}
	if leftType != "object" {
		return false
	}
	if disjointClosedCompatibilityObjects(left, right) {
		return true
	}
	leftProperties, leftOK := left["properties"].(map[string]any)
	rightProperties, rightOK := right["properties"].(map[string]any)
	if !leftOK || !rightOK {
		return false
	}
	required := stringSet(left["required"])
	for name := range stringSet(right["required"]) {
		required[name] = true
	}
	for name := range required {
		leftField, leftOK := leftProperties[name].(map[string]any)
		rightField, rightOK := rightProperties[name].(map[string]any)
		if leftOK && rightOK && disjointCompatibilitySchemas(leftField, rightField, leftDefinitions, rightDefinitions, depth+1) {
			return true
		}
	}
	return false
}

func disjointClosedCompatibilityObjects(left, right map[string]any) bool {
	if left["type"] != "object" || right["type"] != "object" ||
		left["additionalProperties"] != false || right["additionalProperties"] != false {
		return false
	}
	leftProperties, leftOK := left["properties"].(map[string]any)
	rightProperties, rightOK := right["properties"].(map[string]any)
	if !leftOK || !rightOK {
		return false
	}
	for name := range stringSet(left["required"]) {
		if _, allowed := rightProperties[name]; !allowed {
			return true
		}
	}
	for name := range stringSet(right["required"]) {
		if _, allowed := leftProperties[name]; !allowed {
			return true
		}
	}
	return false
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
