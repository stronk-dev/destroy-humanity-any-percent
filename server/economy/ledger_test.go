package economy

import (
	"errors"
	"testing"

	"cloud-clicker/server/decimal"
)

const ledgerCatalogJSON = `{
  "schema_version": 1,
  "resources": [
    {
      "id": "company.cash",
      "scope": "company",
      "numeric_kind": "decimal",
      "initial": "5e1",
      "minimum": "0",
      "hardcap": {"amount": "1e2", "reason_key": "resource.company_cash.cap.test"}
    },
    {
      "id": "company.users",
      "scope": "company",
      "numeric_kind": "decimal",
      "initial": "1e1",
      "minimum": "0",
      "hardcap": null
    }
  ],
  "generator_classes": []
}`

func newTestLedger(t *testing.T) *Ledger {
	t.Helper()
	catalog, err := LoadCatalog([]byte(ledgerCatalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewLedger(catalog, ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	return ledger
}

func mustDecimal(t *testing.T, source string) decimal.Decimal {
	t.Helper()
	value, err := decimal.ParseCanonical(source)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestLedgerAppliesAtomicallyAndReturnsSortedReceipt(t *testing.T) {
	ledger := newTestLedger(t)
	receipt, err := ledger.Apply(Transaction{Entries: []Entry{
		{ResourceID: "company.users", Delta: mustDecimal(t, "5e0")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "2.5e1")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "-5e0")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Changes) != 2 {
		t.Fatalf("changes = %d, want 2", len(receipt.Changes))
	}
	if got, want := receipt.Changes[0], (Change{ResourceID: "company.cash", Before: "5e1", Delta: "2e1", After: "7e1"}); got != want {
		t.Fatalf("first change = %#v, want %#v", got, want)
	}
	if got, want := receipt.Changes[1], (Change{ResourceID: "company.users", Before: "1e1", Delta: "5e0", After: "1.5e1"}); got != want {
		t.Fatalf("second change = %#v, want %#v", got, want)
	}
}

func TestLedgerRejectsWholeTransactionOnAnyInvariantFailure(t *testing.T) {
	ledger := newTestLedger(t)
	before := ledger.Snapshot()
	_, err := ledger.Apply(Transaction{Entries: []Entry{
		{ResourceID: "company.users", Delta: mustDecimal(t, "5e0")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "-6e1")},
	}})
	if !errors.Is(err, ErrBelowMinimum) {
		t.Fatalf("error = %v, want ErrBelowMinimum", err)
	}
	after := ledger.Snapshot()
	for resourceID, balance := range before {
		if after[resourceID] != balance {
			t.Fatalf("%s mutated from %s to %s", resourceID, balance, after[resourceID])
		}
	}
}

func TestLedgerRejectsHardcapUnknownAndNonFinite(t *testing.T) {
	tests := []struct {
		name  string
		entry Entry
		want  error
	}{
		{name: "hardcap", entry: Entry{ResourceID: "company.cash", Delta: mustDecimal(t, "5.1e1")}, want: ErrAboveHardcap},
		{name: "unknown", entry: Entry{ResourceID: "company.missing", Delta: decimal.One}, want: ErrUnknownResource},
		{name: "non-finite", entry: Entry{ResourceID: "company.cash", Delta: decimal.NaN}, want: ErrInvalidTransaction},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ledger := newTestLedger(t)
			before := ledger.Snapshot()
			if _, err := ledger.Apply(Transaction{Entries: []Entry{test.entry}}); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
			if ledger.Snapshot()["company.cash"] != before["company.cash"] {
				t.Fatal("failed transaction mutated the ledger")
			}
		})
	}
}

func TestLedgerRejectsAggregateNumericOverflow(t *testing.T) {
	catalogJSON := `{
      "schema_version": 1,
      "resources": [{
        "id": "company.value", "scope": "company", "numeric_kind": "decimal",
        "initial": "0", "minimum": "0", "hardcap": null
      }],
      "generator_classes": []
    }`
	catalog, err := LoadCatalog([]byte(catalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	ledger, _ := NewLedger(catalog, ScopeCompany)
	maximum := mustDecimal(t, "9e8999999999999999")
	_, err = ledger.Apply(Transaction{Entries: []Entry{
		{ResourceID: "company.value", Delta: maximum},
		{ResourceID: "company.value", Delta: maximum},
	}})
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("error = %v, want ErrInvalidTransaction", err)
	}
	if got := ledger.Snapshot()["company.value"]; got != "0" {
		t.Fatalf("balance = %s, want 0", got)
	}
}

func TestLedgerAggregationIsPermutationInvariantAtDomainEdge(t *testing.T) {
	catalogJSON := `{
      "schema_version": 1,
      "resources": [{
        "id": "company.value", "scope": "company", "numeric_kind": "decimal",
        "initial": "0", "minimum": "0", "hardcap": null
      }],
      "generator_classes": []
    }`
	catalog, err := LoadCatalog([]byte(catalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	positive := mustDecimal(t, "9e8999999999999999")
	negative := positive.Neg()
	permutations := [][]decimal.Decimal{
		{positive, positive, negative, negative},
		{positive, negative, positive, negative},
		{negative, negative, positive, positive},
	}
	for _, values := range permutations {
		ledger, _ := NewLedger(catalog, ScopeCompany)
		entries := make([]Entry, len(values))
		for index, value := range values {
			entries[index] = Entry{ResourceID: "company.value", Delta: value}
		}
		if _, err := ledger.Apply(Transaction{Entries: entries}); err != nil {
			t.Fatal(err)
		}
		if got := ledger.Snapshot()["company.value"]; got != "0" {
			t.Fatalf("balance = %s, want zero", got)
		}
	}
}

func TestLedgerAggregatesSubResolutionSourcesBeforeCommit(t *testing.T) {
	catalogJSON := `{
      "schema_version": 1,
      "resources": [{
        "id": "company.bank", "scope": "company", "numeric_kind": "decimal",
        "initial": "1e100", "minimum": "0",
        "hardcap": {"amount": "1e200", "reason_key": "resource.company_bank.cap.test"}
      }],
      "generator_classes": []
    }`
	catalog, err := LoadCatalog([]byte(catalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	const sourceCount = 1_000_000
	tiny := mustDecimal(t, "1e87")
	entries := make([]Entry, sourceCount)
	for index := range entries {
		entries[index] = Entry{ResourceID: "company.bank", Delta: tiny}
	}

	aggregated, _ := NewLedger(catalog, ScopeCompany)
	if _, err := aggregated.Apply(Transaction{Entries: entries}); err != nil {
		t.Fatal(err)
	}
	if got := aggregated.Snapshot()["company.bank"]; got != "1.0000001e100" {
		t.Fatalf("aggregated balance = %s, want 1.0000001e100", got)
	}

	perEntry, _ := NewLedger(catalog, ScopeCompany)
	for index := 0; index < sourceCount; index++ {
		if _, err := perEntry.Apply(Transaction{Entries: entries[index : index+1]}); err != nil {
			t.Fatal(err)
		}
	}
	if got := perEntry.Snapshot()["company.bank"]; got != "1e100" {
		t.Fatalf("per-entry balance = %s, want unchanged 1e100", got)
	}
}

func TestLedgerEnforcesScopeBoundary(t *testing.T) {
	catalog, err := LoadCatalog(loadKernelFixture(t).Catalog)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewLedger(catalog, ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	if ledger.Scope() != ScopeCompany {
		t.Fatalf("scope = %q, want company", ledger.Scope())
	}
	if _, exists := ledger.Balance("founder.reputation"); exists {
		t.Fatal("company ledger exposed a founder balance")
	}
	_, err = ledger.Apply(Transaction{Entries: []Entry{{
		ResourceID: "founder.reputation",
		Delta:      decimal.One,
	}}})
	if !errors.Is(err, ErrResourceScope) {
		t.Fatalf("error = %v, want ErrResourceScope", err)
	}
	if _, err := NewLedger(catalog, Scope("universe")); !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("invalid scope error = %v, want ErrInvalidTransaction", err)
	}
}

func TestRestoreLedgerRequiresExactCanonicalScopedSnapshot(t *testing.T) {
	catalog, err := LoadCatalog(loadKernelFixture(t).Catalog)
	if err != nil {
		t.Fatal(err)
	}
	valid := map[string]string{
		"company.cash":  "5e1",
		"company.users": "1e1",
	}
	ledger, err := RestoreLedger(catalog, ScopeCompany, valid)
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.Snapshot()["company.cash"]; got != "5e1" {
		t.Fatalf("restored cash = %s", got)
	}

	tests := []map[string]string{
		{"company.cash": "5e1"},
		{"company.cash": "5e1", "company.users": "NaN"},
		{"company.cash": "5e1", "founder.reputation": "0"},
		{"company.cash": "1e1000001", "company.users": "1e1"},
	}
	for _, snapshot := range tests {
		if _, err := RestoreLedger(catalog, ScopeCompany, snapshot); !errors.Is(err, ErrInvalidRestore) {
			t.Fatalf("RestoreLedger(%v) error = %v", snapshot, err)
		}
	}
}
