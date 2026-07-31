package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

const (
	CurrentVersion           = 11
	millisecondCursorVersion = 4
	maxOfflineSpans          = 256
)

var ErrInvalidState = errors.New("invalid saved state")
var stateMechanicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)

type State struct {
	Ledger                *economy.Ledger
	GeneratorCounts       map[string]int64
	EvaluatedThrough      time.Time
	ComputeCreditMS       int64
	ManualTokenMilli      int64
	ManualTokenRefilledAt time.Time
	GatesCrossed          map[string]bool
	RunSeq                int64
	DoctrinesByTransition map[string]string
	StructureID           string
	LedgerFactKinds       map[string]bool
	MeterBands            map[string]int
	RegionTraits          map[string]bool
	RouteKnowledgeBalance int64
	HintsUnlocked         map[string]bool
	CompactMember         bool
	CompactTithePPM       int64
	CompactSolidarityPPM  int64
	CompactSamples        []CompactSample
	Tier                  int64
	LifetimeValue         decimal.Decimal
	OfferState            *ExitOfferState
	RunStartedAt          time.Time
	RunPreTimer           bool
	OfflineSpans          []OfflineSpan
	CollapsedOfflineMS    int64
	ReputationLevel       int64
	ReputationUnlockPPM   int64
	NetworkSlots          []NetworkSlot
	CloutLifetime         int64
	Soul                  int64
	AgeMS                 int64
	Notoriety             int64
	AdvisorMode           bool
	ExitHistory           []ExitRecord
	FactionID             string
	IncorporatedAt        time.Time
	StockUnits            int64
	StockProgressMS       int64
	ConsumedStockUnits    int64
	GuildTitheCarryPPM    int64
	GuildBoundarySeq      int64
	GuildConsumedWindow   int64
	// FactionStockResource is derived from FactionID and the pinned catalog at
	// runtime. It is intentionally not persisted as a second source of truth.
	FactionStockResource string
}

type CompactSample struct {
	HourStart     time.Time
	CompliancePPM int64
	CoveredMS     int64
}

type ExitOfferState struct {
	OfferID   string
	ExitType  string
	TermsJSON json.RawMessage
	SpawnedAt time.Time
	ExpiresAt time.Time
}

type OfflineSpan struct {
	From time.Time
	To   time.Time
}

type NetworkSlot struct {
	Slot       string `json:"slot"`
	CarriedRef string `json:"carried_ref"`
}

type ExitRecord struct {
	RunID           int64
	ExitType        string
	OccurredAt      time.Time
	ReputationDelta int64
}

type stateV1 struct {
	Balances map[string]string `json:"balances"`
}

type stateV2 struct {
	Balances         map[string]string `json:"balances"`
	Generators       map[string]int64  `json:"generators"`
	EvaluatedThrough string            `json:"evaluated_through"`
}

type stateV4 struct {
	Balances              map[string]string `json:"balances"`
	Generators            map[string]int64  `json:"generators"`
	EvaluatedThrough      string            `json:"evaluated_through"`
	ComputeCreditMS       int64             `json:"compute_credit_ms"`
	ManualTokenMilli      int64             `json:"manual_token_milli"`
	ManualTokenRefilledAt string            `json:"manual_token_refilled_at"`
}

type stateV5 struct {
	Balances              map[string]string `json:"balances"`
	Generators            map[string]int64  `json:"generators"`
	EvaluatedThrough      string            `json:"evaluated_through"`
	ComputeCreditMS       int64             `json:"compute_credit_ms"`
	ManualTokenMilli      int64             `json:"manual_token_milli"`
	ManualTokenRefilledAt string            `json:"manual_token_refilled_at"`
	GatesCrossed          map[string]bool   `json:"gates_crossed"`
	RunSeq                int64             `json:"run_seq"`
	DoctrinesByTransition map[string]string `json:"doctrines_by_transition"`
	StructureID           string            `json:"structure_id"`
	LedgerFactKinds       []string          `json:"ledger_fact_kinds"`
	MeterBands            map[string]int    `json:"meter_bands"`
	RegionTraits          []string          `json:"region_traits"`
	RouteKnowledgeBalance int64             `json:"route_knowledge_balance"`
	HintsUnlocked         []string          `json:"hints_unlocked"`
}

type stateV6 struct {
	stateV5
	CompactMember        bool               `json:"compact_member"`
	CompactTithePPM      int64              `json:"compact_tithe_ppm"`
	CompactSolidarityPPM int64              `json:"compact_solidarity_ppm"`
	CompactSamples       []rawCompactSample `json:"compact_solidarity_samples"`
}

type stateV7 struct {
	stateV6
	Tier                int64              `json:"tier"`
	LifetimeValue       string             `json:"lifetime_value"`
	OfferState          *rawExitOfferState `json:"offer_state"`
	RunStartedAtMS      int64              `json:"run_started_at_ms"`
	OfflineSpans        []rawOfflineSpan   `json:"offline_spans"`
	ReputationLevel     int64              `json:"reputation_level"`
	ReputationUnlockPPM int64              `json:"reputation_unlock_ppm"`
	NetworkSlots        []NetworkSlot      `json:"network_slots"`
	CloutLifetime       int64              `json:"clout_lifetime"`
	Soul                int64              `json:"soul"`
	AgeMS               int64              `json:"age_ms"`
	Notoriety           int64              `json:"notoriety"`
	AdvisorMode         bool               `json:"advisor_mode"`
	ExitHistory         []rawExitRecord    `json:"exit_history"`
}

type stateV8 struct {
	stateV7
	RunPreTimer bool `json:"run_pre_timer"`
}

