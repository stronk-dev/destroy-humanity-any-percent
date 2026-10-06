package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const reputationCompanyRunPath = "../../testdata/reputation/company-run-v1.json"

var updateReputationCompanyRun = flag.Bool("update-reputation-company-run-fixture", false, "generate the R8 Company-run test fixture")

type reputationCompanyRunProfile struct {
	Name           string                     `json:"name"`
	BonusFactor    string                     `json:"bonus_factor"`
	Entries        []crossRuntimeFullRunEntry `json:"entries"`
	FinalStateJSON string                     `json:"final_state_json"`
}

type reputationCompanyRunFixture struct {
	Version        int                           `json:"schema_version"`
	SourceSHA256   string                        `json:"source_sha256"`
	ConstantsHash  string                        `json:"constants_hash"`
	CompanyVersion int                           `json:"company_version"`
	Genesis        json.RawMessage               `json:"genesis"`
	Profiles       []reputationCompanyRunProfile `json:"profiles"`
}

func reputationCompanyRunSource(t *testing.T) (reputationCorpus, []byte, CatalogBundle) {
	t.Helper()
	source, raw := reputationInputShapeSource(t)
	if err := validateReputationHistorySource(source, raw); err != nil {
		t.Fatal(err)
	}
	bundle := reputationShapeCatalogs(t, source)["tree"]
	if source.Exit.ConstantsHash != bundle.ConstantsHash || source.Exit.NextConstantsHash != bundle.ConstantsHash ||
		canonicalFixtureValue(t, source.Exit.Artifacts) != canonicalFixtureValue(t, stringArtifacts(bundle.Artifacts)) ||
		canonicalFixtureValue(t, source.Exit.NextArtifacts) != canonicalFixtureValue(t, stringArtifacts(bundle.Artifacts)) {
		t.Fatal("Company-run source bundle is not the complete pinned tree bundle")
	}
	return source, raw, bundle
}

