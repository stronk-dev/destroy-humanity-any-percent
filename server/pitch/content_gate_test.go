package pitch

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"sort"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/minigame"
)

var updatePitchCorpus = flag.Bool("update-pitch-corpus", false, "regenerate the checked-in Pitch content corpus")

type contentGateCorpus struct {
	Version            int               `json:"version"`
	PitchContentHash   string            `json:"pitch_content_hash"`
	TransitionBudget   int               `json:"transition_budget"`
	Scenarios          []contentScenario `json:"scenarios"`
	ExponentBoundaries []exponentVector  `json:"exponent_boundaries"`
}

type contentScenario struct {
	Name             string            `json:"name"`
	Seed             uint64            `json:"seed"`
	Commands         []json.RawMessage `json:"commands"`
	ExpectedTerminal json.RawMessage   `json:"expected_terminal"`
	ExpectedResult   *minigame.Result  `json:"expected_result"`
	CoversCards      []string          `json:"covers_cards"`
	CoversHacks      []string          `json:"covers_hacks"`
	Assertions       []string          `json:"assertions"`
}

type exponentVector struct {
	Valuation string `json:"valuation"`
	Expected  int64  `json:"expected"`
}

type scenarioTarget struct {
	name       string
	desired    []string
	control    string
	assertions []string
}

