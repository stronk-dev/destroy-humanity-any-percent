package garden

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGardenIndependentCommandWitnesses(t *testing.T) {
	var population struct {
		Version int `json:"version"`
		Cases   []struct {
			Name    string              `json:"name"`
			Anchor  *int64              `json:"anchor"`
			Stamp   *int64              `json:"stamp"`
			Initial [][]json.RawMessage `json:"initial"`
			Step    corpusStep          `json:"step"`
			Error   *string             `json:"error"`
			Result  json.RawMessage     `json:"result"`
			Post    [][]json.RawMessage `json:"post"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/garden/command-witnesses-v1.json"), &population); err != nil || population.Version != 1 || len(population.Cases) != 15 {
		t.Fatalf("invalid command population: %v", err)
	}
	catalog := fixtureCatalog(t)
	seen := map[string]bool{}
	for _, test := range population.Cases {
		if seen[test.Name] || test.Name == "" {
			t.Fatal("duplicate/empty command case")
		}
		seen[test.Name] = true
		t.Run(test.Name, func(t *testing.T) {
			state := NewState(catalog)
			salt := "0123456789abcdef"
			state.SaltHex, state.TickAnchorWallMS, state.TickSeq = &salt, newInt64(1_000_000), 17
			if test.Anchor != nil {
				state.TickAnchorWallMS = cloneInt(test.Anchor)
			}
			state.SubstrateSetWallMS, state.Plots = cloneInt(test.Stamp), witnessPlots(t, test.Initial)
			initial, err := EncodeState(state)
			if err != nil {
				t.Fatal(err)
			}
			state, err = DecodeState(initial)
			if err != nil || ValidateAgainst(catalog, state) != nil {
				t.Fatalf("invalid admitted input: %v", err)
			}
			want := state.Clone()
			want.Plots = witnessPlots(t, test.Post)
			if test.Error == nil && test.Step.Op == "set_substrate" {
				want.SubstrateID = test.Step.SubstrateID
				want.TickAnchorWallMS, want.SubstrateSetWallMS = newInt64(test.Step.ServerMS), newInt64(test.Step.ServerMS)
			}
			// Direct call: no clone-and-discard wrapper can conceal refusal mutations.
			result, err := runStep(catalog, state, test.Step)
			if test.Error != nil {
				rejection, ok := AsRejection(err)
				if !ok || rejection.Category+"/"+rejection.Detail != *test.Error {
					t.Fatalf("rejection=%v want=%s", err, *test.Error)
				}
				post, _ := EncodeState(state)
				if string(post) != string(initial) {
					t.Fatalf("refusal mutated state: %s -> %s", initial, post)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				bytes, marshalErr := json.Marshal(result)
				if marshalErr != nil || string(bytes) != string(test.Result) {
					t.Fatalf("result=%s want=%s error=%v", bytes, test.Result, marshalErr)
				}
			}
			gotBytes, gotErr := EncodeState(state)
			wantBytes, wantErr := EncodeState(want)
			if gotErr != nil || wantErr != nil || ValidateAgainst(catalog, state) != nil || string(gotBytes) != string(wantBytes) {
				t.Fatalf("post=%s want=%s errors=%v/%v", gotBytes, wantBytes, gotErr, wantErr)
			}
		})
	}
}

func TestGardenGatePrecedence(t *testing.T) {
	count := 0
	for _, mode := range []string{"unrelated", "human_hobby"} {
		catalog := scenarioCatalog(t, []fixtureOp{setOp(mode, "soul_gate")})
		for _, unlocked := range []bool{false, true} {
			for _, humanLocked := range []bool{false, true} {
				count++
				t.Run(fmt.Sprintf("%s/unlocked_%t/humanLocked_%t", mode, unlocked, humanLocked), func(t *testing.T) {
					want := ""
					if !unlocked {
						want = "not_eligible/fiscal_unlock_required"
					} else if mode == "human_hobby" && humanLocked {
						want = "not_eligible/human_content_locked"
					}
					err := catalog.Gate(unlocked, humanLocked)
					if want == "" {
						if err != nil {
							t.Fatal(err)
						}
						return
					}
					rejection, ok := AsRejection(err)
					if !ok || rejection.Category+"/"+rejection.Detail != want {
						t.Fatalf("rejection=%v want=%s", err, want)
					}
				})
			}
		}
	}
	if count != 8 {
		t.Fatalf("gate population=%d", count)
	}
}
