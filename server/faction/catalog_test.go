package faction

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/commons"
)

func TestPhase0CatalogAndCycleValidation(t *testing.T) {
	commonsData, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalog, err := commons.LoadCatalog(commonsData)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/factions/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	band := CompactTitheBand{MinimumPPM: commonsCatalog.MinimumTithePPM, DefaultPPM: commonsCatalog.DefaultTithePPM, MaximumPPM: commonsCatalog.MaximumTithePPM}
	catalog, err := LoadCatalog(data, band)
	if err != nil || catalog.StockCap != 100_000 || catalog.StockIntervalMS != 60_000 || len(catalog.Factions) != 4 {
		t.Fatalf("catalog=%+v err=%v", catalog, err)
	}
	openSource, ok := catalog.Faction("open_source")
	if !ok || openSource.Compact == nil || openSource.Compact.TithePPM != 130_000 || openSource.Produces != "libraries" {
		t.Fatalf("open source=%+v ok=%v", openSource, ok)
	}

	var selfRaw rawCatalog
	if err := json.Unmarshal(data, &selfRaw); err != nil {
		t.Fatal(err)
	}
	selfRaw.Factions[0].Consumes = selfRaw.Factions[0].Produces
	selfConsuming, _ := json.Marshal(selfRaw)
	if _, err := LoadCatalog(selfConsuming, band); err == nil {
		t.Fatal("self-consuming fixture accepted")
	}
	var cycleRaw rawCatalog
	if err := json.Unmarshal(data, &cycleRaw); err != nil {
		t.Fatal(err)
	}
	cycleRaw.Factions[0].Consumes = "hype"
	cycleRaw.Factions[2].Consumes = "compliance"
	twoCycles, _ := json.Marshal(cycleRaw)
	if _, err := LoadCatalog(twoCycles, band); err == nil {
		t.Fatal("two-cycle fixture accepted")
	}
}

func TestCatalogRejectsUnknownAndTrailingFields(t *testing.T) {
	commonsData, _ := os.ReadFile("../../balance/commons/phase0.json")
	commonsCatalog, _ := commons.LoadCatalog(commonsData)
	band := CompactTitheBand{MinimumPPM: commonsCatalog.MinimumTithePPM, DefaultPPM: commonsCatalog.DefaultTithePPM, MaximumPPM: commonsCatalog.MaximumTithePPM}
	data, _ := os.ReadFile("../../balance/factions/phase0.json")
	if _, err := LoadCatalog(append(data, []byte(` {}`)...), band); err == nil {
		t.Fatal("trailing JSON accepted")
	}
	unknown := bytes.Replace(data, []byte(`"stock_cap": 100000`), []byte(`"stock_cap": 100000, "unknown": true`), 1)
	if _, err := LoadCatalog(unknown, band); err == nil {
		t.Fatal("unknown field accepted")
	}
}