func TestPitchContentGate(t *testing.T) {
	generated := generateContentGate(t)
	encoded, err := json.MarshalIndent(generated, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	path := "../../testdata/pitch/content-gate-v1.json"
	if *updatePitchCorpus {
		if err := os.MkdirAll("../../testdata/pitch", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, encoded) {
		t.Fatal("Pitch content corpus drifted; regenerate with make pitch-corpus in a BALANCE-CHANGE commit")
	}
}

func generateContentGate(t *testing.T) contentGateCorpus {
	t.Helper()
	data, catalog := loadFixture(t)
	targets := []scenarioTarget{
		{name: "ab-test", desired: []string{"ab_test"}},
		{name: "buzzword", desired: []string{"buzzword"}},
		{name: "dark-pattern-trigger", desired: []string{"dark_pattern"}, assertions: []string{"shape_trigger:dark_pattern"}},
		{name: "growth-loop", desired: []string{"growth_loop"}},
		{name: "infinite-scroll", desired: []string{"infinite_scroll"}},
		{name: "pivot-trigger", desired: []string{"pivot"}, assertions: []string{"shape_trigger:pivot"}},
		{name: "stealth-partner-absent", desired: []string{"stealth_mode"}, assertions: []string{"chain_partner_absent:stealth_mode"}},
		{name: "synergy-partner-absent", desired: []string{"synergy_deck"}, assertions: []string{"chain_partner_absent:synergy_deck"}},
		{name: "stealth-partner-present", desired: []string{"pivot", "stealth_mode"}, assertions: []string{"chain_partner_present:stealth_mode"}},
		{name: "synergy-partner-present", desired: []string{"ab_test", "synergy_deck"}, assertions: []string{"chain_partner_present:synergy_deck"}},
		{name: "dark-pattern-control", desired: []string{"dark_pattern"}, control: "dark_pattern", assertions: []string{"shape_control:dark_pattern"}},
		{name: "pivot-control", desired: []string{"pivot"}, control: "pivot", assertions: []string{"shape_control:pivot"}},
	}
	corpus := contentGateCorpus{Version: 1, PitchContentHash: ContentHash(data), Scenarios: make([]contentScenario, 0, len(targets)),
		ExponentBoundaries: []exponentVector{{Valuation: "0", Expected: 0}, {Valuation: "9e-1", Expected: -1},
			{Valuation: "9.99999999999e11", Expected: 11}, {Valuation: "1e12", Expected: 12}, {Valuation: "1e1000001", Expected: 1_000_000}}}
	coveredCards, coveredHacks := map[string]bool{}, map[string]bool{}
	for _, target := range targets {
		var scenario contentScenario
		found := false
		for seed := uint64(1); seed <= 20_000 && !found; seed++ {
			candidate, bought, played, ok := generateScenario(data, corpus.PitchContentHash, catalog, seed, target)
			if !ok || !containsAll(bought, target.desired) {
				continue
			}
			candidate.Name, candidate.Assertions = target.name, append([]string{}, target.assertions...)
			candidate.CoversCards = uniqueBaseIDs(played)
			candidate.CoversHacks = append([]string(nil), bought...)
			sort.Strings(candidate.CoversHacks)
			scenario, found = candidate, true
		}
		if !found {
			t.Fatalf("no deterministic scenario found for %s", target.name)
		}
		for _, id := range scenario.CoversCards {
			coveredCards[id] = true
		}
		for _, id := range scenario.CoversHacks {
			coveredHacks[id] = true
		}
		corpus.TransitionBudget += len(scenario.Commands)
		corpus.Scenarios = append(corpus.Scenarios, scenario)
		validateScenarioAssertions(t, scenario)
	}
	for _, row := range catalog.MetricCards {
		if !coveredCards[row.CardID] {
			t.Fatalf("card %s has no affecting corpus scenario", row.CardID)
		}
	}
	for _, row := range catalog.GrowthHacks {
		if !coveredHacks[row.HackID] {
			t.Fatalf("hack %s has no affecting corpus scenario", row.HackID)
		}
	}
	if corpus.TransitionBudget != commandCount(corpus.Scenarios) {
		t.Fatal("Pitch transition budget is not the exact corpus command sum")
	}
	for _, vector := range corpus.ExponentBoundaries {
		if got := bestExponent(vector.Valuation, 1_000_000); got != vector.Expected {
			t.Fatalf("exponent %s=%d want %d", vector.Valuation, got, vector.Expected)
		}
	}
	return corpus
}

func validateScenarioAssertions(t *testing.T, scenario contentScenario) {
	t.Helper()
	for _, assertion := range scenario.Assertions {
		parts := bytes.Split([]byte(assertion), []byte(":"))
		if len(parts) != 2 {
			t.Fatalf("invalid assertion %q", assertion)
		}
		kind, hackID := string(parts[0]), string(parts[1])
		switch kind {
		case "chain_partner_present", "chain_partner_absent":
			partners := map[string]string{"stealth_mode": "pivot", "synergy_deck": "ab_test"}
			partner := partners[hackID]
			var terminal Snapshot
			if json.Unmarshal(scenario.ExpectedTerminal, &terminal) != nil {
				t.Fatalf("scenario %s has invalid terminal bytes", scenario.Name)
			}
			present := contains(terminal.SlottedHacks, partner)
			if partner == "" || present != (kind == "chain_partner_present") {
				t.Fatalf("scenario %s does not prove %s", scenario.Name, assertion)
			}
		case "shape_trigger", "shape_control":
			selected, found := firstPlayedAfterPurchase(scenario.Commands, hackID)
			if !found {
				t.Fatalf("scenario %s never uses %s", scenario.Name, hackID)
			}
			triggered := len(selected) == 4
			if hackID == "dark_pattern" {
				counts := map[string]int{}
				triggered = false
				for _, instance := range selected {
					base, _ := BaseCardID(instance)
					counts[base]++
					triggered = triggered || counts[base] >= 2
				}
			}
			if triggered != (kind == "shape_trigger") {
				t.Fatalf("scenario %s does not prove %s", scenario.Name, assertion)
			}
		default:
			t.Fatalf("unknown assertion %q", assertion)
		}
	}
}

func firstPlayedAfterPurchase(commands []json.RawMessage, hackID string) ([]string, bool) {
	purchased := false
	for _, raw := range commands {
		var header struct {
			Kind    string   `json:"kind"`
			OfferID string   `json:"offer_id"`
			CardIDs []string `json:"card_ids"`
		}
		if json.Unmarshal(raw, &header) != nil {
			return nil, false
		}
		if header.Kind == "buy_hack" && bytes.HasSuffix([]byte(header.OfferID), []byte("."+hackID)) {
			purchased = true
		} else if purchased && header.Kind == "play_hand" {
			return header.CardIDs, true
		}
	}
	return nil, false
}

func generateScenario(data []byte, hash string, catalog *Catalog, seed uint64, target scenarioTarget) (contentScenario, []string, []string, bool) {
	tenant := NewTenant()
	snapshotBytes, err := tenant.Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: seed,
		ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: data, ContentHash: hash, ContentSchemaVersion: 1})
	if err != nil {
		return contentScenario{}, nil, nil, false
	}
	commands, bought, played := []json.RawMessage{}, []string{}, []string{}
	controlUsed := false
	for revision := int64(1); revision <= 40; revision++ {
		snapshot, decodeErr := decodeSnapshot(snapshotBytes)
		if decodeErr != nil {
			return contentScenario{}, bought, played, false
		}
		var command json.RawMessage
		if snapshot.Phase == "playing" {
			selected := bestCorpusHand(snapshot.Hand, snapshot.SlottedHacks, catalog)
			if target.control != "" && !controlUsed && contains(snapshot.SlottedHacks, target.control) {
				selected = bestControlHand(snapshot.Hand, snapshot.SlottedHacks, catalog, target.control)
				controlUsed = true
			}
			played = append(played, selected...)
			command, _ = json.Marshal(map[string]any{"card_ids": selected, "kind": "play_hand"})
		} else if snapshot.Phase == "shop" {
			offerID := ""
			for _, wanted := range target.desired {
				if contains(snapshot.SlottedHacks, wanted) {
					continue
				}
				for _, offer := range snapshot.ShopOffers {
					if offer.HackID == wanted && offer.Price <= snapshot.RunCurrency {
						offerID, bought = offer.OfferID, append(bought, wanted)
						break
					}
				}
				if offerID != "" {
					break
				}
			}
			if offerID == "" {
				command = json.RawMessage(`{"kind":"end_shop"}`)
			} else {
				command, _ = json.Marshal(map[string]any{"kind": "buy_hack", "offer_id": offerID})
			}
		} else {
			return contentScenario{}, bought, played, false
		}
		commands = append(commands, command)
		output, applyErr := tenant.Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: seed, Revision: revision,
			Snapshot: snapshotBytes, Command: command, ScalingInputs: map[string]int64{ScalingDestination: 1},
			Content: data, ContentHash: hash, ContentSchemaVersion: 1})
		if applyErr != nil {
			return contentScenario{}, bought, played, false
		}
		snapshotBytes = output.Snapshot
		if output.Result != nil {
			return contentScenario{Seed: seed, Commands: commands, ExpectedTerminal: bytes.Clone(snapshotBytes), ExpectedResult: output.Result}, bought, played, true
		}
	}
	return contentScenario{}, bought, played, false
}

