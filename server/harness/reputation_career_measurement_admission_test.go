package harness

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func careerMeasurementEncode(t *testing.T, value any) []byte {
	t.Helper()
	data, err := CanonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Complete declaration-shaped SYNTHETIC data exercises admission, not earnings.
func syntheticCareerMeasurement(t *testing.T, inputs reputationCareerMeasurementInputs) (reputationCareerReport, reputationRelevanceReport) {
	t.Helper()
	clock := func(value int64) *int64 { return &value }
	var rows []reputationCareerSeed
	var arms []reputationRelevanceOutcome
	for index, pair := range inputs.pairs {
		purchased := []string{"reputation.unlock.p05", "reputation.starter.cash_small"}
		rows = append(rows, reputationCareerSeed{TreatedSource: pair.treated, ControlSource: pair.control,
			PolicyID: pair.treated.RunKey.PolicyID, Seed: pair.seed, CareerPolicy: string(pair.treated.PurchasePolicy),
			PurchasedNodeIDs: purchased, AppliedStarterIDs: []string{"reputation.starter.cash_small"}, BonusFactor: "1e0",
			TreatedGateMS: clock(10), ControlGateMS: clock(20), SavedMS: clock(10)})
		for _, source := range inputs.arms[index*(len(inputs.nodes)+1) : (index+1)*(len(inputs.nodes)+1)] {
			gate, bought := clock(10), purchased
			switch source.ExcludedNodeID {
			case "reputation.unlock.p05":
				bought = nil
			case "reputation.starter.cash_small":
				gate, bought = clock(20), []string{"reputation.unlock.p05"}
			}
			arms = append(arms, reputationRelevanceOutcome{source: source, gate: gate, purchased: bought})
		}
	}
	h4, err := newReputationCareerReport(rows)
	if err != nil {
		t.Fatal(err)
	}
	h5, err := composeReputationRelevanceReport(inputs.nodes, arms)
	if err != nil {
		t.Fatal(err)
	}
	return h4, h5
}

func TestReputationCareerMeasurementAdmission(t *testing.T) {
	inputs := declaredCareerMeasurementInputs(t)
	h4Report, h5Report := syntheticCareerMeasurement(t, inputs)
	h4, h5 := careerMeasurementEncode(t, h4Report), careerMeasurementEncode(t, h5Report)
	producer := reputationMeasurementProducer{Commit: strings.Repeat("1", 40), ServerTree: strings.Repeat("2", 40), BalanceTree: strings.Repeat("3", 40), KernelVersion: "0.3.161"}
	lineage := reputationCareerLineage{SchemaVersion: 1, Producer: producer, GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH,
		H4SHA256: reputationMeasurementSHA(h4), H5SHA256: reputationMeasurementSHA(h5), Pairs: len(inputs.pairs), H4Sources: len(inputs.pairs) * 2, H5Arms: len(inputs.arms)}
	if err := validateCareerMeasurement(inputs, h4, h5, lineage, producer); err != nil {
		t.Fatalf("healthy complete synthetic admission: %v", err)
	}
	for name, mutate := range map[string]func(*reputationCareerLineage){
		"schema":       func(l *reputationCareerLineage) { l.SchemaVersion++ },
		"commit":       func(l *reputationCareerLineage) { l.Producer.Commit = "invented" },
		"server":       func(l *reputationCareerLineage) { l.Producer.ServerTree = "invented" },
		"balance":      func(l *reputationCareerLineage) { l.Producer.BalanceTree = "invented" },
		"kernel":       func(l *reputationCareerLineage) { l.Producer.KernelVersion = "0.3.160" },
		"runtime":      func(l *reputationCareerLineage) { l.GoVersion = "" },
		"os":           func(l *reputationCareerLineage) { l.GOOS = "" },
		"architecture": func(l *reputationCareerLineage) { l.GOARCH = "" },
		"h4-hash":      func(l *reputationCareerLineage) { l.H4SHA256 = "invented" },
		"h5-hash":      func(l *reputationCareerLineage) { l.H5SHA256 = "invented" },
		"pairs":        func(l *reputationCareerLineage) { l.Pairs-- },
		"h4-sources":   func(l *reputationCareerLineage) { l.H4Sources-- },
		"h5-arms":      func(l *reputationCareerLineage) { l.H5Arms-- },
	} {
		t.Run("lineage/"+name, func(t *testing.T) {
			copy := lineage
			mutate(&copy)
			if err := validateCareerMeasurement(inputs, h4, h5, copy, producer); err == nil {
				t.Fatal("corrupt lineage admitted")
			}
		})
	}
	mutations := map[string]func(*reputationCareerReport, *reputationRelevanceReport){
		"coherent-cross-gate": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			value := int64(11)
			b.Arms[0].Gate = &value
		},
		"coherent-cross-purchases": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms[0].Purchased = nil },
		"coherent-h5-horizon":      func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms[1].Source.HorizonMS++ },
		"coherent-h5-reordered": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1], b.Arms[2] = b.Arms[2], b.Arms[1]
		},
		"coherent-h4-control": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].ControlSource.ExcludedNodeID = "reputation.starter.cash_small"
		},
		"coherent-whole-cohort": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].Seed = 9999
			a.Seeds[0].TreatedSource.RunKey.Seed, a.Seeds[0].ControlSource.RunKey.Seed = "9999", "9999"
			for index := 0; index <= len(inputs.nodes); index++ {
				b.Arms[index].Source.RunKey.Seed = "9999"
			}
		},
		"h4-missing":   func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Seeds = a.Seeds[1:] },
		"h4-duplicate": func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Seeds[1] = a.Seeds[0] },
		"h4-reordered": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0], a.Seeds[1] = a.Seeds[1], a.Seeds[0]
		},
		"h4-treated-source": func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Seeds[0].TreatedSource.HorizonMS++ },
		"h4-control-source": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].ControlSource.PurchasePolicy = CareerCheapest
		},
		"h4-seed": func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Seeds[0].Seed++ },
		"h4-policy-label": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].CareerPolicy = string(CareerNone)
		},
		"h4-unpurchased-starter": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].AppliedStarterIDs = []string{"reputation.starter.cash_large"}
		},
		"h4-duplicate-starter": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].AppliedStarterIDs = append(a.Seeds[0].AppliedStarterIDs, a.Seeds[0].AppliedStarterIDs[0])
		},
		"h4-bonus-as-starter": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.Seeds[0].AppliedStarterIDs = []string{"reputation.unlock.p05"}
		},
		"h4-gate-status": func(a *reputationCareerReport, b *reputationRelevanceReport) { a.GatePassed = !a.GatePassed },
		"h4-violations":  func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Violations = []string{"invented"} },
		"h4-census":      func(a *reputationCareerReport, b *reputationRelevanceReport) { a.GatedSeeds-- },
		"h4-median": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.SavedMS["chaos.t0_t1"] = [3]int64{1, 2, 3}
		},
		"h4-finite-median": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			a.FiniteSavedMS["chaos.t0_t1"] = [3]int64{1, 2, 3}
		},
		"h4-header":    func(a *reputationCareerReport, b *reputationRelevanceReport) { a.Threshold = "1e12" },
		"h5-missing":   func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms = b.Arms[1:] },
		"h5-duplicate": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms[2] = b.Arms[1] },
		"h5-reordered": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1], b.Arms[2] = b.Arms[2], b.Arms[1]
		},
		"h5-horizon": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms[1].Source.HorizonMS++ },
		"h5-prestige": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1].Source.EffectivePrestigePolicyHash = "invented"
		},
		"h5-fixture": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1].Source.RunKey.ConstantsHash = "invented"
		},
		"h5-experiment": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1].Source.Experiment.RouteKnowledgeBonus++
		},
		"h5-policy": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1].Source.FirstHourPolicyHash = "invented"
		},
		"h5-mask": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Arms[1].Source.ExcludedNodeID = "invented"
		},
		"h5-median": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Nodes[1].DeltaMSP50["chaos.t0_t1"]++ },
		"h5-census": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			b.Nodes[1].PurchasedRuns["chaos.t0_t1"]++
		},
		"h5-sources": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Sources = b.Sources[1:] },
		"cross-gate": func(a *reputationCareerReport, b *reputationRelevanceReport) {
			value := int64(11)
			b.Arms[0].Gate = &value
		},
		"cross-purchases": func(a *reputationCareerReport, b *reputationRelevanceReport) { b.Arms[0].Purchased = nil },
	}
	for name, mutate := range mutations {
		t.Run("reports/"+name, func(t *testing.T) {
			var a reputationCareerReport
			var b reputationRelevanceReport
			if err := reputationMeasurementJSON(h4, &a); err != nil {
				t.Fatal(err)
			}
			if err := reputationMeasurementJSON(h5, &b); err != nil {
				t.Fatal(err)
			}
			mutate(&a, &b)
			if strings.HasPrefix(name, "coherent-") {
				var err error
				a, err = newReputationCareerReport(a.Seeds)
				if err != nil {
					t.Fatal(err)
				}
				var outcomes []reputationRelevanceOutcome
				for _, arm := range b.Arms {
					outcomes = append(outcomes, reputationRelevanceOutcome{source: arm.Source, gate: arm.Gate, purchased: arm.Purchased})
				}
				// A horizon change must cover its baseline group to stay coherent.
				if name == "coherent-h5-horizon" {
					for index := 0; index <= len(inputs.nodes); index++ {
						outcomes[index].source.HorizonMS++
					}
					outcomes[1].source.HorizonMS--
				}
				b, err = composeReputationRelevanceReport(inputs.nodes, outcomes)
				if err != nil {
					t.Fatalf("coherent corruption failed fixture construction instead of admission: %v", err)
				}
				if err := validateReputationRelevanceRecomposition(inputs.nodes, b); err != nil {
					t.Fatalf("coherent corruption failed internal recomposition: %v", err)
				}
			}
			x, y := careerMeasurementEncode(t, a), careerMeasurementEncode(t, b)
			copy := lineage
			copy.H4SHA256, copy.H5SHA256 = reputationMeasurementSHA(x), reputationMeasurementSHA(y)
			if err := validateCareerMeasurement(inputs, x, y, copy, producer); err == nil {
				t.Fatal("corrupt reports admitted despite rebound raw hashes")
			}
		})
	}
	for _, name := range []string{"h4-unknown-field", "h5-trailing"} {
		t.Run(name, func(t *testing.T) {
			x, y := bytes.Clone(h4), bytes.Clone(h5)
			if name == "h4-unknown-field" {
				x = append([]byte(`{"unknown":true,`), x[1:]...)
			} else {
				y = append(y, []byte(` {}`)...)
			}
			copy := lineage
			copy.H4SHA256, copy.H5SHA256 = reputationMeasurementSHA(x), reputationMeasurementSHA(y)
			if err := validateCareerMeasurement(inputs, x, y, copy, producer); err == nil {
				t.Fatal("non-strict JSON admitted")
			}
		})
	}
	// A truthful negative H4 observation is valid research, never gate acceptance.
	a, b := syntheticCareerMeasurement(t, inputs)
	value := int64(20)
	a.Seeds[0].TreatedGateMS, b.Arms[0].Gate = &value, &value
	zero := int64(0)
	a.Seeds[0].SavedMS = &zero
	a, err := newReputationCareerReport(a.Seeds)
	if err != nil {
		t.Fatal(err)
	}
	var outcomes []reputationRelevanceOutcome
	for _, arm := range b.Arms {
		outcomes = append(outcomes, reputationRelevanceOutcome{source: arm.Source, gate: arm.Gate, purchased: arm.Purchased})
	}
	b, err = composeReputationRelevanceReport(inputs.nodes, outcomes)
	if err != nil {
		t.Fatal(err)
	}
	x, y := careerMeasurementEncode(t, a), careerMeasurementEncode(t, b)
	copy := lineage
	copy.H4SHA256, copy.H5SHA256 = reputationMeasurementSHA(x), reputationMeasurementSHA(y)
	if a.GatePassed || len(a.Violations) != 1 {
		t.Fatal("synthetic tie incorrectly passed H4")
	}
	if err := validateCareerMeasurement(inputs, x, y, copy, producer); err != nil {
		t.Fatalf("truthful negative measurement refused: %v", err)
	}
}

