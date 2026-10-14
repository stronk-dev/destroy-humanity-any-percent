package publicapi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Growth must not steal a previously exclusive object from another old arm.
// Prove that every new-arm/old-other intersection was already in another old arm:
// such values previously matched at least twice and were never valid replies.
// Unknown intersections fail closed, as for appended union alternatives.
func preserveCompatibilityObjectUnion(oldArms, nextArms []any, oldDefinitions, nextDefinitions map[string]json.RawMessage) error {
	for index, arm := range oldArms {
		old, ok := resolveCompatibilitySchema(arm.(map[string]any), oldDefinitions)
		if !ok {
			return ErrInvalidSchema
		}
		if !compatibilityContainsObject(old, oldDefinitions, 0) {
			continue
		}
		next := nextArms[index].(map[string]any)
		for otherIndex, other := range oldArms {
			if index == otherIndex {
				continue
			}
			covered := false
			for priorIndex, prior := range oldArms {
				if priorIndex != otherIndex && compatibilityIntersectionCovered(prior.(map[string]any), next, other.(map[string]any), oldDefinitions, nextDefinitions, oldDefinitions, 0) {
					covered = true
					break
				}
			}
			if !covered {
				return fmt.Errorf("%w: oneOf object arm %d growth overlaps old arm %d", ErrInvalidSchema, index, otherIndex)
			}
		}
	}
	return nil
}

func compatibilityContainsObject(schema map[string]any, definitions map[string]json.RawMessage, depth int) bool {
	if depth > maximumRuntimeSchemaDepth {
		return false
	}
	resolved, ok := resolveCompatibilitySchema(schema, definitions)
	if !ok {
		return false
	}
	if stringValue(resolved["type"]) == "object" {
		return true
	}
	if arms, ok := resolved["oneOf"].([]any); ok {
		for _, arm := range arms {
			if branch, ok := arm.(map[string]any); ok && compatibilityContainsObject(branch, definitions, depth+1) {
				return true
			}
		}
	}
	return false
}

func compatibilitySchemasEqual(left, right map[string]any, leftDefinitions, rightDefinitions map[string]json.RawMessage, depth int) bool {
	if depth > maximumRuntimeSchemaDepth {
		return false
	}
	left, leftOK := resolveCompatibilitySchema(left, leftDefinitions)
	right, rightOK := resolveCompatibilitySchema(right, rightDefinitions)
	if !leftOK || !rightOK || len(left) != len(right) {
		return false
	}
	for key, value := range left {
		other, ok := right[key]
		if !ok {
			return false
		}
		switch key {
		case "properties":
			fields, leftOK := value.(map[string]any)
			others, rightOK := other.(map[string]any)
			if !leftOK || !rightOK || len(fields) != len(others) {
				return false
			}
			for name, field := range fields {
				leftField, leftOK := field.(map[string]any)
				rightField, rightOK := others[name].(map[string]any)
				if !leftOK || !rightOK || !compatibilitySchemasEqual(leftField, rightField, leftDefinitions, rightDefinitions, depth+1) {
					return false
				}
			}
		case "items":
			leftItems, leftOK := value.(map[string]any)
			rightItems, rightOK := other.(map[string]any)
			if !leftOK || !rightOK || !compatibilitySchemasEqual(leftItems, rightItems, leftDefinitions, rightDefinitions, depth+1) {
				return false
			}
		case "oneOf":
			arms, leftOK := value.([]any)
			others, rightOK := other.([]any)
			if !leftOK || !rightOK || len(arms) != len(others) {
				return false
			}
			for index, arm := range arms {
				leftArm, leftOK := arm.(map[string]any)
				rightArm, rightOK := others[index].(map[string]any)
				if !leftOK || !rightOK || !compatibilitySchemasEqual(leftArm, rightArm, leftDefinitions, rightDefinitions, depth+1) {
					return false
				}
			}
		default:
			if !reflect.DeepEqual(value, other) {
				return false
			}
		}
	}
	return true
}

