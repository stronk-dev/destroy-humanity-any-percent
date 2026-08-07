package pitch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/minigame"
)

const (
	EngineRef            = "pitch"
	RunSubstream         = "pitch.run.v1"
	DeckSubstream        = "pitch.deck.v1"
	ShopSubstream        = "pitch.shop.v1"
	ScalingDestination   = "minigame.pitch"
	OutcomeFunded        = "funded"
	OutcomeFundingFailed = "funding_failed"
)

var errorTaxonomy = []string{"duplicate_card", "hack_slots_full", "hand_too_large", "illegal_phase", "insufficient_currency", "unknown_card", "unknown_offer"}

type ShopOffer struct {
	HackID  string `json:"hack_id"`
	OfferID string `json:"offer_id"`
	Price   int64  `json:"price"`
}

type Snapshot struct {
	Phase              string      `json:"phase"`
	Round              int64       `json:"round"`
	HandsRemaining     int64       `json:"hands_remaining"`
	DeckCount          int64       `json:"deck_count"`
	Hand               []string    `json:"hand"`
	SlottedHacks       []string    `json:"slotted_hacks"`
	RunCurrency        int64       `json:"run_currency"`
	ShopOffers         []ShopOffer `json:"shop_offers"`
	FundingTarget      string      `json:"funding_target"`
	RoundBestValuation string      `json:"round_best_valuation"`
	Revision           int64       `json:"revision"`
	PitchContentHash   string      `json:"pitch_content_hash"`
	PitchSchemaVersion int         `json:"pitch_schema_version"`
}

type Tenant struct{}

func NewTenant() Tenant { return Tenant{} }

func (Tenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: EngineRef, EngineVersion: EngineVersion,
		CommandSchema: "pitch.command.v1", SnapshotSchema: "pitch.snapshot.v1", ResultSchema: "minigame.result.v1",
		Modes: []minigame.Mode{minigame.ModeSolo}, ErrorTaxonomy: append([]string(nil), errorTaxonomy...),
		Destinations: map[string]minigame.DestinationClass{ScalingDestination: minigame.DestinationBreadth}}
}

func (Tenant) ValidateCommand(data json.RawMessage) error {
	_, rejection := decodeCommand(data)
	return rejection
}

func (Tenant) ValidateSnapshot(data json.RawMessage) error {
	_, err := decodeSnapshot(data)
	return err
}

func (Tenant) ValidateResult(result *minigame.Result) error {
	if result == nil {
		return nil
	}
	if result.Outcome != OutcomeFunded && result.Outcome != OutcomeFundingFailed || result.RatingDelta != nil || len(result.ScoreFacts) != 2 ||
		result.ScoreFacts[0].Kind != "pitch.best_hand_exponent" || result.ScoreFacts[1].Kind != "pitch.final_round" ||
		result.ScoreFacts[0].Value > 1_000_000 || result.ScoreFacts[1].Value < 1 || result.ScoreFacts[1].Value > 8 {
		return minigame.ErrInvalidTenant
	}
	return nil
}

func (Tenant) Create(input minigame.CreateInput) (json.RawMessage, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo || !validScaling(input.ScalingInputs) {
		return nil, minigame.ErrInvalidTenant
	}
	target, _ := catalog.FundingTarget(1)
	hand, deckCount := deal(catalog, input.Seed, 1, 1)
	snapshot := Snapshot{Phase: "playing", Round: 1, HandsRemaining: catalog.Policy.HandsPerRound,
		DeckCount: deckCount, Hand: hand, SlottedHacks: []string{}, RunCurrency: catalog.Policy.StartCurrency,
		ShopOffers: []ShopOffer{}, FundingTarget: target, RoundBestValuation: "0", Revision: 1,
		PitchContentHash: input.ContentHash, PitchSchemaVersion: input.ContentSchemaVersion}
	return encodeSnapshot(snapshot)
}

