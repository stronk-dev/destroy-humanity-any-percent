package economy

import (
	"errors"
	"fmt"
	"sort"

	"cloud-clicker/server/decimal"
)

var (
	ErrInvalidTransaction = errors.New("invalid economy transaction")
	ErrUnknownResource    = errors.New("unknown economy resource")
	ErrBelowMinimum       = errors.New("resource balance below minimum")
	ErrAboveHardcap       = errors.New("resource balance above hardcap")
)

type Entry struct {
	ResourceID string
	Delta      decimal.Decimal
}

type Transaction struct {
	Entries []Entry
}

type Change struct {
	ResourceID string
	Before     string
	Delta      string
	After      string
}

type Receipt struct {
	Changes []Change
}

type Ledger struct {
	catalog  *Catalog
	balances map[string]decimal.Decimal
}

func NewLedger(catalog *Catalog) (*Ledger, error) {
	if catalog == nil {
		return nil, fmt.Errorf("%w: nil catalog", ErrInvalidTransaction)
	}
	balances := make(map[string]decimal.Decimal, len(catalog.resources))
	for _, resource := range catalog.resources {
		balances[resource.ID] = resource.Initial
	}
	return &Ledger{catalog: catalog, balances: balances}, nil
}

func (l *Ledger) Balance(resourceID string) (decimal.Decimal, bool) {
	balance, exists := l.balances[resourceID]
	return balance, exists
}

func (l *Ledger) Snapshot() map[string]string {
	snapshot := make(map[string]string, len(l.balances))
	for resourceID, balance := range l.balances {
		snapshot[resourceID] = balance.String()
	}
	return snapshot
}

func (l *Ledger) Apply(transaction Transaction) (Receipt, error) {
	net := make(map[string]decimal.Decimal)
	for index, entry := range transaction.Entries {
		if _, exists := l.catalog.resourceByID[entry.ResourceID]; !exists {
			return Receipt{}, fmt.Errorf("%w: entry %d: %q", ErrUnknownResource, index, entry.ResourceID)
		}
		if !entry.Delta.IsStateValue() {
			return Receipt{}, fmt.Errorf("%w: entry %d has non-state delta", ErrInvalidTransaction, index)
		}
		net[entry.ResourceID] = net[entry.ResourceID].Add(entry.Delta)
		if !net[entry.ResourceID].IsStateValue() {
			return Receipt{}, fmt.Errorf("%w: aggregate for %q is outside the finite Decimal domain at entry %d (%s; mantissa=%g exponent=%d)", ErrInvalidTransaction, entry.ResourceID, index, net[entry.ResourceID].String(), net[entry.ResourceID].Mantissa(), net[entry.ResourceID].Exponent())
		}
	}

	resourceIDs := make([]string, 0, len(net))
	for resourceID := range net {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)

	prospective := make(map[string]decimal.Decimal, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		definition := l.catalog.resourceByID[resourceID]
		next := l.balances[resourceID].Add(net[resourceID]).Quantize(decimal.CanonicalSignificantDigits)
		if !next.IsStateValue() {
			return Receipt{}, fmt.Errorf("%w: result for %q is outside the finite Decimal domain", ErrInvalidTransaction, resourceID)
		}
		if next.Lt(definition.Minimum) {
			return Receipt{}, fmt.Errorf("%w: %q", ErrBelowMinimum, resourceID)
		}
		if definition.Hardcap != nil && next.Gt(definition.Hardcap.Amount) {
			return Receipt{}, fmt.Errorf("%w: %q", ErrAboveHardcap, resourceID)
		}
		prospective[resourceID] = next
	}

	receipt := Receipt{Changes: make([]Change, 0, len(resourceIDs))}
	for _, resourceID := range resourceIDs {
		before := l.balances[resourceID]
		after := prospective[resourceID]
		if before.Eq(after) {
			continue
		}
		l.balances[resourceID] = after
		receipt.Changes = append(receipt.Changes, Change{
			ResourceID: resourceID,
			Before:     before.String(),
			Delta:      after.Sub(before).Quantize(decimal.CanonicalSignificantDigits).String(),
			After:      after.String(),
		})
	}
	return receipt, nil
}