// A sufficient proof of (left ∩ right) ⊆ old over the closed C17 grammar.
// References always resolve with their own side's definitions; integer bounds
// stay exact. No value-space enumeration, floating bounds or new depth budget.
func compatibilityIntersectionCovered(old, left, right map[string]any,
	oldDefinitions, leftDefinitions, rightDefinitions map[string]json.RawMessage, depth int,
) bool {
	if depth > maximumRuntimeSchemaDepth {
		return false
	}
	old, oldOK := resolveCompatibilitySchema(old, oldDefinitions)
	left, leftOK := resolveCompatibilitySchema(left, leftDefinitions)
	right, rightOK := resolveCompatibilitySchema(right, rightDefinitions)
	if !oldOK || !leftOK || !rightOK {
		return false
	}
	if compatibilitySchemasEqual(old, left, oldDefinitions, leftDefinitions, depth) ||
		compatibilitySchemasEqual(old, right, oldDefinitions, rightDefinitions, depth) ||
		disjointCompatibilitySchemas(left, right, leftDefinitions, rightDefinitions, depth) {
		return true
	}
	for side, schema := range []map[string]any{left, right} {
		if arms, ok := schema["oneOf"].([]any); ok {
			for _, arm := range arms {
				branch, ok := arm.(map[string]any)
				if !ok {
					return false
				}
				if side == 0 && !compatibilityIntersectionCovered(old, branch, right, oldDefinitions, leftDefinitions, rightDefinitions, depth+1) ||
					side == 1 && !compatibilityIntersectionCovered(old, left, branch, oldDefinitions, leftDefinitions, rightDefinitions, depth+1) {
					return false
				}
			}
			return len(arms) >= 2
		}
	}
	if oldArms, ok := old["oneOf"].([]any); ok {
		for index, arm := range oldArms {
			branch, ok := arm.(map[string]any)
			if !ok || !compatibilityIntersectionCovered(branch, left, right, oldDefinitions, leftDefinitions, rightDefinitions, depth+1) {
				continue
			}
			exclusive := true
			for otherIndex, other := range oldArms {
				if index == otherIndex {
					continue
				}
				otherBranch, ok := other.(map[string]any)
				if !ok || !disjointCompatibilitySchemas(left, otherBranch, leftDefinitions, oldDefinitions, depth+1) &&
					!disjointCompatibilitySchemas(right, otherBranch, rightDefinitions, oldDefinitions, depth+1) {
					exclusive = false
					break
				}
			}
			if exclusive {
				return true
			}
		}
		return false
	}
	kind := stringValue(old["type"])
	if kind == "" || kind != stringValue(left["type"]) || kind != stringValue(right["type"]) {
		return false
	}
	switch kind {
	case "boolean", "null":
		return true
	case "string":
		for _, schema := range []map[string]any{left, right} {
			if members, finite := schema["enum"].([]any); finite {
				for _, member := range members {
					value, ok := member.(string)
					if !ok {
						return false
					}
					if compatibilityStringUnionMatches([]any{left}, value, leftDefinitions, depth) == 1 &&
						compatibilityStringUnionMatches([]any{right}, value, rightDefinitions, depth) == 1 &&
						compatibilityStringUnionMatches([]any{old}, value, oldDefinitions, depth) != 1 {
						return false
					}
				}
				return true
			}
		}
	case "integer":
		oldMin, oldMax, oldOK := compatibilityIntegerRange(old)
		leftMin, leftMax, leftOK := compatibilityIntegerRange(left)
		rightMin, rightMax, rightOK := compatibilityIntegerRange(right)
		if oldOK && leftOK && rightOK {
			minimum, maximum := max(leftMin, rightMin), min(leftMax, rightMax)
			return minimum > maximum || minimum >= oldMin && maximum <= oldMax
		}
	case "array":
		oldItems, oldOK := old["items"].(map[string]any)
		leftItems, leftOK := left["items"].(map[string]any)
		rightItems, rightOK := right["items"].(map[string]any)
		return oldOK && leftOK && rightOK && compatibilityIntersectionCovered(oldItems, leftItems, rightItems, oldDefinitions, leftDefinitions, rightDefinitions, depth+1)
	case "object":
		if old["additionalProperties"] != false || left["additionalProperties"] != false || right["additionalProperties"] != false {
			return false
		}
		oldFields, oldOK := old["properties"].(map[string]any)
		leftFields, leftOK := left["properties"].(map[string]any)
		rightFields, rightOK := right["properties"].(map[string]any)
		if !oldOK || !leftOK || !rightOK {
			return false
		}
		required := stringSet(left["required"])
		for name := range stringSet(right["required"]) {
			required[name] = true
		}
		for name := range stringSet(old["required"]) {
			if !required[name] {
				return false
			}
		}
		for name, field := range leftFields {
			leftField, leftOK := field.(map[string]any)
			rightField, rightOK := rightFields[name].(map[string]any)
			if !rightOK {
				continue
			} // Closed intersection cannot contain this field.
			if !leftOK {
				return false
			}
			if disjointCompatibilitySchemas(leftField, rightField, leftDefinitions, rightDefinitions, depth+1) {
				if required[name] {
					return true
				} // Intersection is empty.
				continue // This optional field cannot be present.
			}
			oldField, oldOK := oldFields[name].(map[string]any)
			if !oldOK || !compatibilityIntersectionCovered(oldField, leftField, rightField, oldDefinitions, leftDefinitions, rightDefinitions, depth+1) {
				return false
			}
		}
		return true
	}
	return false
}