func buildReputationCompanyRun(t *testing.T) reputationCompanyRunFixture {
	t.Helper()
	source, raw, bundle := reputationCompanyRunSource(t)
	row := source.Exit.Case
	company, err := save.RestoreState(row.PreState, 18, bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	setup, err := ApplyLoggedExit(company, []byte(canonicalFixtureJSON(t, row.CanonicalPayload)), bundle, row.ReplayInputs)
	if err != nil || setup.Decision.Outcome != save.IntentApplied {
		t.Fatalf("source Exit setup: %v", err)
	}
	events := append(fixtureEvents(setup.Decision.FounderEvents), fixtureEvents(setup.Decision.CompanyEndedEvents)...)
	events = append(events, fixtureEvents(setup.Decision.CompanyStartedEvents)...)
	wantEvents := append(slices.Clone(row.FounderEvents), row.CompanyEndedEvents...)
	wantEvents = append(wantEvents, row.CompanyStartedEvents...)
	if !canonicalJSONEqual(setup.Decision.Receipt, []byte(row.ReceiptJSON)) ||
		canonicalFixtureValue(t, events) != canonicalFixtureValue(t, wantEvents) ||
		!canonicalJSONEqual(mustEncodeState(t, setup.Company), []byte(row.FinalCompanyJSON)) ||
		!canonicalJSONEqual(mustEncodeState(t, setup.Decision.NewCompanyState), []byte(row.NewCompanyJSON)) {
		t.Fatal("source Exit setup differs from immutable expectations")
	}
	genesis := json.RawMessage(row.NewCompanyJSON)
	fixture := reputationCompanyRunFixture{Version: 1, SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(raw)),
		ConstantsHash: bundle.ConstantsHash, CompanyVersion: 18, Genesis: genesis}
	wire, err := parseReplayInputs(row.ReplayInputs)
	if err != nil {
		t.Fatal(err)
	}
	carry := founderCarry(setup.Founder)
	var old replayExitResolved
	if err := decodeReplayStrict(wire.Resolved, &old); err != nil {
		t.Fatal(err)
	}
	carry.FounderRevision, carry.FounderConstantsHash = old.FounderCarry.FounderRevision+1, bundle.ConstantsHash
	for _, profile := range []struct{ name, factor string }{{"source-bonus", "1.003e0"}, {"unit-control", "1e0"}} {
		state, err := save.RestoreState(genesis, 18, bundle.Economy, economy.ScopeCompany, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		cash, _ := state.Ledger.Balance("company.cash")
		if state.RunSeq != 2 || state.Tier != 0 || cash.String() != "1e3" || state.GeneratorProvisioned["generator.beige_tower"] != 15 || state.GeneratorCounts["generator.beige_tower"] != 0 {
			t.Fatal("source new-run genesis does not contain the ruled starters")
		}
		frozen := slices.Clone(setup.Decision.NewRunFrozenContributions)
		bonus := slices.IndexFunc(frozen, func(value save.FrozenContribution) bool { return value.SourceID == "reputation.founder_bonus" })
		if bonus < 0 || frozen[bonus].Factor != "1.003e0" {
			t.Fatal("source new run does not freeze the declared non-unit bonus")
		}
		frozen[bonus].Factor = profile.factor // Explicit counterfactual only in the unit control.
		contributions, err := ResolveFrozenContributions(bundle.Economy, frozen)
		if err != nil {
			t.Fatal(err)
		}
		result := reputationCompanyRunProfile{Name: profile.name, BonusFactor: profile.factor}
		base := time.UnixMilli(state.RunStartedAt.UnixMilli()).UTC()
		bodies := []string{`"kind":"perform_manual_batch","action_id":"manual.click","count":1,"window_ms":1`,
			`"kind":"cross_gate","gate_id":"gate.t0_to_t1","route_id":null`, fmt.Sprintf(`"kind":"wind_down","expected_founder_revision":%d`, carry.FounderRevision)}
		for index, body := range bodies {
			now := base.Add(time.Duration(7000+index) * time.Second)
			request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":"01986666-6d%02d-7000-8000-000000000001","expected_revision":%d,%s}`, index+10, index+1, body)))
			if err != nil || request.InvalidDetail != "" {
				t.Fatalf("run command: %v detail=%s", err, request.InvalidDetail)
			}
			active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, wire.Command.FounderID, now)
			if err != nil {
				t.Fatal(err)
			}
			build := replayBuild{Command: save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: wire.Command.CompanyStreamID,
				FounderID: wire.Command.FounderID, Revision: int64(index + 1), RunSeq: 2, RunLogSeq: int64(index + 1)},
				Mode: ModeOnline, Now: now, IntentKind: request.Kind, Contributions: contributions,
				RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &active}
			terminal := index == 2
			if terminal {
				spawn, err := bundle.Opportunities.Spawn(wire.Command.FounderID, 3, 0, 0)
				if err != nil {
					t.Fatal(err)
				}
				build.Terminal, build.SelectedExitType, build.SelectedTerms = true, "collapse", json.RawMessage(`{}`)
				build.NextConstantsHash, build.NextActivePlay = bundle.ConstantsHash, spawnEvidence(spawn)
				build.MinigameSessionActive = new(bool)
			}
			inputs, err := buildReplayInputs(build)
			if err != nil {
				t.Fatal(err)
			}
			entry := crossRuntimeFullRunEntry{Seq: int64(index + 1), CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Terminal: terminal}
			if terminal {
				transition, err := ApplyLoggedExit(state, request.CanonicalPayload, bundle, inputs)
				if err != nil || transition.Decision.Outcome != save.IntentApplied {
					t.Fatalf("terminal run command: %v outcome=%s", err, transition.Decision.Outcome)
				}
				all := append(slices.Clone(transition.Decision.FounderEvents), transition.Decision.CompanyEndedEvents...)
				all = append(all, transition.Decision.CompanyStartedEvents...)
				entry.ReceiptJSON, entry.EventsJSON = canonicalFixtureJSON(t, transition.Decision.Receipt), string(marshalReplayEvents(all))
				state = transition.Company
			} else {
				transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
				if err != nil || transition.Outcome != save.IntentApplied {
					t.Fatalf("ordinary run command %d: %v outcome=%s", index+1, err, transition.Outcome)
				}
				entry.ReceiptJSON, entry.EventsJSON = canonicalFixtureJSON(t, transition.Receipt), string(marshalReplayEvents(transition.Events))
				state = transition.State
				if index == 0 {
					factor, _ := decimal.ParseCanonical(profile.factor)
					// The all-target prestige row multiplies generator production;
					// this manual action matches only explicit manual.click targets.
					want := decimal.FromFloat64(15).Mul(factor).Mul(decimal.FromFloat64(7000)).Add(decimal.FromFloat64(1000)).Add(decimal.One)
					got, _ := state.Ledger.Balance("company.cash")
					if !got.Eq(want) {
						t.Fatalf("independent starter/bonus arithmetic: cash=%s want=%s", got, want)
					}
				} else if state.Tier != 1 || !state.GatesCrossed["gate.t0_to_t1"] {
					t.Fatal("ordinary gate command did not cross the gate")
				}
			}
			result.Entries = append(result.Entries, entry)
		}
		result.FinalStateJSON = canonicalFixtureJSON(t, mustEncodeState(t, state))
		fixture.Profiles = append(fixture.Profiles, result)
	}
	return fixture
}

