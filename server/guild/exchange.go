package guild

import (
	"errors"
	"sort"
)

var ErrInvalidExchange = errors.New("invalid guild exchange")

type MemberStock struct {
	AccountID      string
	Produces       string
	Consumes       string
	AvailableUnits int64
	ReceivedUnits  int64
}

type Allocation struct {
	AccountID string `json:"account_id"`
	Units     int64  `json:"units"`
}

type Clearing struct {
	ProducerAccountID string       `json:"producer_account_id"`
	Resource          string       `json:"resource"`
	Allocations       []Allocation `json:"allocations"`
	NPC               bool         `json:"npc"`
}

// Clear applies GC's ordered, one-pass allocation in memory. Callers persist
// all returned member states and events in one transaction.
func Clear(catalog *Catalog, members []MemberStock, stockCap int64) ([]MemberStock, []Clearing, error) {
	if catalog == nil || stockCap < 1 {
		return nil, nil, ErrInvalidExchange
	}
	result := append([]MemberStock(nil), members...)
	sort.Slice(result, func(left, right int) bool { return result[left].AccountID < result[right].AccountID })
	seen := map[string]bool{}
	for _, member := range result {
		neutral := member.Produces == "" && member.Consumes == "" && member.AvailableUnits == 0 && member.ReceivedUnits == 0
		if !uuidPattern.MatchString(member.AccountID) || !neutral && (member.Produces == "" || member.Consumes == "" || member.Produces == member.Consumes) ||
			member.AvailableUnits < 0 || member.AvailableUnits > stockCap ||
			member.ReceivedUnits < 0 || member.ReceivedUnits > stockCap || seen[member.AccountID] {
			return nil, nil, ErrInvalidExchange
		}
		seen[member.AccountID] = true
	}
	intakeUsed := make([]int64, len(result))
	clearings := make([]Clearing, 0)
	for producerIndex := range result {
		producer := &result[producerIndex]
		offered := producer.AvailableUnits * catalog.ClearingRatePPM / 1_000_000
		if offered <= 0 {
			continue
		}
		consumerIndexes := make([]int, 0)
		for index := range result {
			capacity := min64(catalog.StockIntakeCap-intakeUsed[index], stockCap-result[index].ReceivedUnits)
			if result[index].Consumes == producer.Produces && capacity > 0 {
				consumerIndexes = append(consumerIndexes, index)
			}
		}
		if len(consumerIndexes) == 0 {
			continue
		}
		base, remainder := offered/int64(len(consumerIndexes)), offered%int64(len(consumerIndexes))
		allocations := make([]Allocation, 0, len(consumerIndexes))
		var debited int64
		for order, consumerIndex := range consumerIndexes {
			requested := base
			if int64(order) < remainder {
				requested++
			}
			capacity := min64(catalog.StockIntakeCap-intakeUsed[consumerIndex], stockCap-result[consumerIndex].ReceivedUnits)
			units := min64(requested, capacity)
			if units <= 0 {
				continue
			}
			result[consumerIndex].ReceivedUnits += units
			intakeUsed[consumerIndex] += units
			debited += units
			allocations = append(allocations, Allocation{AccountID: result[consumerIndex].AccountID, Units: units})
		}
		if debited > 0 {
			producer.AvailableUnits -= debited
			clearings = append(clearings, Clearing{ProducerAccountID: producer.AccountID, Resource: producer.Produces, Allocations: allocations})
		}
	}
	return result, clearings, nil
}

func ClearNPC(catalog *Catalog, member MemberStock, stockCap int64) (MemberStock, *Clearing, error) {
	if catalog == nil || stockCap < 1 || !uuidPattern.MatchString(member.AccountID) || member.AvailableUnits < 0 || member.AvailableUnits > stockCap ||
		member.ReceivedUnits < 0 || member.ReceivedUnits > stockCap {
		return MemberStock{}, nil, ErrInvalidExchange
	}
	offered := member.AvailableUnits * catalog.NPCExchangePPM / 1_000_000
	units := min64(offered, min64(catalog.StockIntakeCap, stockCap-member.ReceivedUnits))
	if units <= 0 {
		return member, nil, nil
	}
	member.AvailableUnits -= units
	member.ReceivedUnits += units
	return member, &Clearing{ProducerAccountID: member.AccountID, Resource: member.Produces,
		Allocations: []Allocation{{AccountID: member.AccountID, Units: units}}, NPC: true}, nil
}

func min64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}