type stateV9 struct {
	stateV8
	CollapsedOfflineMS int64 `json:"collapsed_offline_ms"`
}

type stateV10 struct {
	stateV9
	FactionID          *string `json:"faction_id"`
	IncorporatedAtMS   *int64  `json:"incorporated_at_ms"`
	StockUnits         int64   `json:"stock_units"`
	StockProgressMS    int64   `json:"stock_progress_ms"`
	ConsumedStockUnits int64   `json:"consumed_stock_units"`
}

type stateV11 struct {
	stateV10
	GuildTitheCarryPPM  int64 `json:"guild_tithe_carry_ppm"`
	GuildBoundarySeq    int64 `json:"guild_boundary_seq"`
	GuildConsumedWindow int64 `json:"guild_consumed_window_units"`
}

type rawExitOfferState struct {
	OfferID     string          `json:"offer_id"`
	ExitType    string          `json:"exit_type"`
	TermsJSON   json.RawMessage `json:"terms_json"`
	SpawnedAtMS int64           `json:"spawned_at_ms"`
	ExpiresAtMS int64           `json:"expires_at_ms"`
}

type rawOfflineSpan struct {
	FromMS int64 `json:"from_ms"`
	ToMS   int64 `json:"to_ms"`
}

type rawExitRecord struct {
	RunID           int64  `json:"run_id"`
	ExitType        string `json:"exit_type"`
	OccurredAtMS    int64  `json:"occurred_at_ms"`
	ReputationDelta int64  `json:"reputation_delta"`
}

type rawCompactSample struct {
	HourStart     string `json:"hour_start"`
	CompliancePPM int64  `json:"compliance_ppm"`
	CoveredMS     int64  `json:"covered_ms"`
}

func ConstantsHash(catalogBytes []byte) string {
	digest := sha256.Sum256(catalogBytes)
	return "sha256:" + hex.EncodeToString(digest[:])
}