func validateReputationCompanyRun(fixture reputationCompanyRunFixture, source reputationCorpus, raw []byte) error {
	if fixture.Version != 1 || fixture.SourceSHA256 != fmt.Sprintf("%x", sha256.Sum256(raw)) || fixture.ConstantsHash != source.Exit.NextConstantsHash ||
		fixture.CompanyVersion != 18 || !canonicalJSONEqual(fixture.Genesis, []byte(source.Exit.Case.NewCompanyJSON)) || len(fixture.Profiles) != 2 {
		return fmt.Errorf("invalid source-bound Company-run population")
	}
	for index, profile := range fixture.Profiles {
		if profile.Name != []string{"source-bonus", "unit-control"}[index] || profile.BonusFactor != []string{"1.003e0", "1e0"}[index] ||
			len(profile.Entries) != 3 || !profile.Entries[2].Terminal || !json.Valid([]byte(profile.FinalStateJSON)) {
			return fmt.Errorf("invalid Company-run profile %d", index)
		}
	}
	return nil
}

func reputationCompanyRun(t *testing.T) (reputationCompanyRunFixture, CatalogBundle) {
	t.Helper()
	source, raw, bundle := reputationCompanyRunSource(t)
	data, err := os.ReadFile(reputationCompanyRunPath)
	if err != nil {
		t.Fatal(err)
	}
	var fixture reputationCompanyRunFixture
	if err := decodeReplayStrict(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if err := validateReputationCompanyRun(fixture, source, raw); err != nil {
		t.Fatal(err)
	}
	return fixture, bundle
}

func TestReputationCompanyRunFixture(t *testing.T) {
	fixture := buildReputationCompanyRun(t)
	data, err := json.MarshalIndent(fixture, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if *updateReputationCompanyRun {
		if err := os.WriteFile(reputationCompanyRunPath, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	committed, err := os.ReadFile(reputationCompanyRunPath)
	if err != nil || !bytes.Equal(data, committed) {
		t.Fatalf("retained Company-run fixture differs from source-derived generation: %v", err)
	}
}

func reputationCompanyEntries(t *testing.T, profile reputationCompanyRunProfile) []ReplayLogEntry {
	t.Helper()
	entries := make([]ReplayLogEntry, len(profile.Entries))
	for index, row := range profile.Entries {
		// JSON fixture presentation adds indentation to raw objects; the
		// immutable command semantics remain the source of canonical bytes.
		entries[index] = ReplayLogEntry{Sequence: row.Seq, CanonicalPayload: []byte(canonicalFixtureJSON(t, row.CanonicalPayload)), ReplayInputs: row.ReplayInputs,
			ReceiptJSON: []byte(row.ReceiptJSON), EventsJSON: []byte(row.EventsJSON), Terminal: row.Terminal}
	}
	return entries
}

func TestReputationCompanyRunVerifies(t *testing.T) {
	fixture, bundle := reputationCompanyRun(t)
	for _, profile := range fixture.Profiles {
		t.Run(profile.Name, func(t *testing.T) {
			entries := reputationCompanyEntries(t, profile)
			if verdict := VerifyReplayRun(fixture.Genesis, fixture.CompanyVersion, bundle, entries, fixture.ConstantsHash, false); verdict != ReplayVerified {
				t.Fatalf("public Company-run verdict=%s", verdict)
			}
			if !reputationCompanyFinalMatches(t, fixture, bundle, profile) {
				t.Fatal("terminal full Company state differs")
			}
		})
	}
}

func reputationCompanyFinalMatches(t *testing.T, fixture reputationCompanyRunFixture, bundle CatalogBundle, profile reputationCompanyRunProfile) bool {
	t.Helper()
	verdict, state := verifyReplayRunDetailed(fixture.Genesis, fixture.CompanyVersion, bundle, reputationCompanyEntries(t, profile), fixture.ConstantsHash, false)
	return verdict == ReplayVerified && state != nil && canonicalJSONEqual(mustEncodeState(t, state), []byte(profile.FinalStateJSON))
}

func TestReputationCompanyRunRefusesFalseHead(t *testing.T) {
	fixture, bundle := reputationCompanyRun(t)
	for _, profile := range fixture.Profiles {
		t.Run(profile.Name, func(t *testing.T) {
			copy := profile
			head := reputationShapeObject(t, []byte(copy.FinalStateJSON))
			balances := reputationShapeObject(t, head["balances"])
			if string(balances["company.cash"]) == `"0"` {
				t.Fatal("false-head control has no cash target")
			}
			balances["company.cash"] = json.RawMessage(`"0"`)
			head["balances"] = reputationShapeJSON(t, balances)
			copy.FinalStateJSON = string(reputationShapeJSON(t, head))
			if reputationCompanyFinalMatches(t, fixture, bundle, copy) {
				t.Fatal("forged expected Company head admitted")
			}
		})
	}
}

func TestReputationCompanyRunRefusesCorruption(t *testing.T) {
	fixture, bundle := reputationCompanyRun(t)
	for _, profile := range fixture.Profiles {
		for _, mutation := range []string{"starter-cash", "starter-units", "first-factor", "terminal-factor", "missing-terminal", "log-gap", "wrong-pin", "receipt", "event"} {
			t.Run(profile.Name+"/"+mutation, func(t *testing.T) {
				var copy reputationCompanyRunProfile
				if err := json.Unmarshal(reputationShapeJSON(t, profile), &copy); err != nil {
					t.Fatal(err)
				}
				genesis, pin, want := bytes.Clone(fixture.Genesis), fixture.ConstantsHash, ReplayStateDivergence
				switch mutation {
				case "starter-cash", "starter-units":
					object := reputationShapeObject(t, genesis)
					field, key, value := "balances", "company.cash", "0"
					if mutation == "starter-units" {
						field, key, value = "generators_provisioned", "generator.beige_tower", "10"
					}
					nested := reputationShapeObject(t, object[field])
					if mutation == "starter-cash" {
						nested[key] = reputationShapeJSON(t, value)
					} else {
						nested[key] = json.RawMessage(value)
					}
					object[field], genesis = reputationShapeJSON(t, nested), nil
					genesis = reputationShapeJSON(t, object)
				case "first-factor", "terminal-factor":
					index := 0
					if mutation == "terminal-factor" {
						index = 2
					}
					wire := reputationShapeObject(t, copy.Entries[index].ReplayInputs)
					resolved := reputationShapeObject(t, wire["resolved"])
					accrual := reputationShapeObject(t, resolved["accrual"])
					var values []replayContribution
					if err := json.Unmarshal(accrual["contributions"], &values); err != nil {
						t.Fatal(err)
					}
					bonus := slices.IndexFunc(values, func(row replayContribution) bool { return row.SourceID == "reputation.founder_bonus" })
					if bonus < 0 || values[bonus].Factor != profile.BonusFactor {
						t.Fatal("factor mutation lacks its target")
					}
					values[bonus].Factor = "1.004e0"
					accrual["contributions"] = reputationShapeJSON(t, values)
					resolved["accrual"] = reputationShapeJSON(t, accrual)
					wire["resolved"] = reputationShapeJSON(t, resolved)
					copy.Entries[index].ReplayInputs = reputationShapeJSON(t, wire)
				case "missing-terminal":
					copy.Entries, want = copy.Entries[:2], ReplayLogGap
				case "log-gap":
					copy.Entries[1].Seq, want = 4, ReplayLogGap
				case "wrong-pin":
					pin, want = "sha256:"+strings.Repeat("0", 64), ReplayConstantsMismatch
				case "receipt":
					object := reputationShapeObject(t, []byte(copy.Entries[0].ReceiptJSON))
					object["outcome"] = json.RawMessage(`"rejected"`)
					copy.Entries[0].ReceiptJSON = string(reputationShapeJSON(t, object))
				case "event":
					var events []fixtureEvent
					if err := json.Unmarshal([]byte(copy.Entries[0].EventsJSON), &events); err != nil {
						t.Fatal(err)
					}
					events = append(events, fixtureEvent{Kind: "gate_crossed", SchemaVersion: 1, Payload: json.RawMessage(`{}`)})
					copy.Entries[0].EventsJSON = string(reputationShapeJSON(t, events))
				}
				if verdict := VerifyReplayRun(genesis, fixture.CompanyVersion, bundle, reputationCompanyEntries(t, copy), pin, false); verdict != want {
					t.Fatalf("corrupt Company-run verdict=%s want=%s", verdict, want)
				}
			})
		}
	}
}

func TestReputationCompanyRunRefusesMissingPopulation(t *testing.T) {
	fixture, _ := reputationCompanyRun(t)
	source, raw := reputationInputShapeSource(t)
	for _, mutation := range []string{"source-sha", "missing-profile", "missing-command", "genesis"} {
		t.Run(mutation, func(t *testing.T) {
			copy := fixture
			copy.Profiles = slices.Clone(fixture.Profiles)
			switch mutation {
			case "source-sha":
				copy.SourceSHA256 = strings.Repeat("0", 64)
			case "missing-profile":
				copy.Profiles = copy.Profiles[:1]
			case "missing-command":
				copy.Profiles[0].Entries = copy.Profiles[0].Entries[:2]
			case "genesis":
				copy.Genesis = json.RawMessage(`{}`)
			}
			if err := validateReputationCompanyRun(copy, source, raw); err == nil {
				t.Fatal("corrupt Company-run population admitted")
			}
		})
	}
}
