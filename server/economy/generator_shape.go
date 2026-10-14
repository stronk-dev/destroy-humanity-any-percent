package economy

import (
	"bytes"
	"encoding/json"
	"errors"
)

func validateGeneratorPolicyShapes(data []byte) error {
	var source struct {
		Version    int               `json:"schema_version"`
		Generators []json.RawMessage `json:"generator_classes"`
		Manual     json.RawMessage   `json:"manual_policy"`
		Offline    json.RawMessage   `json:"offline_policy"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		return catalogError("decode", err)
	}
	if source.Version < 1 || source.Version > CatalogSchemaVersion {
		return catalogError("schema_version", ErrInvalidCatalog)
	}
	check := func(value json.RawMessage, keys ...string) bool {
		var fields map[string]json.RawMessage
		return json.Unmarshal(value, &fields) == nil && exactGeneratorPolicyKeys(fields, keys...)
	}
	for _, data := range source.Generators {
		var row map[string]json.RawMessage
		if json.Unmarshal(data, &row) != nil {
			return catalogError("generator_classes", ErrInvalidCatalog)
		}
		keys := []string{"id", "price"}
		if source.Version >= 2 {
			keys = append(keys, "production")
		}
		if source.Version >= 4 {
			keys = append(keys, "tier", "category", "ladder", "roles")
			for _, optional := range []string{"provisions", "provisioned_hardcap"} {
				if _, present := row[optional]; present {
					keys = append(keys, optional)
				}
			}
		}
		if !exactGeneratorPolicyKeys(row, keys...) || !check(row["price"], "resource_id", "base", "curve") {
			return catalogError("generator_classes", errors.New("generator/price requires exact versioned keys"))
		}
		if source.Version >= 2 && !check(row["production"], "resource_id", "base_rate") {
			return catalogError("production", ErrInvalidCatalog)
		}
		if source.Version < 4 {
			continue
		}
		if value, present := row["provisions"]; present && !check(value, "generator_id", "rate_ppm") {
			return catalogError("provisions", ErrInvalidCatalog)
		}
		if value, present := row["provisioned_hardcap"]; present && !check(value, "count", "reason_key") {
			return catalogError("provisioned_hardcap", ErrInvalidCatalog)
		}
		var ladder, roles []json.RawMessage
		if json.Unmarshal(row["ladder"], &ladder) != nil || ladder == nil || json.Unmarshal(row["roles"], &roles) != nil || roles == nil {
			return catalogError("generator_classes", ErrInvalidCatalog)
		}
		for _, rung := range ladder {
			if !check(rung, "purchased_at", "multiplier_ppm") {
				return catalogError("ladder", ErrInvalidCatalog)
			}
		}
		for _, role := range roles {
			var header struct {
				Kind GeneratorRoleKind `json:"kind"`
			}
			if json.Unmarshal(role, &header) != nil {
				return catalogError("roles", ErrInvalidCatalog)
			}
			var keys []string
			switch header.Kind {
			case RoleProvision:
				keys = []string{"kind", "generator_id"}
			case RoleSynergyFeed:
				keys = []string{"kind", "pool_id"}
			case RoleManualOutput:
				keys = []string{"kind", "action_id", "per_purchased_ppm"}
			case RoleStockRate:
				keys = []string{"kind", "per_purchased_ppm"}
			default:
				return catalogError("roles", ErrInvalidCatalog)
			}
			if !check(role, keys...) {
				return catalogError("roles", errors.New("role requires exact active arm keys"))
			}
		}
	}
	if source.Version >= 3 && (!check(source.Manual, "refill_milli_per_ms", "bucket_cap_milli") ||
		!check(source.Offline, "efficiency", "accrual_cap_ms", "bank_ratio_numerator", "bank_ratio_denominator", "bank_cap_ms", "burst_speed", "burst_max_duration_ms")) {
		return catalogError("production_policy", ErrInvalidCatalog)
	}
	return nil
}

func exactGeneratorPolicyKeys(fields map[string]json.RawMessage, keys ...string) bool {
	if len(fields) != len(keys) {
		return false
	}
	for _, key := range keys {
		value, present := fields[key]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return false
		}
	}
	return true
}