// ConstantsHashArtifacts identifies an immutable, named set of exact balance
// artifacts. Length framing keeps names and bytes unambiguous; sorted names
// make map iteration irrelevant.
func ConstantsHashArtifacts(artifacts map[string][]byte) (string, error) {
	if len(artifacts) == 0 {
		return "", fmt.Errorf("%w: constants artifact bundle is empty", ErrInvalidState)
	}
	names := make([]string, 0, len(artifacts))
	for name := range artifacts {
		if name == "" {
			return "", fmt.Errorf("%w: constants artifact name is empty", ErrInvalidState)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	var frame [8]byte
	for _, name := range names {
		binary.BigEndian.PutUint64(frame[:], uint64(len(name)))
		_, _ = hash.Write(frame[:])
		_, _ = hash.Write([]byte(name))
		data := artifacts[name]
		binary.BigEndian.PutUint64(frame[:], uint64(len(data)))
		_, _ = hash.Write(frame[:])
		_, _ = hash.Write(data)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func EncodeState(state *State) ([]byte, error) {
	if state == nil || state.Ledger == nil {
		return nil, fmt.Errorf("%w: nil state or ledger", ErrInvalidState)
	}
	if state.GeneratorCounts == nil {
		return nil, fmt.Errorf("%w: generators are required", ErrInvalidState)
	}
	for id, count := range state.GeneratorCounts {
		if count < 0 || count > decimal.MaxExactInteger {
			return nil, fmt.Errorf("%w: invalid generator count for %q", ErrInvalidState, id)
		}
	}
	if state.ComputeCreditMS < 0 || state.ComputeCreditMS > decimal.MaxExactInteger ||
		state.ManualTokenMilli < 0 || state.ManualTokenMilli > decimal.MaxExactInteger ||
		state.RouteKnowledgeBalance < 0 || state.RouteKnowledgeBalance > decimal.MaxExactInteger {
		return nil, fmt.Errorf("%w: production integers exceed the exact domain", ErrInvalidState)
	}
	normalized := *state
	if normalized.Ledger.Scope() == economy.ScopeCompany && normalized.RunSeq == 0 {
		normalized.RunSeq = 1
	}
	if err := validateRouteState(&normalized, normalized.Ledger.Scope()); err != nil {
		return nil, err
	}
	if err := validateCompactState(&normalized, normalized.Ledger.Scope()); err != nil {
		return nil, err
	}
	if err := validatePrestigeState(&normalized, normalized.Ledger.Scope()); err != nil {
		return nil, err
	}
	if err := validateFactionState(&normalized, normalized.Ledger.Scope()); err != nil {
		return nil, err
	}
	if err := validateGuildState(&normalized, normalized.Ledger.Scope()); err != nil {
		return nil, err
	}
	cursor, err := formatCursor(state.EvaluatedThrough)
	if err != nil {
		return nil, err
	}
	refilledAt, err := formatCursor(state.ManualTokenRefilledAt)
	if err != nil {
		return nil, fmt.Errorf("%w: manual_token_refilled_at is required", ErrInvalidState)
	}
	if state.ManualTokenRefilledAt.After(state.EvaluatedThrough) {
		return nil, fmt.Errorf("%w: manual_token_refilled_at exceeds evaluated_through", ErrInvalidState)
	}
	encoded, err := json.Marshal(stateV11{stateV10: stateV10{stateV9: stateV9{stateV8: stateV8{stateV7: stateV7{stateV6: stateV6{stateV5: stateV5{
		Balances: state.Ledger.Snapshot(), Generators: state.GeneratorCounts, EvaluatedThrough: cursor,
		ComputeCreditMS: state.ComputeCreditMS, ManualTokenMilli: state.ManualTokenMilli,
		ManualTokenRefilledAt: refilledAt, GatesCrossed: cloneBoolMap(normalized.GatesCrossed), RunSeq: normalized.RunSeq,
		DoctrinesByTransition: cloneStringMap(normalized.DoctrinesByTransition), StructureID: normalized.StructureID,
		LedgerFactKinds: sortedTrueKeys(normalized.LedgerFactKinds), MeterBands: cloneIntMap(normalized.MeterBands),
		RegionTraits: sortedTrueKeys(normalized.RegionTraits), RouteKnowledgeBalance: normalized.RouteKnowledgeBalance,
		HintsUnlocked: sortedTrueKeys(normalized.HintsUnlocked),
	}, CompactMember: normalized.CompactMember, CompactTithePPM: normalized.CompactTithePPM,
		CompactSolidarityPPM: normalized.CompactSolidarityPPM, CompactSamples: encodeCompactSamples(normalized.CompactSamples)},
		Tier: normalized.Tier, LifetimeValue: normalized.LifetimeValue.String(), OfferState: encodeExitOffer(normalized.OfferState),
		RunStartedAtMS: timeToExactMS(normalized.RunStartedAt), OfflineSpans: encodeOfflineSpans(normalized.OfflineSpans),
		ReputationLevel: normalized.ReputationLevel, ReputationUnlockPPM: normalized.ReputationUnlockPPM,
		NetworkSlots: cloneNetworkSlots(normalized.NetworkSlots), CloutLifetime: normalized.CloutLifetime,
		Soul: normalized.Soul, AgeMS: normalized.AgeMS, Notoriety: normalized.Notoriety,
		AdvisorMode: normalized.AdvisorMode, ExitHistory: encodeExitHistory(normalized.ExitHistory)},
		RunPreTimer: normalized.RunPreTimer}, CollapsedOfflineMS: normalized.CollapsedOfflineMS},
		FactionID: optionalString(normalized.FactionID), IncorporatedAtMS: optionalTimeMS(normalized.IncorporatedAt),
		StockUnits: normalized.StockUnits, StockProgressMS: normalized.StockProgressMS, ConsumedStockUnits: normalized.ConsumedStockUnits},
		GuildTitheCarryPPM: normalized.GuildTitheCarryPPM, GuildBoundarySeq: normalized.GuildBoundarySeq,
		GuildConsumedWindow: normalized.GuildConsumedWindow})
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrInvalidState, err)
	}
	return encoded, nil
}

func RestoreState(data []byte, version int, catalog *economy.Catalog, scope economy.Scope, migrationBaseline time.Time) (*State, error) {
	if version > CurrentVersion {
		return nil, fmt.Errorf("%w: save version %d is newer than supported version %d", ErrInvalidState, version, CurrentVersion)
	}
	if version < 1 {
		return nil, fmt.Errorf("%w: unsupported save version %d", ErrInvalidState, version)
	}
	if catalog == nil {
		return nil, fmt.Errorf("%w: nil catalog", ErrInvalidState)
	}

	var source stateV11
	if version == 1 {
		var legacy stateV1
		if err := decodeState(data, &legacy); err != nil {
			return nil, err
		}
		cursor, err := formatCursor(CanonicalServerTime(migrationBaseline))
		if err != nil {
			return nil, fmt.Errorf("%w: version-1 migration baseline: %v", ErrInvalidState, err)
		}
		source.stateV10.stateV9.stateV8.stateV7.stateV6.stateV5 = stateV5{
			Balances: legacy.Balances, Generators: zeroGeneratorCounts(catalog, scope), EvaluatedThrough: cursor,
		}
	} else if version == 2 {
		var previous stateV2
		if err := decodeState(data, &previous); err != nil {
			return nil, err
		}
		source.stateV10.stateV9.stateV8.stateV7.stateV6.stateV5 = stateV5{Balances: previous.Balances, Generators: previous.Generators, EvaluatedThrough: previous.EvaluatedThrough}
	} else if version < 5 {
		var previous stateV4
		if err := decodeState(data, &previous); err != nil {
			return nil, err
		}
		source.stateV10.stateV9.stateV8.stateV7.stateV6.stateV5 = stateV5{
			Balances: previous.Balances, Generators: previous.Generators, EvaluatedThrough: previous.EvaluatedThrough,
			ComputeCreditMS: previous.ComputeCreditMS, ManualTokenMilli: previous.ManualTokenMilli,
			ManualTokenRefilledAt: previous.ManualTokenRefilledAt,
		}
	} else if version == 5 {
		var previous stateV5
		if err := decodeState(data, &previous); err != nil {
			return nil, err
		}
		source.stateV10.stateV9.stateV8.stateV7.stateV6.stateV5 = previous
	} else if version == 6 {
		var previous stateV6
		if err := decodeState(data, &previous); err != nil {
			return nil, err
		}
		source.stateV10.stateV9.stateV8.stateV7.stateV6 = previous
	} else if version == 7 {
		if err := decodeState(data, &source.stateV10.stateV9.stateV8.stateV7); err != nil {
			return nil, err
		}
	} else if version == 8 {
		if err := decodeState(data, &source.stateV10.stateV9.stateV8); err != nil {
			return nil, err
		}
	} else if version == 9 {
		if err := decodeState(data, &source.stateV10.stateV9); err != nil {
			return nil, err
		}
	} else if version == 10 {
		if err := decodeState(data, &source.stateV10); err != nil {
			return nil, err
		}
	} else if err := decodeState(data, &source); err != nil {
		return nil, err
	}
	if version == 5 && (source.GatesCrossed == nil || source.DoctrinesByTransition == nil || source.LedgerFactKinds == nil ||
		source.MeterBands == nil || source.RegionTraits == nil || source.HintsUnlocked == nil) {
		return nil, fmt.Errorf("%w: route state collections are required", ErrInvalidState)
	}
	if version < 5 {
		source.GatesCrossed = map[string]bool{}
		source.DoctrinesByTransition = map[string]string{}
		source.LedgerFactKinds = []string{}
		source.MeterBands = map[string]int{}
		source.RegionTraits = []string{}
		source.HintsUnlocked = []string{}
		if scope == economy.ScopeCompany {
			source.RunSeq = 1
		}
	}
	if version == 6 && source.CompactSamples == nil {
		return nil, fmt.Errorf("%w: compact_solidarity_samples are required", ErrInvalidState)
	}
	if version < 6 {
		source.CompactSamples = []rawCompactSample{}
	}
	if version < 7 {
		source.LifetimeValue = decimal.Zero.String()
		source.OfflineSpans = []rawOfflineSpan{}
		source.NetworkSlots = []NetworkSlot{}
		source.ExitHistory = []rawExitRecord{}
	}
	if version == 7 && (source.OfflineSpans == nil || source.NetworkSlots == nil || source.ExitHistory == nil) {
		return nil, fmt.Errorf("%w: prestige state collections are required", ErrInvalidState)
	}

	if source.Balances == nil || source.Generators == nil {
		return nil, fmt.Errorf("%w: balances and generators are required", ErrInvalidState)
	}
	ledger, err := economy.RestoreLedger(catalog, scope, source.Balances)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidState, err)
	}
	counts, err := validateGeneratorCounts(catalog, scope, source.Generators)
	if err != nil {
		return nil, err
	}
	cursor, err := restoreCursor(source.EvaluatedThrough, version)
	if err != nil {
		return nil, err
	}
	if version < 3 {
		source.ComputeCreditMS = 0
		source.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
		source.ManualTokenRefilledAt = source.EvaluatedThrough
	}
	refilledAt, err := restoreCursor(source.ManualTokenRefilledAt, version)
	if err != nil {
		return nil, fmt.Errorf("%w: manual_token_refilled_at: %v", ErrInvalidState, err)
	}
	if refilledAt.After(cursor) {
		return nil, fmt.Errorf("%w: manual_token_refilled_at exceeds evaluated_through", ErrInvalidState)
	}
	if err := validateProductionState(catalog, scope, source.ComputeCreditMS, source.ManualTokenMilli); err != nil {
		return nil, err
	}
	ledgerFacts, err := uniqueMechanicalKeys(source.LedgerFactKinds, "ledger_fact_kinds")
	if err != nil {
		return nil, err
	}
	regionTraits, err := uniqueMechanicalKeys(source.RegionTraits, "region_traits")
	if err != nil {
		return nil, err
	}
	hints, err := uniqueMechanicalKeys(source.HintsUnlocked, "hints_unlocked")
	if err != nil {
		return nil, err
	}
	state := &State{
		Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: cursor,
		ComputeCreditMS: source.ComputeCreditMS, ManualTokenMilli: source.ManualTokenMilli,
		ManualTokenRefilledAt: refilledAt, GatesCrossed: source.GatesCrossed, RunSeq: source.RunSeq,
		DoctrinesByTransition: source.DoctrinesByTransition, StructureID: source.StructureID,
		LedgerFactKinds: ledgerFacts, MeterBands: source.MeterBands,
		RegionTraits: regionTraits, RouteKnowledgeBalance: source.RouteKnowledgeBalance,
		HintsUnlocked: hints,
		CompactMember: source.CompactMember, CompactTithePPM: source.CompactTithePPM,
		CompactSolidarityPPM: source.CompactSolidarityPPM,
		Tier:                 source.Tier, ReputationLevel: source.ReputationLevel,
		ReputationUnlockPPM: source.ReputationUnlockPPM, NetworkSlots: cloneNetworkSlots(source.NetworkSlots),
		CloutLifetime: source.CloutLifetime, Soul: source.Soul, AgeMS: source.AgeMS,
		Notoriety: source.Notoriety, AdvisorMode: source.AdvisorMode,
		StockUnits: source.StockUnits, StockProgressMS: source.StockProgressMS,
		ConsumedStockUnits: source.ConsumedStockUnits,
		GuildTitheCarryPPM: source.GuildTitheCarryPPM, GuildBoundarySeq: source.GuildBoundarySeq,
		GuildConsumedWindow: source.GuildConsumedWindow,
	}
	if source.FactionID != nil {
		state.FactionID = *source.FactionID
	}
	if source.IncorporatedAtMS != nil {
		state.IncorporatedAt, err = exactMSToTime(*source.IncorporatedAtMS)
		if err != nil || state.IncorporatedAt.IsZero() {
			return nil, fmt.Errorf("%w: incorporated_at_ms", ErrInvalidState)
		}
	}
	state.LifetimeValue, err = decimal.ParseCanonical(source.LifetimeValue)
	if err != nil {
		return nil, fmt.Errorf("%w: lifetime_value", ErrInvalidState)
	}
	state.OfferState, err = decodeExitOffer(source.OfferState)
	if err != nil {
		return nil, err
	}
	state.RunStartedAt, err = exactMSToTime(source.RunStartedAtMS)
	if err != nil {
		return nil, err
	}
	if version < 7 && scope == economy.ScopeCompany {
		state.RunStartedAt = cursor
		state.RunPreTimer = true
	} else {
		state.RunPreTimer = source.RunPreTimer
	}
	state.OfflineSpans, err = decodeOfflineSpans(source.OfflineSpans)
	if err != nil {
		return nil, err
	}
	state.CollapsedOfflineMS = source.CollapsedOfflineMS
	state.ExitHistory, err = decodeExitHistory(source.ExitHistory)
	if err != nil {
		return nil, err
	}
	state.CompactSamples, err = decodeCompactSamples(source.CompactSamples)
	if err != nil {
		return nil, err
	}
	if err := validateRouteState(state, scope); err != nil {
		return nil, err
	}
	if err := validateCompactState(state, scope); err != nil {
		return nil, err
	}
	if err := validatePrestigeState(state, scope); err != nil {
		return nil, err
	}
	if err := validateFactionState(state, scope); err != nil {
		return nil, err
	}
	if err := validateGuildState(state, scope); err != nil {
		return nil, err
	}
	return state, nil
}