func (Tenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo || !validScaling(input.ScalingInputs) {
		return minigame.ApplyOutput{}, minigame.ErrInvalidTenant
	}
	snapshot, err := decodeSnapshot(input.Snapshot)
	if err != nil || snapshot.Revision != input.Revision || snapshot.PitchContentHash != input.ContentHash ||
		snapshot.PitchSchemaVersion != input.ContentSchemaVersion || validateSnapshotAgainstCatalog(snapshot, catalog) != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	command, rejection := decodeCommand(input.Command)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	var result *minigame.Result
	switch command.Kind {
	case "play_hand":
		result, rejection = playHand(&snapshot, command.CardIDs, catalog, input.Seed)
	case "buy_hack":
		rejection = buyHack(&snapshot, command.OfferID, catalog)
	case "end_shop":
		rejection = endShop(&snapshot, catalog, input.Seed)
	default:
		rejection = reject("illegal_phase", "unknown command kind")
	}
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	snapshot.Revision = input.Revision + 1
	encoded, err := encodeSnapshot(snapshot)
	if err != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	return minigame.ApplyOutput{Snapshot: encoded, Result: result}, nil
}

func encodeSnapshot(snapshot Snapshot) (json.RawMessage, error) {
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	var canonical map[string]json.RawMessage
	if json.Unmarshal(encoded, &canonical) != nil {
		return nil, minigame.ErrTenantDivergence
	}
	return json.Marshal(canonical)
}

type command struct {
	Kind    string
	CardIDs []string
	OfferID string
}

func decodeCommand(data []byte) (command, error) {
	if !uniqueJSONKeys(data) {
		return command{}, reject("illegal_phase", "command keys are not unique")
	}
	var header struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(data, &header) != nil {
		return command{}, reject("illegal_phase", "command is not an object")
	}
	switch header.Kind {
	case "play_hand":
		var wire struct {
			Kind    string   `json:"kind"`
			CardIDs []string `json:"card_ids"`
		}
		if strictDecode(data, &wire) != nil || wire.CardIDs == nil {
			return command{}, reject("unknown_card", "play_hand schema mismatch")
		}
		if len(wire.CardIDs) > 4 {
			return command{}, reject("hand_too_large", "played hand exceeds play_size")
		}
		if !strictlySorted(wire.CardIDs) {
			return command{}, reject("duplicate_card", "card_ids must be unique and byte-sorted")
		}
		return command{Kind: header.Kind, CardIDs: wire.CardIDs}, nil
	case "buy_hack":
		var wire struct {
			Kind    string `json:"kind"`
			OfferID string `json:"offer_id"`
		}
		if strictDecode(data, &wire) != nil || wire.OfferID == "" {
			return command{}, reject("unknown_offer", "buy_hack schema mismatch")
		}
		return command{Kind: header.Kind, OfferID: wire.OfferID}, nil
	case "end_shop":
		var wire struct {
			Kind string `json:"kind"`
		}
		if strictDecode(data, &wire) != nil {
			return command{}, reject("illegal_phase", "end_shop schema mismatch")
		}
		return command{Kind: header.Kind}, nil
	default:
		return command{}, reject("illegal_phase", "unknown command kind")
	}
}

func decodeSnapshot(data []byte) (Snapshot, error) {
	if !uniqueJSONKeys(data) {
		return Snapshot{}, minigame.ErrInvalidTenant
	}
	var value Snapshot
	if strictDecode(data, &value) != nil || (value.Phase != "playing" && value.Phase != "shop" && value.Phase != "terminal") ||
		value.Round < 1 || value.Round > 8 || value.HandsRemaining < 0 || value.HandsRemaining > 3 || value.DeckCount < 0 || value.DeckCount > 24 ||
		value.Hand == nil || value.SlottedHacks == nil || value.ShopOffers == nil || !strictlySorted(value.Hand) || !strictlySorted(value.SlottedHacks) ||
		value.RunCurrency < 0 || value.RunCurrency > decimal.MaxExactInteger || value.Revision < 1 || value.PitchSchemaVersion != SchemaVersion ||
		!strings.HasPrefix(value.PitchContentHash, "sha256:") || len(value.PitchContentHash) != 71 {
		return Snapshot{}, minigame.ErrInvalidTenant
	}
	for _, encoded := range []string{value.FundingTarget, value.RoundBestValuation} {
		parsed, err := decimal.ParseCanonical(encoded)
		if err != nil || parsed.Lt(decimal.Zero) {
			return Snapshot{}, minigame.ErrInvalidTenant
		}
	}
	prior := ""
	for _, offer := range value.ShopOffers {
		if offer.OfferID == "" || !idPattern.MatchString(offer.HackID) || offer.Price < 0 || offer.Price > decimal.MaxExactInteger || prior >= offer.OfferID && prior != "" {
			return Snapshot{}, minigame.ErrInvalidTenant
		}
		prior = offer.OfferID
	}
	return value, nil
}

