package publicapi

import (
	"encoding/json"
	"fmt"
	"sort"
)

const compatibilityIntegerFloor = int64(-1 << 63)
const compatibilityIntegerCeiling = int64(1<<63 - 1)

// Integer membership is constant between adjacent minimum/max+1 cut points.
// Testing every region's first integer is therefore exhaustive, not sampling.
func preserveCompatibilityIntegerUnion(oldArms, nextArms []any, oldDefinitions, nextDefinitions map[string]json.RawMessage) error {
	points := map[int64]bool{compatibilityIntegerFloor: true}
	var collect func(map[string]any, map[string]json.RawMessage, int)
	collect = func(schema map[string]any, definitions map[string]json.RawMessage, depth int) {
		if depth > maximumRuntimeSchemaDepth {
			return
		}
		resolved, ok := resolveCompatibilitySchema(schema, definitions)
		if !ok {
			return
		}
		if nested, ok := resolved["oneOf"].([]any); ok {
			for _, arm := range nested {
				if schema, ok := arm.(map[string]any); ok {
					collect(schema, definitions, depth+1)
				}
			}
			return
		}
		minimum, maximum, ok := compatibilityIntegerRange(resolved)
		if !ok {
			return
		}
		points[minimum] = true
		if maximum < compatibilityIntegerCeiling {
			points[maximum+1] = true
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
	ordered := make([]int64, 0, len(points))
	for value := range points {
		ordered = append(ordered, value)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for _, value := range ordered {
		if compatibilityIntegerUnionMatches(oldArms, value, oldDefinitions, 0) == 1 &&
			compatibilityIntegerUnionMatches(nextArms, value, nextDefinitions, 0) != 1 {
			return fmt.Errorf("%w: oneOf range growth invalidates old integer %d", ErrInvalidSchema, value)
		}
	}
	return nil
}

func compatibilityIntegerRange(schema map[string]any) (int64, int64, bool) {
	if stringValue(schema["type"]) != "integer" {
		return 0, 0, false
	}
	minimum, maximum := compatibilityIntegerFloor, compatibilityIntegerCeiling
	for _, bound := range []struct {
		name   string
		target *int64
	}{{"minimum", &minimum}, {"maximum", &maximum}} {
		if raw, present := schema[bound.name]; present {
			number, ok := raw.(json.Number)
			if !ok {
				return 0, 0, false
			}
			value, err := number.Int64()
			if err != nil {
				return 0, 0, false
			}
			*bound.target = value
		}
	}
	return minimum, maximum, minimum <= maximum
}

func compatibilityIntegerUnionMatches(arms []any, value int64, definitions map[string]json.RawMessage, depth int) int {
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
			if compatibilityIntegerUnionMatches(nested, value, definitions, depth+1) == 1 {
				matches++
			}
			continue
		}
		minimum, maximum, ok := compatibilityIntegerRange(schema)
		if ok && value >= minimum && value <= maximum {
			matches++
		}
	}
	return matches
}
