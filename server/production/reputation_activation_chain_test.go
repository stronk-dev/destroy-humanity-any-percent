package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

type reputationEarlierCorpusBundle struct {
	ConstantsHash string            `json:"constants_hash"`
	Artifacts     map[string]string `json:"artifacts"`
}

type reputationEarlierCorpusCase struct {
	SourceBundle string                  `json:"source_bundle"`
	NextBundle   string                  `json:"next_bundle"`
	Case         crossRuntimeFounderCase `json:"case"`
}

type reputationEarlierCorpus struct {
	Version        int                                      `json:"schema_version"`
	SourceVersions []int                                    `json:"source_versions"`
	Bundles        map[string]reputationEarlierCorpusBundle `json:"bundles"`
	Cases          []reputationEarlierCorpusCase            `json:"cases"`
}

// These are fixture epochs, not new production epochs. Keep the minigame IDs
// and Fiscal rows stable across the boundary: this measures save activation,
// not an unrelated content-key migration.
func reputationEarlierBundles(t *testing.T) ([]CatalogBundle, CatalogBundle) {
	t.Helper()
	legacy, foundation := foundationTestBundles(t)
	minigames, _ := founderFeatureBundles(t, foundation)
	data, err := os.ReadFile("../../testdata/minigame/pitch-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	minigames.Artifacts["minigames"] = data
	minigames.Minigames, err = minigame.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	minigames.ConstantsHash, err = save.ConstantsHashArtifacts(minigames.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	_, pets := founderFeatureBundles(t, foundation)
	pets.Artifacts["minigames"], pets.Minigames = data, minigames.Minigames
	pets.ConstantsHash, err = save.ConstantsHashArtifacts(pets.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	fiscal := fiscalFeatureBundle(t, pets)
	soul := soulFeatureBundle(t, fiscal)
	active := activeContentBundle(t)
	next := soul
	next.Artifacts = cloneArtifactMap(soul.Artifacts)
	next.Artifacts["pitch"], next.Pitch = active.Artifacts["pitch"], active.Pitch
	next.Artifacts["minigame_api"], next.MinigameAPI = active.Artifacts["minigame_api"], active.MinigameAPI
	var root map[string]any
	if err := json.Unmarshal(next.Artifacts["economy"], &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{
		"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider,
	})
	next.Artifacts["economy"], err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	next.Economy, err = economy.LoadCatalog(next.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	next.Artifacts["reputation_tree"], err = os.ReadFile("../../balance/testdata/reputation-tree/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	// Legacy fixture economies do not declare the later starter upgrade.
	// Retain the fixture's unchanged unlock chain: no starters are bought or
	// applied by this population, and this is not a starter migration proof.
	var treeRoot map[string]any
	if err := json.Unmarshal(next.Artifacts["reputation_tree"], &treeRoot); err != nil {
		t.Fatal(err)
	}
	unlockNodes := []any{}
	for _, node := range treeRoot["nodes"].([]any) {
		if node.(map[string]any)["kind"] == "bonus_unlock" {
			unlockNodes = append(unlockNodes, node)
		}
	}
	treeRoot["nodes"] = unlockNodes
	next.Artifacts["reputation_tree"], err = json.Marshal(treeRoot)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	next.ReputationTree, err = reputation.LoadTree(next.Artifacts["reputation_tree"], reputation.Declarations{Economy: next.Economy, CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
	if err != nil || !next.valid(next.ConstantsHash) {
		t.Fatalf("invalid next activation bundle: %v", err)
	}
	return []CatalogBundle{legacy, foundation, minigames, pets, fiscal, soul}, next
}

func TestReputationEarlierFounderActivationChain(t *testing.T) {
	sources, fixtureNext := reputationEarlierBundles(t)
	sources = append(sources, activeContentBundle(t))
	wantVersions := []int{14, 16, 17, 18, 19, 20, 21}
	if len(sources) != len(wantVersions) {
		t.Fatal("incomplete earlier-Founder population")
	}
	corpus := reputationEarlierCorpus{Version: 1, SourceVersions: wantVersions,
		Bundles: map[string]reputationEarlierCorpusBundle{}, Cases: []reputationEarlierCorpusCase{}}
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	for index, source := range sources {
		version, companyVersion := source.versionFloors()
		if version != wantVersions[index] {
			t.Fatalf("source[%d] version=%d want=%d", index, version, wantVersions[index])
		}
		t.Run(fmt.Sprintf("v%d_to_v22", version), func(t *testing.T) {
			next := fixtureNext
			if version == 21 {
				next = reputationContentBundle(t)
			}
			if nextFounder, _ := next.versionFloors(); nextFounder != 22 {
				t.Fatal("incorrect next Founder floor")
			}
			if !source.valid(source.ConstantsHash) {
				t.Fatal("invalid source bundle")
			}
			source.Next = &next
			initial := replayFounderFixtureState(t, source, now)
			initial.WireVersion = version
			initial.ReputationLevel, initial.AgeMS, initial.RouteKnowledgeBalance = 11, 12_345, 9
			if version >= 17 {
				if err := activateMinigameState(initial, source.Minigames); err != nil {
					t.Fatal(err)
				}
			}
			if version >= 18 {
				initial.Pets = map[string]pet.CareState{}
			}
			if version >= 19 {
				initial.FiscalPeriodOpenedWallMS = now.UnixMilli()
				initial.FiscalGeneratorLevels, initial.FiscalUnlocks = map[string]int64{}, map[string]bool{}
				for _, row := range source.Fiscal.GeneratorLevelRows() {
					initial.FiscalGeneratorLevels[row.GeneratorID] = 0
				}
			}
			if version >= 20 {
				initial.Soul, initial.SoulExhaustedSourceIDs = source.Soul.Policy.Initial, []string{}
			}
			input := mustEncodeState(t, initial)
			restore := func() *save.State {
				state, err := save.RestoreState(input, version, source.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				if save.VersionForState(state) != version || state.ReputationSpent != 0 || state.ReputationNodesOwned != nil || state.ReputationLevel != 11 {
					t.Fatal("load activated or changed legacy accounting")
				}
				if err := source.ValidateFoundationState(state); err != nil {
					t.Fatal(err)
				}
				return state
			}
			live := restore()
			company := foundationScopeState(t, source.Economy, economy.ScopeCompany)
			company.WireVersion, company.RunStartedAt = companyVersion, now.Add(-time.Hour)
			if source.foundationsActive() {
				// Use the actual activation kernel to prepare the matching old
				// Company's meter/achievement state, without mutating the source Founder.
				legacy := sources[0]
				seedFounder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
				seedCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
				if err := settleAndActivateFoundations(legacy, source, seedFounder, seedCompany, company); err != nil {
					t.Fatal(err)
				}
			}
			newCompany := foundationScopeState(t, next.Economy, economy.ScopeCompany)
			newCompany.RunStartedAt = now
			if err := settleAndActivateFoundations(source, next, live, company, newCompany); err != nil {
				t.Fatalf("live boundary: %v", err)
			}
			exit := save.ExitRecord{RunID: 1, ExitType: "collapse", OccurredAt: now}
			live.ExitHistory = append(live.ExitHistory, exit)
			command := save.FounderReplayCommand{IntentID: "01986666-8d01-7000-8000-000000000001",
				FounderStreamID: "01986666-8c00-4000-8000-000000000001", FounderID: "01986666-8d00-7000-8000-000000000001",
				Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
			resolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied),
				CompanyStreamID: "01986666-8e00-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 1,
				ResultConstantsHash: next.ConstantsHash, AgeMSBefore: 12_345, AgeMSAfter: 12_345,
				AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{},
				ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.UnixMilli()}, ResultFounderWireVersion: 22}
			if version < 20 {
				band, ok := next.Soul.BandFor(next.Soul.Policy.Initial)
				if !ok {
					t.Fatal("invalid next Soul initial")
				}
				resolved.NextSoul = &nextSoulWire{SoulInitial: next.Soul.Policy.Initial, BandMember: string(band.Member)}
			}
			inputs, err := save.MarshalFounderReplayInputs(command, resolved)
			if err != nil {
				t.Fatal(err)
			}
			request, err := ParseIntent([]byte(`{"intent_id":"01986666-8d01-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
			if err != nil || request.InvalidDetail != "" {
				t.Fatalf("invalid Exit request: %v/%s", err, request.InvalidDetail)
			}
			replayed, err := ApplyFounderLogged(restore(), request.CanonicalPayload, source, inputs)
			if err != nil || replayed.Outcome != save.IntentApplied || replayed.ResultConstantsHash != next.ConstantsHash {
				t.Fatalf("Founder replay: outcome=%s err=%v", replayed.Outcome, err)
			}
			for arm, state := range map[string]*save.State{"live": live, "founder_replay": replayed.State} {
				if save.VersionForState(state) != 22 || state.ReputationLevel != 11 || state.ReputationSpent != 0 ||
					state.ReputationNodesOwned == nil || len(state.ReputationNodesOwned) != 0 || state.ReputationUnlockPPM != 0 ||
					state.AgeMS != 12_345 || state.RouteKnowledgeBalance != 9 {
					t.Fatalf("%s accounting/metadata changed: %+v", arm, state)
				}
				available, err := reputation.Available(state.ReputationLevel, state.ReputationSpent)
				if err != nil || available != 11 {
					t.Fatalf("%s earned Reputation not fully spendable: %d/%v", arm, available, err)
				}
				encoded := mustEncodeState(t, state)
				decoded, err := save.RestoreState(encoded, 22, next.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatalf("%s full v22 codec: %v", arm, err)
				}
				if err := next.ValidateFoundationState(decoded); err != nil {
					t.Fatalf("%s pinned admission: %v", arm, err)
				}
			}
			if !bytes.Equal(mustEncodeState(t, live), mustEncodeState(t, replayed.State)) {
				t.Fatal("live boundary and Founder replay differ in complete encoded state")
			}
			sourceID, nextID := fmt.Sprintf("source-v%d", version), "target-v22-legacy"
			if version == 21 {
				nextID = "target-v22-epoch8"
			}
			corpus.Bundles[sourceID] = reputationEarlierCorpusBundle{source.ConstantsHash, artifactStrings(source.Artifacts)}
			corpus.Bundles[nextID] = reputationEarlierCorpusBundle{next.ConstantsHash, artifactStrings(next.Artifacts)}
			post, events := mustEncodeState(t, replayed.State), fixtureEvents(replayed.Events)
			corpus.Cases = append(corpus.Cases, reputationEarlierCorpusCase{sourceID, nextID, crossRuntimeFounderCase{
				Name: fmt.Sprintf("founder-v%d-to-v22", version), StateVersion: version, PreState: input,
				CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Outcome: string(replayed.Outcome),
				Receipt: replayed.Receipt, Events: events, PostState: post, ResultConstantsHash: replayed.ResultConstantsHash,
				ReceiptJSON: canonicalFixtureJSON(t, replayed.Receipt), EventsJSON: canonicalFixtureValue(t, events),
				PostStateJSON: canonicalFixtureJSON(t, post),
			}})
		})
	}
	if t.Failed() {
		return // A failing transition must never rewrite its own expectation.
	}
	if len(corpus.Cases) != 7 || len(corpus.Bundles) != 9 {
		t.Fatal("incomplete shared activation corpus")
	}
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	path := "../../testdata/replay/reputation-earlier-activation-v1.json"
	if *updateReplayFixture {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(committed, encoded) {
		t.Fatal("shared earlier-Founder activation corpus differs from executed Go transitions")
	}
}
