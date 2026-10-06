package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// Report admission reconstructs the declared experiment independently of the
// returned metadata; diagnostics also check the producer against known inputs.
func validateReputationCareerResultSource(suite *FirstHourSuite, spec RunSpec, seed uint64, experiment FirstHourExperiment,
	config ReputationCareerConfig, result ReputationCareerResult) error {
	if result.Run.Outcome != "completed" || len(result.Run.InvariantFailures) != 0 {
		return fmt.Errorf("career source admission requires a completed non-truncated run")
	}
	policy := *suite.Bundle.Prestige
	policy.Threshold = config.Threshold
	wanted, err := describeReputationCareerSource(suite, spec, seed, experiment, config, &policy)
	if err != nil {
		return err
	}
	if result.MeasurementSource != wanted || result.Run.Key != wanted.RunKey || result.Run.PolicyHash != wanted.FirstHourPolicyHash {
		return fmt.Errorf("career measurement source differs from declared experiment/run")
	}
	return nil
}

// JSON-side decoding deliberately compiles against the pre-source result, so
// the missing observation can be demonstrated without a compiler-error probe.
type reputationExpectedCareerSource struct {
	RunKey                      RunKey                 `json:"run_key"`
	FirstHourPolicyHash         string                 `json:"first_hour_policy_hash"`
	EffectivePrestigePolicyHash string                 `json:"effective_prestige_policy_hash"`
	Experiment                  FirstHourExperiment    `json:"experiment"`
	HorizonMS                   int64                  `json:"horizon_ms"`
	PurchasePolicy              ReputationCareerPolicy `json:"purchase_policy"`
	ExcludedNodeID              string                 `json:"excluded_node_id"`
}

func expectedReputationCareerSource(t *testing.T, suite *FirstHourSuite, spec RunSpec, seed uint64,
	experiment FirstHourExperiment, config ReputationCareerConfig) reputationExpectedCareerSource {
	t.Helper()
	policy := *suite.Bundle.Prestige
	policy.Threshold = config.Threshold
	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	key := suite.RunKey(spec, seed)
	key.ConstantsHash = config.Bundle.ConstantsHash
	return reputationExpectedCareerSource{RunKey: key, FirstHourPolicyHash: suite.PolicyHash,
		EffectivePrestigePolicyHash: "sha256:" + hex.EncodeToString(digest[:]), Experiment: experiment,
		HorizonMS: spec.HorizonMS, PurchasePolicy: config.Policy, ExcludedNodeID: config.Exclude}
}

func TestReputationCareerMeasurementSource(t *testing.T) {
	suite, spec, experiment, base := reputationCareerAdmissionInputs(t)
	var catalogKey *RunKey
	var sources []reputationExpectedCareerSource
	semanticPins := map[string]string{
		"base":           "fc02332820547f0bc437fa0c342e361673cf66be53363e9c3b97edacf6c5125c",
		"no-purchases":   "d0ce1b7ee0a102dc1a5258d775e04d9d3e707835a3f7c39601b9735d0fdd9d1f",
		"live-threshold": "225832823eab3e83fc9eb2595660e113c14645a51605aa66b14d442ea72bcd65",
		"excluded-node":  "8be070f54cb5ee12f1e32cafe6b83dca9d0856b7eaf6fc113f0fb6878b56cc7a",
	}
	for _, name := range []string{"base", "no-purchases", "live-threshold", "excluded-node"} {
		t.Run(name, func(t *testing.T) {
			config := base
			switch name {
			case "no-purchases":
				config.Policy = CareerNone
			case "live-threshold":
				config.Threshold = "1e12"
			case "excluded-node":
				config.Exclude = "reputation.starter.cash_small"
			}
			result, err := suite.RunReputationCareer(spec, 0, experiment, config)
			if err != nil || result.Run.Outcome != "completed" || len(result.Run.ReputationExits) != 2 {
				t.Fatalf("actual career control failed: %v / %+v", err, result.Run)
			}
			if catalogKey == nil {
				key := result.Run.Key
				catalogKey = &key
			} else if *catalogKey != result.Run.Key {
				t.Fatal("catalog RunKey should remain catalog authority, not smuggle experiment overrides")
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var observation struct {
				Source reputationExpectedCareerSource `json:"measurement_source"`
			}
			if err := json.Unmarshal(encoded, &observation); err != nil {
				t.Fatal(err)
			}
			var semantic map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &semantic); err != nil {
				t.Fatal(err)
			}
			delete(semantic, "measurement_source")
			semanticBytes, err := json.Marshal(semantic)
			if err != nil {
				t.Fatal(err)
			}
			semanticSHA := sha256.Sum256(semanticBytes)
			if hex.EncodeToString(semanticSHA[:]) != semanticPins[name] {
				t.Fatal("source metadata changed the complete observed career result")
			}
			var gate any
			if result.RunThreeGateMS != nil {
				gate = *result.RunThreeGateMS
			}
			t.Logf("actual %s: semanticSHA=%x gate=%v", name, semanticSHA, gate)
			t.Logf("actual %s: catalog=%s threshold=%s policy=%s exclude=%q exits=%+v purchases=%v gate=%v source=%+v", name,
				result.Run.Key.ConstantsHash, config.Threshold, config.Policy, config.Exclude, result.Run.ReputationExits,
				result.PurchasedNodeIDs, result.RunThreeGateMS, observation.Source)
			wanted := expectedReputationCareerSource(t, suite, spec, 0, experiment, config)
			if !reflect.DeepEqual(observation.Source, wanted) {
				t.Fatalf("actual career discarded effective measurement inputs: got %+v want %+v", observation.Source, wanted)
			}
			for _, earlier := range sources {
				if observation.Source == earlier {
					t.Fatal("different actual experiment inputs share one measurement identity")
				}
			}
			sources = append(sources, observation.Source)
		})
	}
}