func validateFactionState(state *State, scope economy.Scope) error {
	if state.StockUnits < 0 || state.StockUnits > decimal.MaxExactInteger ||
		state.StockProgressMS < 0 || state.StockProgressMS > decimal.MaxExactInteger ||
		state.ConsumedStockUnits < 0 || state.ConsumedStockUnits > decimal.MaxExactInteger {
		return fmt.Errorf("%w: faction stock values outside their exact domains", ErrInvalidState)
	}
	if scope != economy.ScopeCompany {
		if state.FactionID != "" || !state.IncorporatedAt.IsZero() || state.StockUnits != 0 || state.StockProgressMS != 0 || state.ConsumedStockUnits != 0 {
			return fmt.Errorf("%w: company faction state leaked outside company scope", ErrInvalidState)
		}
		return nil
	}
	if state.FactionID == "" {
		if !state.IncorporatedAt.IsZero() || state.StockUnits != 0 || state.StockProgressMS != 0 || state.ConsumedStockUnits != 0 {
			return fmt.Errorf("%w: stock state exists before incorporation", ErrInvalidState)
		}
		return nil
	}
	if !stateMechanicalIDPattern.MatchString(state.FactionID) || state.IncorporatedAt.IsZero() ||
		!isCanonicalMillisecond(state.IncorporatedAt) || state.IncorporatedAt.After(state.EvaluatedThrough) {
		return fmt.Errorf("%w: invalid faction incorporation", ErrInvalidState)
	}
	return nil
}