func TestReputationCareerMeasurementArtifacts(t *testing.T) {
	// Before the first recording commit only, absence is explicit, not a proof.
	absent := 0
	for _, path := range careerMeasurementPaths {
		if _, err := os.Lstat(filepath.Join(repositoryRootForReputation, path)); os.IsNotExist(err) {
			absent++
		}
	}
	if absent == len(careerMeasurementPaths) {
		if _, err := reputationMeasurementGit("cat-file", "-e", "HEAD:"+careerMeasurementPaths[2]); err != nil {
			t.Skip("dated career artifacts not yet recorded at HEAD; no retained or fresh evidence claimed")
		}
	}
	h4, h5, lineage := careerMeasurementArtifacts(t)
	producer, err := reputationMeasurementIdentity(lineage.Producer.Commit)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCareerMeasurement(declaredCareerMeasurementInputs(t), h4, h5, lineage, producer); err != nil {
		t.Fatal(err)
	}
	t.Logf("retained career validation only: producer=%s pairs=%d H5arms=%d; not fresh execution", producer.Commit, lineage.Pairs, lineage.H5Arms)
}

func TestReputationCareerMeasurementDeclaration(t *testing.T) {
	inputs := declaredCareerMeasurementInputs(t)
	suite, experiment := currentReputationInputs(t)
	bundle := reputationCareerBundle(t, suite)
	index := 0
	for _, spec := range suite.Scenario.Runs {
		// Pinned population starts at zero; an independently decoded JSON-side
		// descriptor checks every coordinate, not the producer's assertion.
		if spec.SeedStart != "0" {
			t.Fatal("predeclared cohort no longer starts at zero")
		}
		policy := CareerCheapest
		if spec.PolicyID == "chaos.t0_t1" {
			policy = CareerSeededUniform
		}
		for seed := 0; seed < spec.SeedCount; seed++ {
			for offset := -1; offset <= len(inputs.nodes); offset++ {
				config := ReputationCareerConfig{Bundle: bundle, Threshold: reputationCareerFixtureThreshold, Policy: policy}
				actual := inputs.pairs[index].treated
				if offset == -1 {
					config.Policy = CareerNone
					actual = inputs.pairs[index].control
				} else {
					actual = inputs.arms[index*(len(inputs.nodes)+1)+offset]
					if offset > 0 {
						config.Exclude = inputs.nodes[offset-1].NodeID
					}
				}
				wanted := expectedReputationCareerSource(t, suite, spec, uint64(seed), experiment, config)
				if !bytes.Equal(careerMeasurementEncode(t, actual), careerMeasurementEncode(t, wanted)) {
					t.Fatal("declared career source differs from independent expected coordinates")
				}
			}
			index++
		}
	}
	if index != 97 || len(inputs.arms) != 970 {
		t.Fatal("declaration lost part of the full population")
	}
}
