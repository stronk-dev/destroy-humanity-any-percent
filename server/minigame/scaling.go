package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"cloud-clicker/server/decimal"
)

var ErrInvalidScalingPolicy = errors.New("invalid minigame scaling policy")

type ScalingSourceKind string

const (
	ScalingLiteral                 ScalingSourceKind = "literal"
	ScalingTier                    ScalingSourceKind = "tier"
	ScalingPurchasedGeneratorCount ScalingSourceKind = "purchased_generator_count"
	ScalingFounderCarryCounter     ScalingSourceKind = "founder_carry_counter"
	ScalingAttendedQualityGrade    ScalingSourceKind = "attended_quality_grade"
)

type ScalingOp string

const (
	ScalingIdentity ScalingOp = "identity"
	ScalingAdd      ScalingOp = "add"
	ScalingMul      ScalingOp = "mul"
	ScalingFloorDiv ScalingOp = "floordiv"
)

type ScalingRow struct {
	Destination      string            `json:"destination"`
	DestinationClass DestinationClass  `json:"destination_class"`
	SourceKind       ScalingSourceKind `json:"source_kind"`
	SourceRef        string            `json:"source_ref"`
	Op               ScalingOp         `json:"op"`
	Operand          int64             `json:"operand"`
	ClampMin         int64             `json:"clamp_min"`
	ClampMax         int64             `json:"clamp_max"`
}

type ScalingPolicy struct {
	rows []ScalingRow
}

type ScalingContext struct {
	Tier                     int64
	PurchasedGeneratorCounts map[string]int64
	FounderCarryCounters     map[string]int64
	AttendedQualityGrades    map[string]int64
}

type scalingPolicyWire struct {
	SchemaVersion int          `json:"schema_version"`
	ScalingInputs []ScalingRow `json:"scaling_inputs"`
}