func validateSnapshotAgainstCatalog(value Snapshot, catalog *Catalog) error {
	target, ok := catalog.FundingTarget(value.Round)
	if !ok || target != value.FundingTarget || len(value.Hand) > int(catalog.Policy.HandSize) || len(value.SlottedHacks) > int(catalog.Policy.HackSlots) {
		return minigame.ErrInvalidTenant
	}
	seenInstances := map[string]bool{}
	for _, instance := range value.Hand {
		base, ok := BaseCardID(instance)
		if !ok || seenInstances[instance] {
			return minigame.ErrInvalidTenant
		}
		card, exists := catalog.Card(base)
		if !exists {
			return minigame.ErrInvalidTenant
		}
		var ordinal int64
		if _, scanErr := fmt.Sscanf(instance, base+"#%d", &ordinal); scanErr != nil || ordinal < 1 || ordinal > card.Copies {
			return minigame.ErrInvalidTenant
		}
		seenInstances[instance] = true
	}
	for _, id := range value.SlottedHacks {
		if _, ok := catalog.Hack(id); !ok {
			return minigame.ErrInvalidTenant
		}
	}
	for _, offer := range value.ShopOffers {
		hack, ok := catalog.Hack(offer.HackID)
		if !ok || hack.Price != offer.Price {
			return minigame.ErrInvalidTenant
		}
	}
	return nil
}

func playHand(snapshot *Snapshot, selected []string, catalog *Catalog, seed uint64) (*minigame.Result, error) {
	if snapshot.Phase != "playing" {
		return nil, reject("illegal_phase", "play_hand requires playing phase")
	}
	if len(selected) > int(catalog.Policy.PlaySize) {
		return nil, reject("hand_too_large", "played hand exceeds play_size")
	}
	hand := make(map[string]bool, len(snapshot.Hand))
	for _, id := range snapshot.Hand {
		hand[id] = true
	}
	for _, id := range selected {
		if !hand[id] {
			return nil, reject("unknown_card", "selected card is not in hand")
		}
	}
	valuation, err := score(selected, snapshot.SlottedHacks, catalog)
	if err != nil {
		return nil, minigame.ErrTenantDivergence
	}
	best := decimal.FromString(snapshot.RoundBestValuation)
	if valuation.Gt(best) {
		best = valuation
	}
	snapshot.RoundBestValuation = best.String()
	snapshot.HandsRemaining--
	target := decimal.FromString(snapshot.FundingTarget)
	if best.Gte(target) {
		if snapshot.Round == int64(len(catalog.FundingCurve)) {
			snapshot.Phase, snapshot.ShopOffers = "terminal", []ShopOffer{}
			return terminalResult(snapshot, catalog, OutcomeFunded), nil
		}
		if snapshot.RunCurrency > decimal.MaxExactInteger-catalog.Policy.RoundClearCurrency {
			return nil, minigame.ErrTenantDivergence
		}
		snapshot.RunCurrency += catalog.Policy.RoundClearCurrency
		snapshot.Phase, snapshot.Hand = "shop", []string{}
		snapshot.ShopOffers = shopOffers(catalog, seed, snapshot.Round, snapshot.SlottedHacks)
		return nil, nil
	}
	if snapshot.HandsRemaining == 0 {
		snapshot.Phase, snapshot.ShopOffers = "terminal", []ShopOffer{}
		return terminalResult(snapshot, catalog, OutcomeFundingFailed), nil
	}
	handNumber := catalog.Policy.HandsPerRound - snapshot.HandsRemaining + 1
	snapshot.Hand, snapshot.DeckCount = deal(catalog, seed, snapshot.Round, handNumber)
	return nil, nil
}