func validateGuildState(state *State, scope economy.Scope) error {
	if state.GuildTitheCarryPPM < 0 || state.GuildTitheCarryPPM >= 1_000_000 ||
		state.GuildBoundarySeq < 0 || state.GuildBoundarySeq > decimal.MaxExactInteger ||
		state.GuildConsumedWindow < 0 || state.GuildConsumedWindow > decimal.MaxExactInteger {
		return fmt.Errorf("%w: guild values outside their exact domains", ErrInvalidState)
	}
	if scope != economy.ScopeCompany && (state.GuildTitheCarryPPM != 0 || state.GuildBoundarySeq != 0 || state.GuildConsumedWindow != 0) {
		return fmt.Errorf("%w: company guild state leaked outside company scope", ErrInvalidState)
	}
	return nil
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	copyValue := value
	return &copyValue
}

func optionalTimeMS(value time.Time) *int64 {
	if value.IsZero() {
		return nil
	}
	milliseconds := value.UnixMilli()
	return &milliseconds
}

func validatePrestigeState(state *State, scope economy.Scope) error {
	if state.Tier < 0 || state.Tier > 9 || !state.LifetimeValue.IsStateValue() || state.LifetimeValue.Lt(decimal.Zero) ||
		state.ReputationLevel < 0 || state.ReputationLevel > decimal.MaxExactInteger ||
		state.ReputationUnlockPPM < 0 || state.ReputationUnlockPPM > 1_000_000 ||
		state.CloutLifetime < 0 || state.CloutLifetime > decimal.MaxExactInteger ||
		state.Soul < -decimal.MaxExactInteger || state.Soul > decimal.MaxExactInteger ||
		state.AgeMS < 0 || state.AgeMS > decimal.MaxExactInteger ||
		state.Notoriety < 0 || state.Notoriety > decimal.MaxExactInteger ||
		state.CollapsedOfflineMS < 0 || state.CollapsedOfflineMS > decimal.MaxExactInteger {
		return fmt.Errorf("%w: prestige values outside their exact domains", ErrInvalidState)
	}
	if len(state.OfflineSpans) > maxOfflineSpans {
		return fmt.Errorf("%w: too many offline spans", ErrInvalidState)
	}
	if scope == economy.ScopeCompany {
		if state.ReputationLevel != 0 || state.ReputationUnlockPPM != 0 || len(state.NetworkSlots) != 0 ||
			state.CloutLifetime != 0 || state.Soul != 0 || state.AgeMS != 0 || state.Notoriety != 0 ||
			state.AdvisorMode || len(state.ExitHistory) != 0 {
			return fmt.Errorf("%w: founder prestige state leaked into company scope", ErrInvalidState)
		}
		if state.RunPreTimer && state.RunStartedAt.IsZero() || !state.RunStartedAt.IsZero() && (!isCanonicalMillisecond(state.RunStartedAt) || state.RunStartedAt.After(state.EvaluatedThrough)) {
			return fmt.Errorf("%w: invalid run_started_at", ErrInvalidState)
		}
		if state.RunStartedAt.IsZero() && state.CollapsedOfflineMS != 0 {
			return fmt.Errorf("%w: collapsed offline time without a run start", ErrInvalidState)
		}
		last := time.Time{}
		totalOfflineMS := state.CollapsedOfflineMS
		for _, span := range state.OfflineSpans {
			if span.From.IsZero() || !isCanonicalMillisecond(span.From) || !isCanonicalMillisecond(span.To) || !span.To.After(span.From) ||
				!last.IsZero() && span.From.Before(last) || !state.RunStartedAt.IsZero() && span.From.Before(state.RunStartedAt) || span.To.After(state.EvaluatedThrough) {
				return fmt.Errorf("%w: invalid offline span", ErrInvalidState)
			}
			duration := span.To.Sub(span.From).Milliseconds()
			if duration > decimal.MaxExactInteger-totalOfflineMS {
				return fmt.Errorf("%w: offline duration exceeds the exact domain", ErrInvalidState)
			}
			totalOfflineMS += duration
			last = span.To
		}
		if !state.RunStartedAt.IsZero() && totalOfflineMS > state.EvaluatedThrough.Sub(state.RunStartedAt).Milliseconds() {
			return fmt.Errorf("%w: offline duration exceeds run duration", ErrInvalidState)
		}
		if state.OfferState != nil {
			offer := state.OfferState
			if !uuidV7Pattern.MatchString(offer.OfferID) || !validExitType(offer.ExitType) || len(offer.TermsJSON) == 0 ||
				!json.Valid(offer.TermsJSON) || !isCanonicalMillisecond(offer.SpawnedAt) || !isCanonicalMillisecond(offer.ExpiresAt) || !offer.ExpiresAt.After(offer.SpawnedAt) {
				return fmt.Errorf("%w: invalid exit offer", ErrInvalidState)
			}
		}
		return nil
	}
	if state.Tier != 0 || !state.LifetimeValue.Eq(decimal.Zero) || state.OfferState != nil || !state.RunStartedAt.IsZero() || state.RunPreTimer || len(state.OfflineSpans) != 0 || state.CollapsedOfflineMS != 0 {
		return fmt.Errorf("%w: company prestige state leaked outside company scope", ErrInvalidState)
	}
	if scope != economy.ScopeFounder {
		if state.ReputationLevel != 0 || state.ReputationUnlockPPM != 0 || len(state.NetworkSlots) != 0 || state.CloutLifetime != 0 || state.Soul != 0 || state.AgeMS != 0 || state.Notoriety != 0 || state.AdvisorMode || len(state.ExitHistory) != 0 {
			return fmt.Errorf("%w: founder prestige state leaked outside founder scope", ErrInvalidState)
		}
		return nil
	}
	seenSlots := map[string]bool{}
	lastSlot := ""
	for _, slot := range state.NetworkSlots {
		if !stateMechanicalIDPattern.MatchString(slot.Slot) || !stateMechanicalIDPattern.MatchString(slot.CarriedRef) || seenSlots[slot.Slot] || slot.Slot < lastSlot {
			return fmt.Errorf("%w: invalid network slot", ErrInvalidState)
		}
		seenSlots[slot.Slot], lastSlot = true, slot.Slot
	}
	var lastRun int64
	for _, record := range state.ExitHistory {
		if record.RunID <= lastRun || record.RunID > decimal.MaxExactInteger || !validExitType(record.ExitType) ||
			record.OccurredAt.IsZero() || !isCanonicalMillisecond(record.OccurredAt) || record.ReputationDelta < 0 || record.ReputationDelta > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid exit history", ErrInvalidState)
		}
		lastRun = record.RunID
	}
	return nil
}

