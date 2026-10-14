package economy

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Inspect original curve members before pointer decoding erases a wrong-arm
// null field. The enclosing catalog/price grammar remains owned by LoadCatalog.
func validateCostCurveShapes(data []byte) error {
	var source struct {
		Generators []struct {
			Price struct {
				Curve json.RawMessage `json:"curve"`
			} `json:"price"`
		} `json:"generator_classes"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		return catalogError("decode", err)
	}
	for index, generator := range source.Generators {
		var fields map[string]json.RawMessage
		var kind CurveKind
		if json.Unmarshal(generator.Price.Curve, &fields) != nil || json.Unmarshal(fields["kind"], &kind) != nil {
			return catalogError(fmt.Sprintf("generator_classes[%d].price.curve", index), errors.New("curve requires a literal kind"))
		}
		expected := []string{"kind"}
		switch kind {
		case CurveConstant:
		case CurveLinear:
			expected = append(expected, "step")
		case CurveGeometric:
			expected = append(expected, "ratio")
		default:
			return catalogError("curve.kind", fmt.Errorf("unsupported curve kind %q", kind))
		}
		if len(fields) != len(expected) {
			return catalogError("curve", errors.New("curve has wrong-arm or extra fields"))
		}
		for _, key := range expected {
			if _, present := fields[key]; !present {
				return catalogError("curve", errors.New("curve requires its exact arm keys"))
			}
		}
	}
	return nil
}
