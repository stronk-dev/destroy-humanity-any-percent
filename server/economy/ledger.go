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
	ErrResourceScope      = errors.New("economy resource belongs to another ledger scope")
	ErrBelowMinimum       = errors.New("resource balance below minimum")
	ErrAboveHardcap       = errors.New("resource balance above hardcap")
	ErrInvalidRestore     = errors.New("invalid economy ledger restore")
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
	scope    Scope
	balances map[string]decimal.Decimal
}

func NewLedger(catalog *Catalog, scope Scope) (*Ledger, error) {
	if catalog == nil {
		return nil, fmt.Errorf("%w: nil catalog", ErrInvalidTransaction)
	}
	if !validScope(scope) {
		return nil, fmt.Errorf("%w: unsupported ledger scope %q", ErrInvalidTransaction, scope)
	}
	balances := make(map[string]decimal.Decimal)
	for _, resource := range catalog.resources {
		if resource.Scope == scope {
			balances[resource.ID] = resource.Initial
		}
	}
	return &Ledger{catalog: catalog, scope: scope, balances: balances}, nil
}

// RestoreLedger is the persistence-only constructor. It requires one canonical
// balance for every resource in scope and does not expose a mutation path.
func RestoreLedger(catalog *Catalog, scope Scope, snapshot map[string]string) (*Ledger, error) {
	ledger, err := NewLedger(catalog, scope)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRestore, err)
	}
	if len(snapshot) != len(ledger.balances) {
		return nil, fmt.Errorf("%w: got %d balances, want %d", ErrInvalidRestore, len(snapshot), len(ledger.balances))
	}

	restored := make(map[string]decimal.Decimal, len(snapshot))
	for resourceID, encoded := range snapshot {
		definition, exists := catalog.resourceByID[resourceID]
		if !exists {
			return nil, fmt.Errorf("%w: balances.%s: %w", ErrInvalidRestore, resourceID, ErrUnknownResource)
		}
		if definition.Scope != scope {
			return nil, fmt.Errorf("%w: balances.%s: %w", ErrInvalidRestore, resourceID, ErrResourceScope)
		}
		value, parseErr := decimal.ParseCanonical(encoded)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: balances.%s: %v", ErrInvalidRestore, resourceID, parseErr)
		}
		if value.Lt(definition.Minimum) {
			return nil, fmt.Errorf("%w: balances.%s: %w", ErrInvalidRestore, resourceID, ErrBelowMinimum)
		}
		if definition.Hardcap != nil && value.Gt(definition.Hardcap.Amount) {
			return nil, fmt.Errorf("%w: balances.%s: %w", ErrInvalidRestore, resourceID, ErrAboveHardcap)
		}
		restored[resourceID] = value
	}
	ledger.balances = restored
	return ledger, nil
}

func (l *Ledger) Scope() Scope {
	return l.scope
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
	entriesByResource := make(map[string][]decimal.Decimal)
	for index, entry := range transaction.Entries {
		definition, exists := l.catalog.resourceByID[entry.ResourceID]
		if !exists {
			return Receipt{}, fmt.Errorf("%w: entry %d: %q", ErrUnknownResource, index, entry.ResourceID)
		}
		if definition.Scope != l.scope {
			return Receipt{}, fmt.Errorf("%w: entry %d: %q belongs to %q, ledger is %q", ErrResourceScope, index, entry.ResourceID, definition.Scope, l.scope)
		}
		if !entry.Delta.IsStateValue() {
			return Receipt{}, fmt.Errorf("%w: entry %d has non-state delta", ErrInvalidTransaction, index)
		}
		entriesByResource[entry.ResourceID] = append(entriesByResource[entry.ResourceID], entry.Delta)
	}

	net := make(map[string]decimal.Decimal, len(entriesByResource))
	resourceIDs := make([]string, 0, len(entriesByResource))
	for resourceID, entries := range entriesByResource {
		net[resourceID] = decimal.SumDeterministic(entries)
		if !net[resourceID].IsStateValue() {
			return Receipt{}, fmt.Errorf("%w: aggregate for %q is outside the finite Decimal domain", ErrInvalidTransaction, resourceID)
		}
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