func validExitType(value string) bool {
	switch value {
	case "acquihire", "acquisition", "ipo", "collapse", "scripted_first":
		return true
	default:
		return false
	}
}

func isCanonicalMillisecond(value time.Time) bool {
	return value.Location() == time.UTC && value.Nanosecond()%int(time.Millisecond) == 0
}

func timeToExactMS(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

func exactMSToTime(value int64) (time.Time, error) {
	if value == 0 {
		return time.Time{}, nil
	}
	if value < 0 || value > decimal.MaxExactInteger {
		return time.Time{}, fmt.Errorf("%w: millisecond timestamp outside exact domain", ErrInvalidState)
	}
	return time.UnixMilli(value).UTC(), nil
}

func encodeExitOffer(offer *ExitOfferState) *rawExitOfferState {
	if offer == nil {
		return nil
	}
	return &rawExitOfferState{OfferID: offer.OfferID, ExitType: offer.ExitType, TermsJSON: cloneRaw(offer.TermsJSON), SpawnedAtMS: timeToExactMS(offer.SpawnedAt), ExpiresAtMS: timeToExactMS(offer.ExpiresAt)}
}

func decodeExitOffer(offer *rawExitOfferState) (*ExitOfferState, error) {
	if offer == nil {
		return nil, nil
	}
	spawned, err := exactMSToTime(offer.SpawnedAtMS)
	if err != nil {
		return nil, err
	}
	expires, err := exactMSToTime(offer.ExpiresAtMS)
	if err != nil {
		return nil, err
	}
	return &ExitOfferState{OfferID: offer.OfferID, ExitType: offer.ExitType, TermsJSON: cloneRaw(offer.TermsJSON), SpawnedAt: spawned, ExpiresAt: expires}, nil
}

func encodeOfflineSpans(spans []OfflineSpan) []rawOfflineSpan {
	result := make([]rawOfflineSpan, len(spans))
	for index, span := range spans {
		result[index] = rawOfflineSpan{FromMS: timeToExactMS(span.From), ToMS: timeToExactMS(span.To)}
	}
	return result
}

func decodeOfflineSpans(spans []rawOfflineSpan) ([]OfflineSpan, error) {
	result := make([]OfflineSpan, len(spans))
	for index, span := range spans {
		from, err := exactMSToTime(span.FromMS)
		if err != nil {
			return nil, err
		}
		to, err := exactMSToTime(span.ToMS)
		if err != nil {
			return nil, err
		}
		result[index] = OfflineSpan{From: from, To: to}
	}
	return result, nil
}

func cloneNetworkSlots(slots []NetworkSlot) []NetworkSlot {
	result := make([]NetworkSlot, len(slots))
	copy(result, slots)
	return result
}

func encodeExitHistory(records []ExitRecord) []rawExitRecord {
	result := make([]rawExitRecord, len(records))
	for index, record := range records {
		result[index] = rawExitRecord{RunID: record.RunID, ExitType: record.ExitType, OccurredAtMS: timeToExactMS(record.OccurredAt), ReputationDelta: record.ReputationDelta}
	}
	return result
}

func decodeExitHistory(records []rawExitRecord) ([]ExitRecord, error) {
	result := make([]ExitRecord, len(records))
	for index, record := range records {
		occurred, err := exactMSToTime(record.OccurredAtMS)
		if err != nil {
			return nil, err
		}
		result[index] = ExitRecord{RunID: record.RunID, ExitType: record.ExitType, OccurredAt: occurred, ReputationDelta: record.ReputationDelta}
	}
	return result, nil
}

func validateCompactState(state *State, scope economy.Scope) error {
	if state.CompactTithePPM < 0 || state.CompactTithePPM > 1_000_000 || state.CompactSolidarityPPM < 0 || state.CompactSolidarityPPM > 1_000_000 {
		return fmt.Errorf("%w: compact ratios outside ppm domain", ErrInvalidState)
	}
	if scope != economy.ScopeCompany {
		if state.CompactMember || state.CompactTithePPM != 0 || state.CompactSolidarityPPM != 0 || len(state.CompactSamples) != 0 {
			return fmt.Errorf("%w: compact membership is company-scoped", ErrInvalidState)
		}
		return nil
	}
	if !state.CompactMember && (state.CompactTithePPM != 0 || state.CompactSolidarityPPM != 0 || len(state.CompactSamples) != 0) {
		return fmt.Errorf("%w: non-member has compact state", ErrInvalidState)
	}
	last := time.Time{}
	for _, sample := range state.CompactSamples {
		if sample.HourStart.Location() != time.UTC || sample.HourStart.Minute() != 0 || sample.HourStart.Second() != 0 || sample.HourStart.Nanosecond() != 0 || sample.CompliancePPM < 0 || sample.CompliancePPM > 1_000_000 || sample.CoveredMS <= 0 || sample.CoveredMS > int64(time.Hour/time.Millisecond) || !last.IsZero() && !sample.HourStart.After(last) {
			return fmt.Errorf("%w: invalid compact solidarity sample", ErrInvalidState)
		}
		last = sample.HourStart
	}
	return nil
}

func encodeCompactSamples(samples []CompactSample) []rawCompactSample {
	result := make([]rawCompactSample, len(samples))
	for index, sample := range samples {
		result[index] = rawCompactSample{HourStart: sample.HourStart.Format(time.RFC3339Nano), CompliancePPM: sample.CompliancePPM, CoveredMS: sample.CoveredMS}
	}
	return result
}

func decodeCompactSamples(samples []rawCompactSample) ([]CompactSample, error) {
	result := make([]CompactSample, len(samples))
	for index, sample := range samples {
		value, err := time.Parse(time.RFC3339Nano, sample.HourStart)
		if err != nil || value.Location() != time.UTC || value.Format(time.RFC3339Nano) != sample.HourStart {
			return nil, fmt.Errorf("%w: invalid compact sample time", ErrInvalidState)
		}
		result[index] = CompactSample{HourStart: value, CompliancePPM: sample.CompliancePPM, CoveredMS: sample.CoveredMS}
	}
	return result, nil
}

func validateRouteState(state *State, scope economy.Scope) error {
	if state.RunSeq < 0 || state.RunSeq > decimal.MaxExactInteger || state.RouteKnowledgeBalance < 0 || state.RouteKnowledgeBalance > decimal.MaxExactInteger {
		return fmt.Errorf("%w: route integers exceed the exact domain", ErrInvalidState)
	}
	for id, crossed := range state.GatesCrossed {
		if !stateMechanicalIDPattern.MatchString(id) || !crossed {
			return fmt.Errorf("%w: invalid gates_crossed entry", ErrInvalidState)
		}
	}
	for transition, doctrine := range state.DoctrinesByTransition {
		if !stateMechanicalIDPattern.MatchString(transition) || !stateMechanicalIDPattern.MatchString(doctrine) {
			return fmt.Errorf("%w: invalid doctrine entry", ErrInvalidState)
		}
	}
	if state.StructureID != "" && !stateMechanicalIDPattern.MatchString(state.StructureID) {
		return fmt.Errorf("%w: invalid structure_id", ErrInvalidState)
	}
	for id, present := range state.LedgerFactKinds {
		if !stateMechanicalIDPattern.MatchString(id) || !present {
			return fmt.Errorf("%w: invalid ledger fact", ErrInvalidState)
		}
	}
	for id, value := range state.MeterBands {
		if !stateMechanicalIDPattern.MatchString(id) || value < 0 || value > 100 {
			return fmt.Errorf("%w: invalid meter band", ErrInvalidState)
		}
	}
	for id, present := range state.RegionTraits {
		if !stateMechanicalIDPattern.MatchString(id) || !present {
			return fmt.Errorf("%w: invalid region trait", ErrInvalidState)
		}
	}
	for id, unlocked := range state.HintsUnlocked {
		if !stateMechanicalIDPattern.MatchString(id) || !unlocked {
			return fmt.Errorf("%w: invalid hint", ErrInvalidState)
		}
	}
	if scope == economy.ScopeCompany {
		if state.RunSeq < 1 || state.RouteKnowledgeBalance != 0 || len(state.HintsUnlocked) != 0 {
			return fmt.Errorf("%w: invalid company route state", ErrInvalidState)
		}
	} else if scope == economy.ScopeFounder {
		if state.RunSeq != 0 || len(state.GatesCrossed) != 0 || len(state.DoctrinesByTransition) != 0 || state.StructureID != "" || len(state.MeterBands) != 0 || len(state.RegionTraits) != 0 {
			return fmt.Errorf("%w: company route context leaked outside company scope", ErrInvalidState)
		}
	} else if state.RunSeq != 0 || state.RouteKnowledgeBalance != 0 || len(state.GatesCrossed) != 0 || len(state.HintsUnlocked) != 0 {
		return fmt.Errorf("%w: route state is company/founder scoped", ErrInvalidState)
	}
	return nil
}

func sortedTrueKeys(source map[string]bool) []string {
	keys := make([]string, 0, len(source))
	for key, present := range source {
		if present {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func uniqueMechanicalKeys(keys []string, label string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	for _, key := range keys {
		if !stateMechanicalIDPattern.MatchString(key) || result[key] {
			return nil, fmt.Errorf("%w: invalid or duplicate %s", ErrInvalidState, label)
		}
		result[key] = true
	}
	return result, nil
}

func cloneBoolMap(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
func cloneStringMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
func cloneIntMap(source map[string]int) map[string]int {
	result := make(map[string]int, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func validateProductionState(catalog *economy.Catalog, scope economy.Scope, creditMS, tokenMilli int64) error {
	if creditMS < 0 || creditMS > decimal.MaxExactInteger || tokenMilli < 0 || tokenMilli > decimal.MaxExactInteger {
		return fmt.Errorf("%w: production integers exceed the exact domain", ErrInvalidState)
	}
	if scope != economy.ScopeCompany {
		if creditMS != 0 || tokenMilli != 0 {
			return fmt.Errorf("%w: production state is company-scoped", ErrInvalidState)
		}
		return nil
	}
	offlinePolicy, manualPolicy := catalog.OfflinePolicy(), catalog.ManualPolicy()
	if offlinePolicy.BankCapMS == 0 && creditMS != 0 || offlinePolicy.BankCapMS > 0 && creditMS > offlinePolicy.BankCapMS {
		return fmt.Errorf("%w: compute_credit_ms exceeds catalog cap", ErrInvalidState)
	}
	if manualPolicy.BucketCapMilli == 0 && tokenMilli != 0 || manualPolicy.BucketCapMilli > 0 && tokenMilli > manualPolicy.BucketCapMilli {
		return fmt.Errorf("%w: manual_token_milli exceeds catalog cap", ErrInvalidState)
	}
	return nil
}

func validateGeneratorCounts(catalog *economy.Catalog, scope economy.Scope, source map[string]int64) (map[string]int64, error) {
	expected := catalog.GeneratorClassesForScope(scope)
	if len(source) != len(expected) {
		return nil, fmt.Errorf("%w: generators contain %d entries, want %d", ErrInvalidState, len(source), len(expected))
	}
	counts := make(map[string]int64, len(expected))
	for _, definition := range expected {
		count, ok := source[definition.ID]
		if !ok || count < 0 || count > decimal.MaxExactInteger {
			return nil, fmt.Errorf("%w: invalid generator count for %q", ErrInvalidState, definition.ID)
		}
		counts[definition.ID] = count
	}
	return counts, nil
}

func zeroGeneratorCounts(catalog *economy.Catalog, scope economy.Scope) map[string]int64 {
	counts := make(map[string]int64)
	for _, definition := range catalog.GeneratorClassesForScope(scope) {
		counts[definition.ID] = 0
	}
	return counts
}

// CanonicalServerTime returns the only timestamp representation permitted for
// authoritative production cursors: UTC truncated to an exact millisecond.
func CanonicalServerTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Millisecond)
}

func formatCursor(value time.Time) (string, error) {
	if value.IsZero() {
		return "", fmt.Errorf("%w: evaluated_through is required", ErrInvalidState)
	}
	if !isCanonicalServerTime(value) {
		return "", fmt.Errorf("%w: production cursor must be canonical UTC whole milliseconds", ErrInvalidState)
	}
	return value.Format(time.RFC3339Nano), nil
}

func parseCursor(source string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339Nano, source)
	if err != nil || value.Location() != time.UTC || value.Format(time.RFC3339Nano) != source {
		return time.Time{}, fmt.Errorf("%w: evaluated_through must be canonical UTC RFC3339Nano", ErrInvalidState)
	}
	return value, nil
}

func restoreCursor(source string, version int) (time.Time, error) {
	value, err := parseCursor(source)
	if err != nil {
		return time.Time{}, err
	}
	if version < millisecondCursorVersion {
		return CanonicalServerTime(value), nil
	}
	if !isCanonicalServerTime(value) {
		return time.Time{}, fmt.Errorf("%w: production cursor must be canonical UTC whole milliseconds", ErrInvalidState)
	}
	return value, nil
}

func isCanonicalServerTime(value time.Time) bool {
	return value.Location() == time.UTC && value.Nanosecond()%int(time.Millisecond) == 0
}

func decodeState(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("%w: decode: %v", ErrInvalidState, err)
	}
	return ensureJSONEnd(decoder)
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: trailing JSON value", ErrInvalidState)
		}
		return fmt.Errorf("%w: trailing data: %v", ErrInvalidState, err)
	}
	return nil
}