var canonicalIntegerPattern = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)$`)

var founderCarryScalingPaths = map[string]bool{
	"achievement_score_lifetime": true,
	"age_ms":                     true,
	"exit_history_count":         true,
	"notoriety":                  true,
	"reputation_level":           true,
	"route_knowledge_balance":    true,
}

// LoadScalingPolicy closes C28's one-row grammar. Ranked status belongs to
// the owning minigame declaration; passing it here makes the Fairness Law
// enforceable without storing that policy twice in every row.
func LoadScalingPolicy(data []byte, ranked bool) (*ScalingPolicy, error) {
	if !uniqueJSONKeys(data) {
		return nil, ErrInvalidScalingPolicy
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire scalingPolicyWire
	if err := decoder.Decode(&wire); err != nil {
		return nil, ErrInvalidScalingPolicy
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || wire.SchemaVersion != 1 || len(wire.ScalingInputs) == 0 {
		return nil, ErrInvalidScalingPolicy
	}
	policy := &ScalingPolicy{rows: make([]ScalingRow, len(wire.ScalingInputs))}
	seen := make(map[string]bool, len(wire.ScalingInputs))
	for index, row := range wire.ScalingInputs {
		if !validScalingRow(row, ranked) || seen[row.Destination] {
			return nil, ErrInvalidScalingPolicy
		}
		seen[row.Destination] = true
		policy.rows[index] = row
	}
	return policy, nil
}

func validScalingRow(row ScalingRow, ranked bool) bool {
	if !mechanicalPattern.MatchString(row.Destination) ||
		(row.DestinationClass != DestinationPower && row.DestinationClass != DestinationBreadth && row.DestinationClass != DestinationPresentation) ||
		ranked && row.DestinationClass == DestinationPower || row.ClampMin < -decimal.MaxExactInteger ||
		row.ClampMax > decimal.MaxExactInteger || row.ClampMin > row.ClampMax || row.Operand < -decimal.MaxExactInteger || row.Operand > decimal.MaxExactInteger {
		return false
	}
	switch row.SourceKind {
	case ScalingLiteral:
		if !canonicalIntegerPattern.MatchString(row.SourceRef) {
			return false
		}
		value, err := strconv.ParseInt(row.SourceRef, 10, 64)
		if err != nil || value < -decimal.MaxExactInteger || value > decimal.MaxExactInteger || strconv.FormatInt(value, 10) != row.SourceRef {
			return false
		}
	case ScalingTier:
		if row.SourceRef != "tier" {
			return false
		}
	case ScalingPurchasedGeneratorCount, ScalingAttendedQualityGrade:
		if !mechanicalPattern.MatchString(row.SourceRef) {
			return false
		}
	case ScalingFounderCarryCounter:
		if !founderCarryScalingPaths[row.SourceRef] {
			return false
		}
	default:
		return false
	}
	switch row.Op {
	case ScalingIdentity:
		return row.Operand == 0
	case ScalingAdd, ScalingMul:
		return true
	case ScalingFloorDiv:
		return row.Operand > 0
	default:
		return false
	}
}

func (policy *ScalingPolicy) Rows() []ScalingRow {
	if policy == nil {
		return nil
	}
	return append([]ScalingRow(nil), policy.rows...)
}

func (policy *ScalingPolicy) Resolve(context ScalingContext) (map[string]int64, error) {
	if policy == nil || !validScalingContext(context) {
		return nil, ErrInvalidScalingPolicy
	}
	resolved := make(map[string]int64, len(policy.rows))
	for _, row := range policy.rows {
		source, ok := scalingSourceValue(row, context)
		if !ok {
			return nil, ErrInvalidScalingPolicy
		}
		value := big.NewInt(source)
		operand := big.NewInt(row.Operand)
		switch row.Op {
		case ScalingAdd:
			value.Add(value, operand)
		case ScalingMul:
			value.Mul(value, operand)
		case ScalingFloorDiv:
			value = floorBigInt(value, operand)
		}
		minimum, maximum := big.NewInt(row.ClampMin), big.NewInt(row.ClampMax)
		if value.Cmp(minimum) < 0 {
			value.Set(minimum)
		} else if value.Cmp(maximum) > 0 {
			value.Set(maximum)
		}
		if !value.IsInt64() {
			return nil, ErrInvalidScalingPolicy
		}
		resolved[row.Destination] = value.Int64()
	}
	return resolved, nil
}

func validScalingContext(context ScalingContext) bool {
	if context.Tier < 0 || context.Tier > decimal.MaxExactInteger || context.PurchasedGeneratorCounts == nil ||
		context.FounderCarryCounters == nil || context.AttendedQualityGrades == nil {
		return false
	}
	for _, values := range []map[string]int64{context.PurchasedGeneratorCounts, context.FounderCarryCounters, context.AttendedQualityGrades} {
		for key, value := range values {
			if !mechanicalPattern.MatchString(key) || value < -decimal.MaxExactInteger || value > decimal.MaxExactInteger {
				return false
			}
		}
	}
	return true
}

func scalingSourceValue(row ScalingRow, context ScalingContext) (int64, bool) {
	switch row.SourceKind {
	case ScalingLiteral:
		value, err := strconv.ParseInt(row.SourceRef, 10, 64)
		return value, err == nil
	case ScalingTier:
		return context.Tier, true
	case ScalingPurchasedGeneratorCount:
		value, ok := context.PurchasedGeneratorCounts[row.SourceRef]
		return value, ok
	case ScalingFounderCarryCounter:
		value, ok := context.FounderCarryCounters[row.SourceRef]
		return value, ok
	case ScalingAttendedQualityGrade:
		value, ok := context.AttendedQualityGrades[row.SourceRef]
		return value, ok
	default:
		return 0, false
	}
}

func floorBigInt(numerator, denominator *big.Int) *big.Int {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if numerator.Sign() < 0 && remainder.Sign() != 0 {
		quotient.Sub(quotient, big.NewInt(1))
	}
	return quotient
}

func (row ScalingRow) Formula() string {
	source := string(row.SourceKind) + ":" + row.SourceRef
	operation := string(row.Op)
	if row.Op != ScalingIdentity {
		operation += "(" + strconv.FormatInt(row.Operand, 10) + ")"
	}
	return strings.Join([]string{row.Destination, source, operation,
		"clamp[" + strconv.FormatInt(row.ClampMin, 10) + "," + strconv.FormatInt(row.ClampMax, 10) + "]"}, " <- ")
}
