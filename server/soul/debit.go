package soul

import "errors"

var (
	ErrUnknownSource      = errors.New("unknown soul source")
	ErrOwnerMismatch      = errors.New("soul owner mismatch")
	ErrSourceConsumed     = errors.New("soul source consumed")
	ErrMissingEligibility = errors.New("soul eligibility missing")
	ErrUnaffordable       = errors.New("soul debit unaffordable")
	ErrInvalidDebitState  = errors.New("invalid soul debit state")
)

type DebitState struct {
	Soul               int64
	ExhaustedSourceIDs []string
	DepletedFact       bool
}

type DebitCommand struct {
	SourceID       string
	OwnerKind      OwnerKind
	EligibilityRef string
}

type DebitResult struct {
	SoulBefore        int64      `json:"soul_before"`
	Debit             int64      `json:"debit"`
	SoulAfter         int64      `json:"soul_after"`
	BandBefore        BandMember `json:"band_before"`
	BandAfter         BandMember `json:"band_after"`
	CurtainCopyKey    string     `json:"curtain_copy_key"`
	DepletedFirstTime bool       `json:"depleted_first_time"`
}

func ApplyDebit(state *DebitState, catalog *Catalog, command DebitCommand) (DebitResult, error) {
	if state == nil || catalog == nil {
		return DebitResult{}, ErrInvalidDebitState
	}
	source, ok := catalog.DebitSource(command.SourceID)
	if !ok {
		return DebitResult{}, ErrUnknownSource
	}
	if command.OwnerKind != source.OwnerKind {
		return DebitResult{}, ErrOwnerMismatch
	}
	if command.EligibilityRef == "" {
		return DebitResult{}, ErrMissingEligibility
	}
	beforeBand, ok := catalog.BandFor(state.Soul)
	if !ok || !sortedUnique(state.ExhaustedSourceIDs) {
		return DebitResult{}, ErrInvalidDebitState
	}
	if contains(state.ExhaustedSourceIDs, source.SourceID) {
		return DebitResult{}, ErrSourceConsumed
	}
	available := state.Soul - catalog.Policy.Floor
	debit := source.Amount
	if debit > available {
		if !source.MayExhaust || available <= 0 {
			return DebitResult{}, ErrUnaffordable
		}
		debit = available
	}
	if source.MayExhaust {
		state.ExhaustedSourceIDs = insertSorted(state.ExhaustedSourceIDs, source.SourceID)
	}
	state.Soul -= debit
	afterBand, ok := catalog.BandFor(state.Soul)
	if !ok {
		return DebitResult{}, ErrInvalidDebitState
	}
	depleted := state.Soul == catalog.Policy.Floor && !state.DepletedFact
	if depleted {
		state.DepletedFact = true
	}
	return DebitResult{SoulBefore: state.Soul + debit, Debit: debit, SoulAfter: state.Soul,
		BandBefore: beforeBand.Member, BandAfter: afterBand.Member, CurtainCopyKey: source.CurtainCopyKey,
		DepletedFirstTime: depleted}, nil
}

func sortedUnique(values []string) bool {
	for index, value := range values {
		if !mechanicalID.MatchString(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return values != nil
}

func contains(values []string, value string) bool {
	index := sortSearch(values, value)
	return index < len(values) && values[index] == value
}

func insertSorted(values []string, value string) []string {
	index := sortSearch(values, value)
	result := make([]string, len(values)+1)
	copy(result, values[:index])
	result[index] = value
	copy(result[index+1:], values[index:])
	return result
}

func sortSearch(values []string, value string) int {
	low, high := 0, len(values)
	for low < high {
		middle := int(uint(low+high) >> 1)
		if values[middle] < value {
			low = middle + 1
		} else {
			high = middle
		}
	}
	return low
}
