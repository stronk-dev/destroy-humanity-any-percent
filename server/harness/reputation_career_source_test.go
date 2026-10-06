package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

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
