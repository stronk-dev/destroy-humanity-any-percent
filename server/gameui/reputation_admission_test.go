package gameui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

func reputationProjectionSource(t *testing.T) (production.CatalogBundle, *save.State, *save.State, time.Time) {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/replay/reputation-tree-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "f9b129e36af5b536f67c7eddb8d6088c76ce5cb0b170cfeec6a5172e8a009782" {
		t.Fatal("projection source changed")
	}
	var source struct {
		Bundles map[string]struct {
			ConstantsHash string            `json:"constants_hash"`
			Artifacts     map[string]string `json:"artifacts"`
		} `json:"bundles"`
		Cases []struct {
			Name         string          `json:"name"`
			StateVersion int             `json:"state_version"`
			PreState     json.RawMessage `json:"pre_state"`
		} `json:"cases"`
		ExitCases []struct {
			Company struct {
				ConstantsHash string `json:"constants_hash"`
				Case          struct {
					PreState json.RawMessage `json:"pre_state"`
				} `json:"case"`
			} `json:"company"`
		} `json:"exit_cases"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	tree := source.Bundles["tree"]
	artifacts := map[string][]byte{}
	for name, value := range tree.Artifacts {
		artifacts[name] = []byte(value)
	}
	bundle, err := replaycatalog.Load(tree.ConstantsHash, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	index := slices.IndexFunc(source.Cases, func(row struct {
		Name         string          `json:"name"`
		StateVersion int             `json:"state_version"`
		PreState     json.RawMessage `json:"pre_state"`
	}) bool {
		return row.Name == "rejects-owned"
	})
	if index < 0 {
		t.Fatal("missing pinned owned control")
	}
	row := source.Cases[index]
	founder, err := save.RestoreState(row.PreState, row.StateVersion, bundle.Economy, economy.ScopeFounder, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	companyIndex := slices.IndexFunc(source.ExitCases, func(row struct {
		Company struct {
			ConstantsHash string `json:"constants_hash"`
			Case          struct {
				PreState json.RawMessage `json:"pre_state"`
			} `json:"case"`
		} `json:"company"`
	}) bool {
		return row.Company.ConstantsHash == bundle.ConstantsHash
	})
	if companyIndex < 0 {
		t.Fatal("missing matching Company control")
	}
	company, err := save.RestoreState(source.ExitCases[companyIndex].Company.Case.PreState, 18, bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.ValidateFoundationState(company) != nil || bundle.ValidateFoundationState(founder) != nil {
		t.Fatal("invalid pinned projection control")
	}
	now := time.UnixMilli(founder.FiscalPeriodOpenedWallMS).UTC()
	if now.Before(company.EvaluatedThrough) {
		now = company.EvaluatedThrough
	}
	return bundle, company, founder, now
}

func cloneReputationProjectionFounder(t *testing.T, bundle production.CatalogBundle, state *save.State) *save.State {
	t.Helper()
	encoded, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	copy, err := save.RestoreState(encoded, save.VersionForState(state), bundle.Economy, economy.ScopeFounder, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	return copy
}

func TestReputationProjectionAdmitsConsistentPinnedProfiles(t *testing.T) {
	bundle, company, base, now := reputationProjectionSource(t)
	for _, name := range []string{"empty", "owned-p05", "historical-unknown-plus-p05"} {
		t.Run(name, func(t *testing.T) {
			founder := cloneReputationProjectionFounder(t, bundle, base)
			want := "1.0015e0"
			if name == "empty" {
				founder.ReputationSpent, founder.ReputationUnlockPPM, founder.ReputationNodesOwned = 0, 0, []string{}
				want = "1e0"
			}
			if name == "historical-unknown-plus-p05" {
				founder.ReputationNodesOwned = []string{"reputation.retired.unknown", "reputation.unlock.p05"}
			}
			if err := bundle.ValidateFoundationState(founder); err != nil {
				t.Fatal(err)
			}
			frozen, err := production.FrozenFounderContributions(bundle, founder)
			if err != nil {
				t.Fatal(err)
			}
			contributions, err := production.ResolveFrozenContributions(bundle.Economy, frozen)
			if err != nil {
				t.Fatal(err)
			}
			arm, err := projectReputation(bundle, founder, contributions)
			if err != nil || arm == nil || arm.BonusFactorNextRun != want || arm.BonusFactorThisRun == nil || *arm.BonusFactorThisRun != want {
				t.Fatalf("isolated arm=%+v err=%v", arm, err)
			}
			projector := &Projector{catalogs: production.ReplayCatalogSet{bundle.ConstantsHash: bundle}}
			encoded, err := projector.InitialGameUISnapshot(context.Background(), bundle.ConstantsHash, "01986666-9f20-4000-8000-000000000001", company, founder, frozen, now)
			if err != nil {
				t.Fatalf("public positive projection: %v", err)
			}
			var snapshot struct {
				Features struct {
					Reputation reputationArm `json:"reputation"`
				} `json:"features"`
			}
			if err := json.Unmarshal(encoded, &snapshot); err != nil || snapshot.Features.Reputation.BonusFactorNextRun != want || snapshot.Features.Reputation.BonusFactorThisRun == nil || *snapshot.Features.Reputation.BonusFactorThisRun != want {
				t.Fatalf("public arm=%+v err=%v", snapshot.Features.Reputation, err)
			}
		})
	}
}

func TestReputationProjectionRejectsInvalidPinnedOwnership(t *testing.T) {
	bundle, company, base, now := reputationProjectionSource(t)
	frozen, err := production.FrozenFounderContributions(bundle, base)
	if err != nil {
		t.Fatal(err)
	}
	contributions, err := production.ResolveFrozenContributions(bundle.Economy, frozen)
	if err != nil {
		t.Fatal(err)
	}
	mutations := []struct {
		name   string
		change func(*save.State)
	}{
		{"nil-owned", func(state *save.State) { state.ReputationNodesOwned, state.ReputationUnlockPPM = nil, 0 }},
		{"duplicate-owned", func(state *save.State) {
			state.ReputationNodesOwned = []string{"reputation.unlock.p05", "reputation.unlock.p05"}
		}},
		{"unsorted-owned", func(state *save.State) {
			state.ReputationNodesOwned = []string{"reputation.unlock.p05", "reputation.starter.cash_small"}
		}},
		{"nonmechanical-owned", func(state *save.State) {
			state.ReputationNodesOwned = []string{"NOT mechanical", "reputation.unlock.p05"}
		}},
		{"empty-false-mirror", func(state *save.State) { state.ReputationNodesOwned = []string{} }},
		{"owned-zero-mirror", func(state *save.State) { state.ReputationUnlockPPM = 0 }},
		{"owned-other-mirror", func(state *save.State) { state.ReputationUnlockPPM = 250_000 }},
	}
	for _, mutation := range mutations {
		for _, path := range []string{"isolated", "public-initial"} {
			t.Run(mutation.name+"/"+path, func(t *testing.T) {
				founder := cloneReputationProjectionFounder(t, bundle, base)
				mutation.change(founder)
				if path == "isolated" {
					arm, err := projectReputation(bundle, founder, contributions)
					if !errors.Is(err, ErrInvalidProjection) || arm != nil {
						t.Fatalf("invalid ownership projected: arm=%+v err=%v", arm, err)
					}
				} else {
					projector := &Projector{catalogs: production.ReplayCatalogSet{bundle.ConstantsHash: bundle}}
					encoded, err := projector.InitialGameUISnapshot(context.Background(), bundle.ConstantsHash, "01986666-9f20-4000-8000-000000000001", company, founder, frozen, now)
					if !errors.Is(err, ErrInvalidProjection) || len(encoded) != 0 {
						t.Fatalf("invalid ownership in public projection: bytes=%d err=%v", len(encoded), err)
					}
				}
			})
		}
	}
}
