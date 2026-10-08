// Package headcount implements the source-independent allocation arithmetic in
// Headcount Allocation S2. It does not obtain seats or apply gameplay commands.
package headcount

import (
	"errors"
	"regexp"
	"sort"
)

const maxSafeInteger int64 = 1<<53 - 1

var ErrInvalidAllocation = errors.New("invalid headcount allocation")
var mechanicalID = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// Role is only the resolved arithmetic input, not an economy catalog schema.
// Resource/rate/pool bindings remain the catalog loader's responsibility.
type Role struct {
	ID                string `json:"id"`
	Kind              string `json:"kind"`
	FromRoleID        string `json:"from_role_id,omitempty"`
	HeadsPerConverter int64  `json:"heads_per_converter,omitempty"`
}

type Allocation struct {
	ProducingHeads map[string]int64 `json:"producing_heads"`
	Diverted       map[string]int64 `json:"diverted"`
	IdleConverters map[string]int64 `json:"idle_converters"`
	Unassigned     int64            `json:"unassigned"`
}

// Project returns a fresh derived allocation without changing either input.
// Counts and the complete assignment must satisfy 0 <= sum <= seats <= hardcap
// in the shared safe-integer domain. No partial vector or implicit zero is used.
func Project(hardcap, seats int64, roles []Role, assigned map[string]int64) (Allocation, error) {
	invalid := func() (Allocation, error) { return Allocation{}, ErrInvalidAllocation }
	if hardcap <= 0 || hardcap > maxSafeInteger || seats < 0 || seats > hardcap || len(roles) == 0 || len(assigned) != len(roles) {
		return invalid()
	}
	byID := make(map[string]Role, len(roles))
	ordered := append([]Role(nil), roles...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	remaining := seats
	for _, role := range ordered {
		if !mechanicalID.MatchString(role.ID) {
			return invalid()
		}
		if _, duplicate := byID[role.ID]; duplicate {
			return invalid()
		}
		count, present := assigned[role.ID]
		// Subtract from the remaining budget: even a long invalid vector never
		// forms an overflowing sum before it can be rejected.
		if !present || count < 0 || count > remaining {
			return invalid()
		}
		remaining -= count
		switch role.Kind {
		case "produce", "pool_feed":
			if role.FromRoleID != "" || role.HeadsPerConverter != 0 {
				return invalid()
			}
		case "convert":
			if role.HeadsPerConverter <= 0 || role.HeadsPerConverter > maxSafeInteger {
				return invalid()
			}
		default:
			return invalid()
		}
		byID[role.ID] = role
	}
	result := Allocation{ProducingHeads: map[string]int64{}, Diverted: map[string]int64{}, IdleConverters: map[string]int64{}, Unassigned: remaining}
	for _, role := range ordered {
		if role.Kind != "convert" {
			result.ProducingHeads[role.ID] = assigned[role.ID]
		}
	}
	converted := map[string]bool{}
	for _, role := range ordered {
		if role.Kind != "convert" {
			continue
		}
		source, present := byID[role.FromRoleID]
		if !present || source.Kind != "produce" || converted[source.ID] {
			return invalid()
		}
		converted[source.ID] = true
		headCount, converters, k := assigned[source.ID], assigned[role.ID], role.HeadsPerConverter
		diverted := headCount
		if converters <= headCount/k {
			diverted = converters * k // Product is bounded by headCount, hence safe.
		}
		used := diverted / k
		if diverted%k != 0 {
			used++ // Ceil without forming diverted+k-1.
		}
		result.ProducingHeads[source.ID] = headCount - diverted
		result.Diverted[role.ID] = diverted
		result.IdleConverters[role.ID] = converters - used
	}
	return result, nil
}