func TestReputationCareerReportSourceAdmission(t *testing.T) {
	suite, spec, experiment, config := reputationCareerAdmissionInputs(t)
	makeResult := func(config ReputationCareerConfig) ReputationCareerResult {
		data, err := json.Marshal(expectedReputationCareerSource(t, suite, spec, 0, experiment, config))
		if err != nil {
			t.Fatal(err)
		}
		var source ReputationCareerMeasurementSource
		if err := json.Unmarshal(data, &source); err != nil {
			t.Fatal(err)
		}
		return ReputationCareerResult{MeasurementSource: source, Run: FirstHourRunResult{
			Key: source.RunKey, PolicyHash: source.FirstHourPolicyHash, Outcome: "completed"}}
	}
	treated := makeResult(config)
	controlConfig := config
	controlConfig.Policy = CareerNone
	control := makeResult(controlConfig)
	// These are synthetic projection controls, not earned-player populations.
	if row, err := projectReputationCareerPair(suite, spec, 0, experiment, config, treated, control); err != nil ||
		row.TreatedSource != treated.MeasurementSource || row.ControlSource != control.MeasurementSource {
		t.Fatalf("H4 failed to bind/retain both declared arm sources: %v", err)
	}
	var outcomes []reputationRelevanceOutcome
	for _, exclude := range []string{"", "reputation.starter.cash_small"} {
		masked := config
		masked.Exclude = exclude
		result := makeResult(masked)
		row, err := projectReputationRelevanceOutcome(suite, spec, 0, experiment, masked, result)
		if err != nil || row.source != result.MeasurementSource {
			t.Fatalf("H5 failed to bind/retain declared exclusion %q: %v", exclude, err)
		}
		outcomes = append(outcomes, row)
	}
	report := newReputationRelevanceReport(outcomes)
	if len(report.Sources) != len(outcomes) {
		t.Fatal("H5 report did not retain every arm's input source")
	}
	for index, outcome := range outcomes {
		if report.Sources[index] != outcome.source {
			t.Fatal("H5 report source differs from admitted arm's source/order")
		}
	}
	mutations := map[string]func(*ReputationCareerResult){
		"source-key":         func(r *ReputationCareerResult) { r.MeasurementSource.RunKey.ConstantsHash = "invented" },
		"first-hour-policy":  func(r *ReputationCareerResult) { r.MeasurementSource.FirstHourPolicyHash = "invented" },
		"effective-prestige": func(r *ReputationCareerResult) { r.MeasurementSource.EffectivePrestigePolicyHash = "invented" },
		"experiment":         func(r *ReputationCareerResult) { r.MeasurementSource.Experiment.RouteKnowledgeBonus++ },
		"horizon":            func(r *ReputationCareerResult) { r.MeasurementSource.HorizonMS++ },
		"purchase-policy":    func(r *ReputationCareerResult) { r.MeasurementSource.PurchasePolicy = CareerNone },
		"exclusion":          func(r *ReputationCareerResult) { r.MeasurementSource.ExcludedNodeID = "reputation.starter.cash_small" },
		"result-key":         func(r *ReputationCareerResult) { r.Run.Key.Seed = "9999" },
		"result-policy":      func(r *ReputationCareerResult) { r.Run.PolicyHash = "invented" },
		"failed":             func(r *ReputationCareerResult) { r.Run.Outcome = "failed" },
		"guard":              func(r *ReputationCareerResult) { r.Run.InvariantFailures = []string{"guard exhaustion"} },
		"missing-source":     func(r *ReputationCareerResult) { r.MeasurementSource = ReputationCareerMeasurementSource{} },
	}
	for name, mutate := range mutations {
		for _, consumer := range []string{"H4-treated", "H4-control", "H5"} {
			t.Run(consumer+"/"+name, func(t *testing.T) {
				copy := treated
				if consumer == "H4-control" {
					copy = control
				}
				mutate(&copy)
				// A distinct wrong policy is required in the no-purchases control.
				if consumer == "H4-control" && name == "purchase-policy" {
					copy.MeasurementSource.PurchasePolicy = CareerCheapest
				}
				var err error
				switch consumer {
				case "H4-treated":
					_, err = projectReputationCareerPair(suite, spec, 0, experiment, config, copy, control)
				case "H4-control":
					_, err = projectReputationCareerPair(suite, spec, 0, experiment, config, treated, copy)
				case "H5":
					_, err = projectReputationRelevanceOutcome(suite, spec, 0, experiment, config, copy)
				}
				if err == nil {
					t.Fatal("report consumer silently projected a corrupt experiment source")
				}
			})
		}
	}
}
