package minigame

import (
	"os"
	"testing"
)

func TestTyperTenantRowLoads(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/pitch-typer-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	definition, ok := catalog.Definition("typer")
	if !ok || definition.Unlock.Kind != "tier_at_least" || definition.SoulGate != "unrelated" || definition.Payout.PayoutScoreFactID != "typer.clean_lines" {
		t.Fatalf("typer definition=%+v ok=%v", definition, ok)
	}
}
