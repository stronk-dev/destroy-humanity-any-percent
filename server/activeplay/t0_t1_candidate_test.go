package activeplay

import (
	"os"
	"testing"

	"cloud-clicker/server/economy"
)

func TestT0T1OpportunitiesCandidateLoadsAgainstCandidateEconomy(t *testing.T) {
	economyBytes, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	economyCatalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	opportunityBytes, err := os.ReadFile("../../balance/testdata/t0-t1/opportunities-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(opportunityBytes, economyCatalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Effects()) != 4 || catalog.Combo.HardcapReasonKey != "cap.active_combo" {
		t.Fatalf("opportunities candidate = %+v", catalog)
	}
}