func buyHack(snapshot *Snapshot, offerID string, catalog *Catalog) error {
	if snapshot.Phase != "shop" {
		return reject("illegal_phase", "buy_hack requires shop phase")
	}
	if len(snapshot.SlottedHacks) >= int(catalog.Policy.HackSlots) {
		return reject("hack_slots_full", catalog.Policy.HackSlotsReasonKey)
	}
	index := -1
	for at, offer := range snapshot.ShopOffers {
		if offer.OfferID == offerID {
			index = at
			break
		}
	}
	if index < 0 {
		return reject("unknown_offer", "offer is not active")
	}
	offer := snapshot.ShopOffers[index]
	if snapshot.RunCurrency < offer.Price {
		return reject("insufficient_currency", "run currency is below offer price")
	}
	if _, exists := catalog.Hack(offer.HackID); !exists {
		return minigame.ErrTenantDivergence
	}
	snapshot.RunCurrency -= offer.Price
	snapshot.SlottedHacks = append(snapshot.SlottedHacks, offer.HackID)
	sort.Strings(snapshot.SlottedHacks)
	snapshot.ShopOffers = append(snapshot.ShopOffers[:index:index], snapshot.ShopOffers[index+1:]...)
	return nil
}

func endShop(snapshot *Snapshot, catalog *Catalog, seed uint64) error {
	if snapshot.Phase != "shop" || snapshot.Round >= int64(len(catalog.FundingCurve)) {
		return reject("illegal_phase", "end_shop requires non-final shop")
	}
	snapshot.Round++
	snapshot.Phase = "playing"
	snapshot.HandsRemaining = catalog.Policy.HandsPerRound
	snapshot.Hand, snapshot.DeckCount = deal(catalog, seed, snapshot.Round, 1)
	snapshot.ShopOffers = []ShopOffer{}
	snapshot.RoundBestValuation = "0"
	snapshot.FundingTarget, _ = catalog.FundingTarget(snapshot.Round)
	return nil
}

func score(selected, hacks []string, catalog *Catalog) (decimal.Decimal, error) {
	flat := decimal.Zero
	cardFactors := []decimal.Decimal{}
	baseCounts := map[string]int{}
	for _, hackID := range hacks {
		hack, ok := catalog.Hack(hackID)
		if !ok {
			return decimal.NaN, ErrInvalidCatalog
		}
		switch hack.Effect.Kind {
		case "flat_add":
			flat = flat.Add(decimal.FromString(hack.Effect.Amount))
		case "card_factor":
			cardFactors = append(cardFactors, decimal.FromString(hack.Effect.Factor))
		}
	}
	terms := make([]decimal.Decimal, 0, len(selected))
	for _, instance := range selected {
		base, ok := BaseCardID(instance)
		if !ok {
			return decimal.NaN, ErrInvalidCatalog
		}
		card, ok := catalog.Card(base)
		if !ok {
			return decimal.NaN, ErrInvalidCatalog
		}
		baseCounts[base]++
		value := decimal.FromString(card.BaseMetric).Add(flat)
		for _, factor := range cardFactors {
			value = value.Mul(factor)
		}
		terms = append(terms, value)
	}
	total := decimal.SumDeterministic(terms)
	if len(terms) == 0 {
		total = decimal.Zero
	}
	owned := map[string]bool{}
	for _, id := range hacks {
		owned[id] = true
	}
	pair := false
	for _, count := range baseCounts {
		if count >= 2 {
			pair = true
		}
	}
	for _, hackID := range hacks {
		hack, _ := catalog.Hack(hackID)
		switch hack.Effect.Kind {
		case "shape_factor":
			if hack.Effect.Shape == "pair" && pair || hack.Effect.Shape == "full_hand" && len(selected) == int(catalog.Policy.PlaySize) {
				total = total.Mul(decimal.FromString(hack.Effect.Factor))
			}
		case "chain_factor":
			if owned[hack.Effect.PartnerHackID] {
				total = total.Mul(decimal.FromString(hack.Effect.Factor))
			}
		}
	}
	total = total.Quantize(decimal.CanonicalSignificantDigits)
	if !total.IsStateValue() || total.Lt(decimal.Zero) {
		return decimal.NaN, ErrInvalidCatalog
	}
	return total, nil
}

