package garden

import (
	"encoding/json"
	"math/big"
	"testing"
)

// Independent integer reference, not an invocation of Advance or the corpus generator.
func referenceClock(catalog *Catalog, state *State, input AdvanceInput) (*State, Advance, string) {
	post := state.Clone()
	summary := Advance{TickSeqAfter: state.TickSeq, Matured: []PlotRef{}, Spawned: []Spawn{}}
	if !input.Unlocked {
		return post, summary, ""
	}
	if state.TickAnchorWallMS == nil {
		post.SaltHex = cloneString(&input.Salt)
		post.TickAnchorWallMS = cloneInt(&input.ServerMS)
		return post, summary, ""
	}
	if input.ServerMS <= *state.TickAnchorWallMS {
		return post, summary, ""
	}
	substrate, _ := catalog.Substrate(state.SubstrateID)
	elapsed := new(big.Int).Sub(big.NewInt(input.ServerMS), big.NewInt(*state.TickAnchorWallMS))
	cap := big.NewInt(catalog.CatchupCapMS)
	loss := new(big.Int)
	if elapsed.Cmp(cap) > 0 {
		loss.Sub(elapsed, cap)
		elapsed.Set(cap)
	}
	n := new(big.Int).Quo(elapsed, big.NewInt(substrate.TickMS))
	next := new(big.Int).Add(big.NewInt(state.TickSeq), n)
	if next.Cmp(big.NewInt(maxExactInteger)) > 0 {
		return post, summary, next.String()
	}
	anchor := new(big.Int).Add(big.NewInt(*state.TickAnchorWallMS), loss)
	anchor.Add(anchor, new(big.Int).Mul(n, big.NewInt(substrate.TickMS)))
	post.TickSeq = next.Int64()
	post.TickAnchorWallMS = newInt64(anchor.Int64())
	summary.TicksApplied, summary.TickSeqAfter = n.Int64(), next.Int64()
	summary.CatchupForfeitedMS = loss.Int64()
	if loss.Sign() > 0 {
		summary.CatchupReasonKey = cloneString(&catalog.CatchupReasonKey)
	}
	return post, summary, ""
}

func newInt64(value int64) *int64 { return &value }

func TestGardenAdvanceClockBoundaries(t *testing.T) {
	catalog := fixtureCatalog(t)
	const anchor = int64(1_000_000)
	type boundary struct {
		name, substrate string
		at, seq         int64
		locked, initial bool
		growing, noSalt bool
	}
	cases := []boundary{}
	for _, substrate := range catalog.Substrates {
		for _, point := range []struct {
			name string
			at   int64
		}{
			{"regression", anchor - 1}, {"equality", anchor},
			{"tick_minus_one", anchor + substrate.TickMS - 1}, {"tick", anchor + substrate.TickMS}, {"tick_plus_one", anchor + substrate.TickMS + 1},
			{"cap_minus_one", anchor + catalog.CatchupCapMS - 1}, {"cap", anchor + catalog.CatchupCapMS}, {"cap_plus_one", anchor + catalog.CatchupCapMS + 1},
			{"cap_plus_tick", anchor + catalog.CatchupCapMS + substrate.TickMS + 1}, {"maximum_server_ms", maxExactInteger},
		} {
			cases = append(cases, boundary{name: substrate.SubstrateID + "/" + point.name, substrate: substrate.SubstrateID, at: point.at})
		}
	}
	substrate, _ := catalog.Substrate("bare_metal")
	cases = append(cases,
		boundary{name: "locked", at: anchor, locked: true, initial: true},
		boundary{name: "initialize_at_zero", at: 0, initial: true},
		boundary{name: "maximum_counter_zero_ticks", at: anchor, seq: maxExactInteger},
		boundary{name: "last_safe_tick", at: anchor + substrate.TickMS, seq: maxExactInteger - 1},
		boundary{name: "last_two_safe_ticks", at: anchor + 2*substrate.TickMS, seq: maxExactInteger - 2},
		boundary{name: "overflow_one", at: anchor + substrate.TickMS, seq: maxExactInteger},
		boundary{name: "overflow_three", at: anchor + 3*substrate.TickMS, seq: maxExactInteger - 1},
		boundary{name: "overflow_before_growth", at: anchor + substrate.TickMS, seq: maxExactInteger, growing: true},
		boundary{name: "overflow_before_salt", at: anchor + substrate.TickMS, seq: maxExactInteger, noSalt: true},
	)
	if len(cases) != 49 {
		t.Fatalf("clock population changed: %d", len(cases))
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(catalog)
			state.TickSeq = test.seq
			if test.substrate != "" {
				state.SubstrateID = test.substrate
			}
			if !test.initial {
				salt := "0123456789abcdef"
				state.SaltHex, state.TickAnchorWallMS = &salt, newInt64(anchor)
			}
			if test.noSalt {
				state.SaltHex = nil
			}
			if test.growing {
				state.Plots = append(state.Plots, Plot{Row: 0, Col: 0, SpeciesID: "strain_a"})
			}
			before, err := EncodeState(state)
			if err != nil || ValidateAgainst(catalog, state) != nil {
				t.Fatalf("input is outside real SG2 grammar: %v", err)
			}
			input := AdvanceInput{ServerMS: test.at, Unlocked: !test.locked}
			if NeedsSalt(state, input.Unlocked) {
				input.Salt = "0123456789abcdef"
			}
			wantState, wantSummary, overflow := referenceClock(catalog, state, input)
			got, err := catalog.Advance(state, input)
			if overflow != "" {
				if err == nil {
					t.Fatalf("accepted out-of-domain counter: exact=%s got=%d post_valid=%v", overflow, state.TickSeq, ValidateShape(state))
				}
				post, _ := json.Marshal(state)
				if string(post) != string(before) {
					t.Fatalf("counter refusal mutated input: %s -> %s", before, post)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			post, err := EncodeState(state)
			want, _ := EncodeState(wantState)
			actualSummary, _ := json.Marshal(got)
			expectedSummary, _ := json.Marshal(wantSummary)
			if err != nil || ValidateAgainst(catalog, state) != nil || string(post) != string(want) || string(actualSummary) != string(expectedSummary) {
				t.Fatalf("clock bytes differ: state=%s want=%s summary=%s want=%s error=%v", post, want, actualSummary, expectedSummary, err)
			}
		})
	}
}
