package economy

import (
	"encoding/json"
	"fmt"
)

// Validate original member presence before encoding/json folds case or turns
// inactive/version-forbidden nulls and zeros into absent fields. Semantics,
// reference topology, Decimal domains and numeric values stay in the loader.
func validateRemainingCatalogShapes(data []byte) error {
	object := func(data json.RawMessage, keys ...string) (map[string]json.RawMessage, error) {
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil || !exactGeneratorPolicyKeys(fields, keys...) {
			return nil, ErrInvalidCatalog
		}
		return fields, nil
	}
	array := func(data json.RawMessage) ([]json.RawMessage, error) {
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil || values == nil {
			return nil, ErrInvalidCatalog
		}
		return values, nil
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(data, &root) != nil {
		return catalogError("root", ErrInvalidCatalog)
	}
	var version int
	if json.Unmarshal(root["schema_version"], &version) != nil || version < 1 || version > CatalogSchemaVersion {
		return catalogError("schema_version", ErrInvalidCatalog)
	}
	keys := []string{"schema_version", "resources", "generator_classes"}
	if version >= 3 {
		keys = append(keys, "manual_actions", "multiplier_sources", "progress_coordinates", "manual_policy", "offline_policy")
	}
	if version >= 4 {
		keys = append(keys, "upgrades", "synergy_pools", "provision_tick_ms")
	}
	if version == 5 {
		if _, present := root["axis_stack"]; present {
			keys = append(keys, "axis_stack")
		}
	}
	if !exactGeneratorPolicyKeys(root, keys...) {
		return catalogError("root", ErrInvalidCatalog)
	}
	for _, family := range []string{"resources", "generator_classes"} {
		if _, err := array(root[family]); err != nil {
			return catalogError(family, err)
		}
	}
	if version < 3 {
		return nil
	}
	for _, family := range []string{"manual_actions", "multiplier_sources", "progress_coordinates", "upgrades", "synergy_pools"} {
		if version < 4 && (family == "upgrades" || family == "synergy_pools") {
			continue
		}
		rows, err := array(root[family])
		if err != nil {
			return catalogError(family, err)
		}
		for index, row := range rows {
			path := fmt.Sprintf("%s[%d]", family, index)
			switch family {
			case "manual_actions":
				fields, err := object(row, "id", "output")
				if err != nil {
					return catalogError(path, err)
				}
				if _, err := object(fields["output"], "resource_id", "amount_per_action"); err != nil {
					return catalogError(path+".output", err)
				}
			case "multiplier_sources":
				if _, err := object(row, "id", "slot", "target", "provider"); err != nil {
					return catalogError(path, err)
				}
			case "progress_coordinates":
				if err := validateProgressShape(row, false); err != nil {
					return catalogError(path, err)
				}
			case "upgrades":
				fields, err := object(row, "id", "cost", "window", "requires", "effects", "roles", "copy_key")
				if err != nil {
					return catalogError(path, err)
				}
				if _, err := object(fields["cost"], "resource", "amount"); err != nil {
					return catalogError(path+".cost", err)
				}
				var window map[string]json.RawMessage
				if json.Unmarshal(fields["window"], &window) != nil || len(window) != 2 || window["from_gate"] == nil || window["to_gate"] == nil {
					return catalogError(path+".window", ErrInvalidCatalog)
				} // explicit null gate bounds are valid
				effects, err := array(fields["effects"])
				if err != nil {
					return catalogError(path+".effects", err)
				}
				for _, effect := range effects {
					var header struct {
						Slot string `json:"slot"`
					}
					if json.Unmarshal(effect, &header) != nil {
						return catalogError(path+".effects", ErrInvalidCatalog)
					}
					operand := "factor"
					if header.Slot == string(SlotAxisStack) {
						operand = "factor_ppm"
					}
					if _, err := object(effect, "source_id", "slot", "target", operand); err != nil {
						return catalogError(path+".effects", err)
					}
				}
				requires, err := array(fields["requires"])
				if err != nil {
					return catalogError(path+".requires", err)
				}
				for _, condition := range requires {
					var header struct {
						Kind string `json:"kind"`
					}
					if json.Unmarshal(condition, &header) != nil {
						return catalogError(path+".requires", ErrInvalidCatalog)
					}
					if header.Kind == "axis_at_least" {
						if _, err := object(condition, "kind", "minimum"); err != nil {
							return catalogError(path+".requires", err)
						}
					}
				} // the shared Route owner validates all non-axis arms
			case "synergy_pools":
				fields, err := object(row, "id", "sources", "slot", "curve")
				if err != nil {
					return catalogError(path, err)
				}
				sources, err := array(fields["sources"])
				if err != nil {
					return catalogError(path+".sources", err)
				}
				for _, source := range sources {
					if _, err := object(source, "kind", "id_or_class", "per_count_ppm"); err != nil {
						return catalogError(path+".sources", err)
					}
				}
			}
		}
	}
	if axis, present := root["axis_stack"]; present {
		if _, err := object(axis, "input", "input_cap", "cap_reason_key"); err != nil {
			return catalogError("axis_stack", err)
		}
	}
	return nil
}

func validateProgressShape(data json.RawMessage, weighted bool) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return ErrInvalidCatalog
	}
	var kind string
	if json.Unmarshal(fields["kind"], &kind) != nil {
		return ErrInvalidCatalog
	}
	keys := []string{"kind"}
	if weighted {
		keys = append(keys, "weight")
	} else {
		keys = append(keys, "tier")
	}
	switch ProgressKind(kind) {
	case ProgressResourceLog:
		keys = append(keys, "resource", "target")
	case ProgressCountFraction:
		keys = append(keys, "count", "required")
	case ProgressComposite:
		if weighted {
			return ErrInvalidCatalog
		}
		keys = append(keys, "terms")
	default:
		return ErrInvalidCatalog
	}
	if !exactGeneratorPolicyKeys(fields, keys...) {
		return ErrInvalidCatalog
	}
	if kind == string(ProgressComposite) {
		var terms []json.RawMessage
		if json.Unmarshal(fields["terms"], &terms) != nil || terms == nil {
			return ErrInvalidCatalog
		}
		for _, term := range terms {
			if err := validateProgressShape(term, true); err != nil {
				return err
			}
		}
	}
	return nil
}
