package routes

import (
	"bytes"
	"encoding/json"
)

// UnmarshalJSON makes every catalog/ParseConditions caller use the owner's
// literal, non-null active arm before pointer decoding erases wrong-arm nulls.
func (condition *rawCondition) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return ErrInvalidCatalog
	}
	var kind ConditionKind
	if json.Unmarshal(fields["kind"], &kind) != nil {
		return ErrInvalidCatalog
	}
	keys := []string{"kind"}
	switch kind {
	case ConditionResourceAtLeast, ConditionResourceAtMost:
		keys = append(keys, "resource_id", "value")
	case ConditionMeterBand:
		keys = append(keys, "meter_id", "min", "max")
	case ConditionDoctrineIs, ConditionDoctrineIsNot:
		keys = append(keys, "transition", "doctrine_id")
	case ConditionLedgerFact:
		keys = append(keys, "fact_kind")
	case ConditionStructureIs:
		keys = append(keys, "structure_id")
	case ConditionRegionTrait:
		keys = append(keys, "trait_id")
	default:
		return ErrInvalidCatalog
	}
	if len(fields) != len(keys) {
		return ErrInvalidCatalog
	}
	for _, key := range keys {
		value, present := fields[key]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ErrInvalidCatalog
		}
	}
	type decodedCondition rawCondition
	var result decodedCondition
	if err := decodeStrict(data, &result); err != nil {
		return err
	}
	*condition = rawCondition(result)
	return nil
}
