package gameui

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

const cosmeticsArmFixturePath = "../../testdata/gameui/cosmetics-arm-v1.json"

func cosmeticsBundle(t *testing.T) production.CatalogBundle {
	t.Helper()
	species := petSpeciesBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range species.Artifacts {
		artifacts[name] = data
	}
	data, err := os.ReadFile("../../balance/testdata/cosmetics/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	artifacts["cosmetics"] = data
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := replaycatalog.Load(hash, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

type cosmeticsArmFixture struct {
	Version int                        `json:"version"`
	Cases   map[string]json.RawMessage `json:"cases"`
}

// AC10 (Go half): the projector emits the §7.1 arm; the shared fixture is what
// the TS decoder must accept. COSMETICS_ARM_UPDATE_FIXTURE=1 regenerates it.
func TestCosmeticsArmProjection(t *testing.T) {
	bundle := cosmeticsBundle(t)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	project := func(tier int64, cosmetics *cosmetic.State, pets map[string]pet.CareState) *cosmeticsArm {
		company, founder := featureStates(bundle)
		founder.WireVersion, company.Tier = 24, tier
		founder.Pets, founder.Cosmetics = pets, cosmetics
		features, err := projectFeatures(bundle, company, founder, now, false, 0)
		if err != nil || features.Cosmetics == nil {
			t.Fatalf("project: arm=%+v err=%v", features.Cosmetics, err)
		}
		return features.Cosmetics
	}
	const cat = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	pets := map[string]pet.CareState{cat: {}}
	cases := map[string]*cosmeticsArm{
		"locked-at-tier-0":     project(0, cosmetic.NewState(), map[string]pet.CareState{}),
		"acquirable-at-tier-1": project(1, cosmetic.NewState(), map[string]pet.CareState{}),
		"owned-no-wearer":      project(0, &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{}}, map[string]pet.CareState{}),
		"owned-worn":           project(1, &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{cat: "horse_armor"}}, pets),
		"owned-not-worn":       project(1, &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{}}, pets),
	}
	if arm := cases["locked-at-tier-0"]; !arm.Active || arm.Items[0].Acquirable || arm.Items[0].Lock == nil || arm.Items[0].Lock.Tier != 1 {
		t.Fatalf("tier-0 row = %+v", arm.Items[0])
	}
	if arm := cases["acquirable-at-tier-1"]; !arm.Items[0].Acquirable || arm.Items[0].Lock != nil || arm.Items[0].Owned {
		t.Fatalf("tier-1 row = %+v", arm.Items[0])
	}
	// §7.3: ownership survives a later Tier-0 run; no lock is shown once owned.
	if arm := cases["owned-no-wearer"]; !arm.Items[0].Owned || arm.Items[0].Acquirable || arm.Items[0].Lock != nil || len(arm.Wearers) != 0 {
		t.Fatalf("owned-at-tier-0 arm = %+v", arm)
	}
	if arm := cases["owned-worn"]; len(arm.Items[0].WornBy) != 1 || arm.Wearers[0].Worn == nil || *arm.Wearers[0].Worn != "horse_armor" {
		t.Fatalf("worn arm = %+v", arm)
	}
	// Absent below v24 is impossible under a cosmetics bundle (floor 24); an
	// unpinned bundle omits the arm entirely.
	company, founder := featureStates(petSpeciesBundle(t))
	founder.WireVersion = 23
	if features, err := projectFeatures(petSpeciesBundle(t), company, founder, now, false, 0); err != nil || features.Cosmetics != nil {
		t.Fatalf("unpinned cosmetics arm=%+v err=%v", features.Cosmetics, err)
	}
	fixture := cosmeticsArmFixture{Version: 1, Cases: map[string]json.RawMessage{}}
	for name, arm := range cases {
		encoded, err := json.Marshal(arm)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(encoded, []byte("price")) {
			t.Fatalf("%s arm carries a price: %s", name, encoded)
		}
		fixture.Cases[name] = encoded
	}
	encoded, err := json.MarshalIndent(fixture, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("COSMETICS_ARM_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile(cosmeticsArmFixturePath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(cosmeticsArmFixturePath)
	if err != nil || !bytes.Equal(pinned, encoded) {
		t.Fatalf("cosmetics arm fixture drifted; regenerate with COSMETICS_ARM_UPDATE_FIXTURE=1 (err=%v)", err)
	}
}