func bestCorpusHand(hand, hacks []string, catalog *Catalog) []string {
	return bestMatchingHand(hand, hacks, catalog, func(selected []string) bool { return len(selected) <= int(catalog.Policy.PlaySize) })
}

func bestControlHand(hand, hacks []string, catalog *Catalog, controlHack string) []string {
	return bestMatchingHand(hand, hacks, catalog, func(selected []string) bool {
		if controlHack == "pivot" {
			return len(selected) < int(catalog.Policy.PlaySize)
		}
		counts := map[string]int{}
		for _, instance := range selected {
			base, _ := BaseCardID(instance)
			counts[base]++
		}
		for _, count := range counts {
			if count > 1 {
				return false
			}
		}
		return len(selected) <= int(catalog.Policy.PlaySize)
	})
}

func bestMatchingHand(hand, hacks []string, catalog *Catalog, allowed func([]string) bool) []string {
	best, bestValue := []string{}, decimal.Zero
	for mask := 0; mask < 1<<len(hand); mask++ {
		selected := []string{}
		for index := range hand {
			if mask&(1<<index) != 0 {
				selected = append(selected, hand[index])
			}
		}
		if !allowed(selected) {
			continue
		}
		value, err := score(selected, hacks, catalog)
		if err == nil && (value.Gt(bestValue) || value.Eq(bestValue) && len(selected) > len(best)) {
			best, bestValue = append([]string(nil), selected...), value
		}
	}
	sort.Strings(best)
	return best
}

func uniqueBaseIDs(instances []string) []string {
	set := map[string]bool{}
	for _, instance := range instances {
		base, _ := BaseCardID(instance)
		set[base] = true
	}
	result := make([]string, 0, len(set))
	for id := range set {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func commandCount(scenarios []contentScenario) int {
	total := 0
	for _, scenario := range scenarios {
		total += len(scenario.Commands)
	}
	return total
}
func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func containsAll(values, wanted []string) bool {
	for _, value := range wanted {
		if !contains(values, value) {
			return false
		}
	}
	return true
}