func terminalResult(snapshot *Snapshot, catalog *Catalog, outcome string) *minigame.Result {
	exponent := bestExponent(snapshot.RoundBestValuation, catalog.Policy.BestExponentHardcap)
	return &minigame.Result{Outcome: outcome, RatingDelta: nil, ScoreFacts: []minigame.ScoreFact{
		{Kind: "pitch.best_hand_exponent", Value: exponent}, {Kind: "pitch.final_round", Value: snapshot.Round}}}
}

func bestExponent(valuation string, hardcap int64) int64 {
	if valuation == "0" {
		return 0
	}
	exponent := decimal.FromString(valuation).Exponent()
	if exponent > hardcap {
		return hardcap
	}
	return exponent
}

func deal(catalog *Catalog, seed uint64, round, handNumber int64) ([]string, int64) {
	deck := catalog.CardInstances()
	random := coordinateRandom(seed, DeckSubstream, round)
	for index := len(deck) - 1; index > 0; index-- {
		swap := int(random.Bound(uint64(index + 1)))
		deck[index], deck[swap] = deck[swap], deck[index]
	}
	start := int((handNumber - 1) * catalog.Policy.HandSize)
	end := start + int(catalog.Policy.HandSize)
	hand := append([]string(nil), deck[start:end]...)
	sort.Strings(hand)
	return hand, int64(len(deck) - end)
}

func shopOffers(catalog *Catalog, seed uint64, round int64, ownedIDs []string) []ShopOffer {
	owned := map[string]bool{}
	for _, id := range ownedIDs {
		owned[id] = true
	}
	pool := make([]GrowthHack, 0, len(catalog.GrowthHacks))
	for _, hack := range catalog.GrowthHacks {
		if !owned[hack.HackID] {
			pool = append(pool, hack)
		}
	}
	random := coordinateRandom(seed, ShopSubstream, round)
	count := min(int(catalog.Policy.ShopSize), len(pool))
	result := make([]ShopOffer, 0, count)
	for slot := 1; slot <= count; slot++ {
		var total uint64
		for _, hack := range pool {
			total += uint64(hack.DraftWeight)
		}
		draw := random.Bound(total)
		chosen := 0
		for index, hack := range pool {
			if draw < uint64(hack.DraftWeight) {
				chosen = index
				break
			}
			draw -= uint64(hack.DraftWeight)
		}
		hack := pool[chosen]
		result = append(result, ShopOffer{OfferID: fmt.Sprintf("pitch.offer.%d.%d.%s", round, slot, hack.HackID), HackID: hack.HackID, Price: hack.Price})
		pool = append(pool[:chosen:chosen], pool[chosen+1:]...)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].OfferID < result[right].OfferID })
	return result
}

func coordinateRandom(seed uint64, label string, coordinate int64) *determinism.SplitMix64 {
	runSeed := determinism.Substream(seed, RunSubstream).Next()
	return determinism.Substream(runSeed^uint64(coordinate), label)
}

func catalogForInput(data []byte, hash string, schemaVersion int) (*Catalog, error) {
	if len(data) == 0 || hash != ContentHash(data) || schemaVersion != SchemaVersion {
		return nil, ErrInvalidCatalog
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return LoadCatalog(data, Declarations{CopyKeys: keys})
}

func validScaling(values map[string]int64) bool {
	return len(values) == 1 && values[ScalingDestination] == 1
}
func strictlySorted(values []string) bool {
	for index, value := range values {
		if value == "" || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}
func reject(code, detail string) *minigame.Rejection {
	return &minigame.Rejection{Code: code, Detail: detail}
}

func strictDecode(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return minigame.ErrInvalidTenant
	}
	return nil
}
