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

func TestLedgerAccrualSaturatesWithReapplicableDelta(t *testing.T) {
	const catalogJSON = `{
      "schema_version": 1,
      "resources": [{
        "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
        "initial": "5.6765610215e6", "minimum": "0",
        "hardcap": {"amount": "9.87256122677e8", "reason_key": "resource.company_cash.cap.r1"}
      }],
      "generator_classes": []
    }`
	catalog, err := LoadCatalog([]byte(catalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewLedger(catalog, ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := ledger.ApplyAccrual(Transaction{Entries: []Entry{{
		ResourceID: "company.cash", Delta: mustDecimal(t, "9.81579561656e8"),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	want := Change{
		ResourceID: "company.cash", Before: "5.6765610215e6",
		Delta: "9.81579561655e8", After: "9.87256122677e8",
	}
	if len(receipt.Changes) != 1 || receipt.Changes[0] != want {
		t.Fatalf("receipt = %+v, want %+v", receipt, want)
	}
	assertReceiptReapplies(t, receipt.Changes[0])
}

func TestLedgerAccrualAtCapSucceedsWithoutChange(t *testing.T) {
	ledger := newTestLedger(t)
	if _, err := ledger.ApplyAccrual(Transaction{Entries: []Entry{{
		ResourceID: "company.cash", Delta: mustDecimal(t, "5e1"),
	}}}); err != nil {
		t.Fatal(err)
	}
	receipt, err := ledger.ApplyAccrual(Transaction{Entries: []Entry{{
		ResourceID: "company.cash", Delta: mustDecimal(t, "1e50"),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Changes) != 0 || ledger.Snapshot()["company.cash"] != "1e2" {
		t.Fatalf("receipt=%+v balance=%s", receipt, ledger.Snapshot()["company.cash"])
	}
}

func TestLedgerAccrualRejectsNegativeAndInvalidStartingStateAtomically(t *testing.T) {
	ledger := newTestLedger(t)
	before := ledger.Snapshot()
	_, err := ledger.ApplyAccrual(Transaction{Entries: []Entry{
		{ResourceID: "company.users", Delta: mustDecimal(t, "5e0")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "-1e0")},
	}})
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("negative accrual error = %v, want ErrInvalidTransaction", err)
	}
	assertSnapshotEqual(t, ledger.Snapshot(), before)

	ledger.balances["company.cash"] = mustDecimal(t, "1.01e2")
	before = ledger.Snapshot()
	_, err = ledger.ApplyAccrual(Transaction{Entries: []Entry{
		{ResourceID: "company.users", Delta: mustDecimal(t, "5e0")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "1e0")},
	}})
	if !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("invalid starting balance error = %v, want ErrInvalidTransaction", err)
	}
	assertSnapshotEqual(t, ledger.Snapshot(), before)
}

func TestLedgerAccrualSaturatesResourcesIndependentlyInSortedReceipt(t *testing.T) {
	ledger := newTestLedger(t)
	receipt, err := ledger.ApplyAccrual(Transaction{Entries: []Entry{
		{ResourceID: "company.users", Delta: mustDecimal(t, "5e0")},
		{ResourceID: "company.cash", Delta: mustDecimal(t, "6e1")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := []Change{
		{ResourceID: "company.cash", Before: "5e1", Delta: "5e1", After: "1e2"},
		{ResourceID: "company.users", Before: "1e1", Delta: "5e0", After: "1.5e1"},
	}
	if len(receipt.Changes) != len(want) {
		t.Fatalf("receipt = %+v, want %d changes", receipt, len(want))
	}
	for index := range want {
		if receipt.Changes[index] != want[index] {
			t.Fatalf("change %d = %+v, want %+v", index, receipt.Changes[index], want[index])
		}
		assertReceiptReapplies(t, receipt.Changes[index])
	}
}

func TestLedgerAccrualTwoMillionNearCapCases(t *testing.T) {
	const catalogJSON = `{
      "schema_version": 1,
      "resources": [{
        "id": "company.value", "scope": "company", "numeric_kind": "decimal",
        "initial": "0", "minimum": "0",
        "hardcap": {"amount": "1e1", "reason_key": "resource.company_value.cap.property"}
      }],
      "generator_classes": []
    }`
	catalog, err := LoadCatalog([]byte(catalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := NewLedger(catalog, ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	exponents := [...]int64{
		-8_999_999_999_999_999, -1_000_000, -100, -12, -1, 0,
		1, 11, 12, 100, 1_000_000, 8_999_999_999_999_999,
	}
	random := uint64(0x6a09e667f3bcc909)
	for caseIndex := 0; caseIndex < 2_000_000; caseIndex++ {
		first := nextLedgerRandom(&random)
		exponent := exponents[first%uint64(len(exponents))]
		capCoefficient := int64(200_000_000_000 + nextLedgerRandom(&random)%800_000_000_000)
		cap := decimal.New(float64(capCoefficient)/1e11, exponent).Quantize(decimal.CanonicalSignificantDigits)

		var before decimal.Decimal
		switch {
		case exponent == -8_999_999_999_999_999:
			before = decimal.Zero
		case caseIndex%4 == 0:
			before = decimal.Zero
		case caseIndex%4 == 1:
			before = decimal.New(9, exponent-1).Quantize(decimal.CanonicalSignificantDigits)
		default:
			beforeCoefficient := int64(100_000_000_000 + nextLedgerRandom(&random)%uint64(capCoefficient-99_999_999_999))
			before = decimal.New(float64(beforeCoefficient)/1e11, exponent).Quantize(decimal.CanonicalSignificantDigits)
		}
		if !before.IsStateValue() || before.Gt(cap) {
			caseIndex--
			continue
		}

		headroom := cap.Sub(before).Quantize(decimal.CanonicalSignificantDigits)
		delta := headroom
		if !headroom.IsStateValue() || headroom.Eq(decimal.Zero) {
			delta = cap
		} else {
			offset := float64(int64(nextLedgerRandom(&random)%5) - 2)
			candidate := decimal.New(headroom.Mantissa()+offset*1e-11, headroom.Exponent()).Quantize(decimal.CanonicalSignificantDigits)
			if candidate.IsStateValue() && candidate.Gte(decimal.Zero) {
				delta = candidate
			}
		}

		definition := catalog.resourceByID["company.value"]
		definition.Hardcap = &Hardcap{Amount: cap, ReasonKey: "resource.company_value.cap.property"}
		catalog.resourceByID["company.value"] = definition
		ledger.balances["company.value"] = before
		receipt, applyErr := ledger.ApplyAccrual(Transaction{Entries: []Entry{{
			ResourceID: "company.value", Delta: delta,
		}}})
		if applyErr != nil {
			t.Fatalf("case %d: before=%s delta=%s cap=%s: %v", caseIndex, before.String(), delta.String(), cap.String(), applyErr)
		}
		after := ledger.balances["company.value"]
		if after.Gt(cap) || after.Lt(decimal.Zero) {
			t.Fatalf("case %d: after=%s outside [0,%s]", caseIndex, after.String(), cap.String())
		}
		if len(receipt.Changes) > 1 {
			t.Fatalf("case %d: receipt has %d changes", caseIndex, len(receipt.Changes))
		}
		if len(receipt.Changes) == 1 {
			assertReceiptReapplies(t, receipt.Changes[0])
			if receipt.Changes[0].After != after.String() {
				t.Fatalf("case %d: receipt after=%s, ledger after=%s", caseIndex, receipt.Changes[0].After, after.String())
			}
		}
	}
}

func nextLedgerRandom(state *uint64) uint64 {
	value := *state
	value ^= value << 13
	value ^= value >> 7
	value ^= value << 17
	*state = value
	return value
}

func assertReceiptReapplies(t *testing.T, change Change) {
	t.Helper()
	before := mustDecimal(t, change.Before)
	delta := mustDecimal(t, change.Delta)
	if got := before.Add(delta).Quantize(decimal.CanonicalSignificantDigits).String(); got != change.After {
		t.Fatalf("receipt delta does not reapply: %s + %s = %s, want %s", change.Before, change.Delta, got, change.After)
	}
}

func assertSnapshotEqual(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("snapshot size = %d, want %d", len(got), len(want))
	}
	for resourceID, value := range want {
		if got[resourceID] != value {
			t.Fatalf("%s = %s, want %s", resourceID, got[resourceID], value)
		}
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
