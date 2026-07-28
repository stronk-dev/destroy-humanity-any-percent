package save

import (
	"errors"
	"testing"

	"cloud-clicker/server/economy"
)

const stateCatalogJSON = `{
  "schema_version": 1,
  "resources": [{
    "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0",
    "hardcap": {"amount": "1e100", "reason_key": "resource.company_cash.cap.test"}
  }],
  "generator_classes": []
}`

func stateCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	catalog, err := economy.LoadCatalog([]byte(stateCatalogJSON))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestLedgerStateRoundTrip(t *testing.T) {
	catalog := stateCatalog(t)
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{
		"company.cash": "1.23456789012e42",
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := EncodeLedger(ledger)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLedger(encoded, CurrentVersion, catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	if got := restored.Snapshot()["company.cash"]; got != "1.23456789012e42" {
		t.Fatalf("balance = %s", got)
	}
}

func TestRestoreRejectsPoisonedAndMalformedState(t *testing.T) {
	tests := []string{
		`{"balances":{"company.cash":"NaN"}}`,
		`{"balances":{"company.cash":1}}`,
		`{"balances":{}}`,
		`{"balances":{"company.cash":"0","company.extra":"0"}}`,
		`{"balances":{"company.cash":"0"},"unknown":true}`,
	}
	for _, data := range tests {
		if _, err := RestoreLedger([]byte(data), CurrentVersion, stateCatalog(t), economy.ScopeCompany); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("RestoreLedger(%s) error = %v", data, err)
		}
	}
}

func TestRestoreRejectsFutureVersion(t *testing.T) {
	_, err := RestoreLedger([]byte(`{"balances":{"company.cash":"0"}}`), CurrentVersion+1, stateCatalog(t), economy.ScopeCompany)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("error = %v", err)
	}
}

func TestConstantsHashUsesExactArtifactBytes(t *testing.T) {
	first := ConstantsHash([]byte("{}"))
	second := ConstantsHash([]byte("{}\n"))
	if first == second || len(first) != len("sha256:")+64 {
		t.Fatalf("unexpected hashes %q %q", first, second)
	}
}

func TestStreamOwnerScopeValidation(t *testing.T) {
	ledger, err := economy.NewLedger(stateCatalog(t), economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	validHash := ConstantsHash([]byte(stateCatalogJSON))
	context := WriteContext{Cause: "test"}
	valid := StreamKey{OwnerKind: OwnerFounder, OwnerID: "11111111-1111-4111-8111-111111111111", Scope: economy.ScopeCompany}
	if err := validateWrite(valid, validHash, ledger, context); err != nil {
		t.Fatal(err)
	}
	invalid := []StreamKey{
		{OwnerKind: OwnerGuild, OwnerID: valid.OwnerID, Scope: economy.ScopeCompany},
		{OwnerKind: OwnerWorld, OwnerID: valid.OwnerID, Scope: economy.ScopeFounder},
		{OwnerKind: OwnerFounder, OwnerID: "not-a-uuid", Scope: economy.ScopeCompany},
	}
	for _, key := range invalid {
		if err := validateWrite(key, validHash, ledger, context); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("validateWrite(%+v) error = %v", key, err)
		}
	}
}
