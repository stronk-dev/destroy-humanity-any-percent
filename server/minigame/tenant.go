package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

var (
	ErrInvalidTenant    = errors.New("invalid minigame tenant")
	ErrTenantRejected   = errors.New("minigame command rejected")
	ErrUnknownTenant    = errors.New("unknown minigame tenant")
	ErrTenantDivergence = errors.New("minigame tenant returned invalid output")
)

type DestinationClass string

const (
	DestinationPower        DestinationClass = "power"
	DestinationBreadth      DestinationClass = "breadth"
	DestinationPresentation DestinationClass = "presentation"
)

type Descriptor struct {
	EngineRef      string
	EngineVersion  string
	CommandSchema  string
	SnapshotSchema string
	ResultSchema   string
	Modes          []Mode
	ErrorTaxonomy  []string
	Destinations   map[string]DestinationClass
}

type CreateInput struct {
	Mode          Mode
	Seed          uint64
	ScalingInputs map[string]int64
}

type ApplyInput struct {
	Mode          Mode
	Revision      int64
	Snapshot      json.RawMessage
	Command       json.RawMessage
	ScalingInputs map[string]int64
}

type ScoreFact struct {
	Kind  string `json:"kind"`
	Value int64  `json:"value"`
}

type Result struct {
	Outcome     string      `json:"outcome"`
	ScoreFacts  []ScoreFact `json:"score_facts"`
	RatingDelta *int64      `json:"rating_delta"`
}

type ApplyOutput struct {
	Snapshot json.RawMessage
	Result   *Result
}

type Rejection struct {
	Code   string
	Detail string
}

func (rejection *Rejection) Error() string {
	if rejection == nil {
		return ErrTenantRejected.Error()
	}
	return rejection.Code + ": " + rejection.Detail
}

func (rejection *Rejection) Unwrap() error { return ErrTenantRejected }

// Tenant is deliberately pure: every value an engine may read is an argument,
// and every authoritative effect is returned. Economy payout is not part of
// this interface; the platform derives it from a validated Result.
type Tenant interface {
	Descriptor() Descriptor
	Create(CreateInput) (json.RawMessage, error)
	Apply(ApplyInput) (ApplyOutput, error)
}

type TenantRegistry struct {
	byEngine map[string]registeredTenant
}

type registeredTenant struct {
	engine     Tenant
	descriptor Descriptor
}

func NewTenantRegistry(tenants ...Tenant) (*TenantRegistry, error) {
	registry := &TenantRegistry{byEngine: map[string]registeredTenant{}}
	for _, tenant := range tenants {
		if tenant == nil {
			return nil, ErrInvalidTenant
		}
		descriptor := tenant.Descriptor()
		if err := validateDescriptor(descriptor); err != nil {
			return nil, err
		}
		if _, exists := registry.byEngine[descriptor.EngineRef]; exists {
			return nil, fmt.Errorf("%w: duplicate engine_ref %s", ErrInvalidTenant, descriptor.EngineRef)
		}
		registry.byEngine[descriptor.EngineRef] = registeredTenant{engine: tenant, descriptor: cloneDescriptor(descriptor)}
	}
	if len(registry.byEngine) == 0 {
		return nil, ErrInvalidTenant
	}
	return registry, nil
}

func (registry *TenantRegistry) Descriptor(engineRef string) (Descriptor, bool) {
	if registry == nil {
		return Descriptor{}, false
	}
	registered, ok := registry.byEngine[engineRef]
	if !ok {
		return Descriptor{}, false
	}
	return cloneDescriptor(registered.descriptor), true
}

func (registry *TenantRegistry) Create(engineRef string, input CreateInput) (json.RawMessage, error) {
	tenant, descriptor, err := registry.resolve(engineRef, input.Mode)
	if err != nil {
		return nil, err
	}
	if !validScalingInputs(input.ScalingInputs, descriptor.Destinations) {
		return nil, ErrInvalidTenant
	}
	snapshot, tenantErr := tenant.Create(cloneCreateInput(input))
	if err := validateTenantError(tenantErr, descriptor); err != nil {
		return nil, err
	}
	canonical, ok := canonicalJSONObject(snapshot)
	if !ok || !bytes.Equal(canonical, snapshot) {
		return nil, ErrTenantDivergence
	}
	return canonical, nil
}

func (registry *TenantRegistry) Apply(engineRef string, input ApplyInput) (ApplyOutput, error) {
	tenant, descriptor, err := registry.resolve(engineRef, input.Mode)
	if err != nil {
		return ApplyOutput{}, err
	}
	canonicalSnapshot, snapshotOK := canonicalJSONObject(input.Snapshot)
	canonicalCommand, commandOK := canonicalJSONObject(input.Command)
	if input.Revision < 1 || !snapshotOK || !commandOK || !bytes.Equal(canonicalSnapshot, input.Snapshot) ||
		!bytes.Equal(canonicalCommand, input.Command) || !validScalingInputs(input.ScalingInputs, descriptor.Destinations) {
		return ApplyOutput{}, ErrInvalidTenant
	}
	input.Snapshot = bytes.Clone(canonicalSnapshot)
	input.Command = bytes.Clone(canonicalCommand)
	input.ScalingInputs = cloneScaling(input.ScalingInputs)
	output, tenantErr := tenant.Apply(input)
	if err := validateTenantError(tenantErr, descriptor); err != nil {
		return ApplyOutput{}, err
	}
	canonicalOutput, ok := canonicalJSONObject(output.Snapshot)
	if !ok || !bytes.Equal(canonicalOutput, output.Snapshot) || !validResult(output.Result) {
		return ApplyOutput{}, ErrTenantDivergence
	}
	output.Snapshot = canonicalOutput
	output.Result = cloneResult(output.Result)
	return output, nil
}

func (registry *TenantRegistry) resolve(engineRef string, mode Mode) (Tenant, Descriptor, error) {
	if registry == nil || !mechanicalPattern.MatchString(engineRef) {
		return nil, Descriptor{}, ErrUnknownTenant
	}
	registered, ok := registry.byEngine[engineRef]
	if !ok {
		return nil, Descriptor{}, ErrUnknownTenant
	}
	for _, allowed := range registered.descriptor.Modes {
		if mode == allowed {
			return registered.engine, registered.descriptor, nil
		}
	}
	return nil, Descriptor{}, ErrInvalidTenant
}

func validateDescriptor(value Descriptor) error {
	if !mechanicalPattern.MatchString(value.EngineRef) || !versionPattern.MatchString(value.EngineVersion) ||
		!mechanicalPattern.MatchString(value.CommandSchema) || !mechanicalPattern.MatchString(value.SnapshotSchema) ||
		!mechanicalPattern.MatchString(value.ResultSchema) || len(value.Modes) == 0 || len(value.ErrorTaxonomy) == 0 ||
		len(value.Destinations) == 0 {
		return ErrInvalidTenant
	}
	seenModes := map[Mode]bool{}
	for _, mode := range value.Modes {
		if mode != ModeSolo && mode != ModeAsyncSnapshot || seenModes[mode] {
			return ErrInvalidTenant
		}
		seenModes[mode] = true
	}
	if !sort.StringsAreSorted(value.ErrorTaxonomy) {
		return ErrInvalidTenant
	}
	for index, code := range value.ErrorTaxonomy {
		if !mechanicalPattern.MatchString(code) || index > 0 && code == value.ErrorTaxonomy[index-1] {
			return ErrInvalidTenant
		}
	}
	for destination, class := range value.Destinations {
		if !mechanicalPattern.MatchString(destination) ||
			class != DestinationPower && class != DestinationBreadth && class != DestinationPresentation {
			return ErrInvalidTenant
		}
	}
	return nil
}

func validateTenantError(err error, descriptor Descriptor) error {
	if err == nil {
		return nil
	}
	var rejection *Rejection
	if !errors.As(err, &rejection) || rejection == nil || !mechanicalPattern.MatchString(rejection.Code) ||
		rejection.Detail == "" {
		return ErrTenantDivergence
	}
	for _, code := range descriptor.ErrorTaxonomy {
		if rejection.Code == code {
			return rejection
		}
	}
	return ErrTenantDivergence
}

func validScalingInputs(values map[string]int64, destinations map[string]DestinationClass) bool {
	if values == nil || len(values) != len(destinations) {
		return false
	}
	for key, value := range values {
		if _, ok := destinations[key]; !ok || value < -9_007_199_254_740_991 || value > 9_007_199_254_740_991 {
			return false
		}
	}
	return true
}

func validResult(value *Result) bool {
	if value == nil {
		return true
	}
	if !mechanicalPattern.MatchString(value.Outcome) || value.ScoreFacts == nil {
		return false
	}
	prior := ""
	for _, fact := range value.ScoreFacts {
		if !mechanicalPattern.MatchString(fact.Kind) || fact.Kind <= prior ||
			fact.Value < -9_007_199_254_740_991 || fact.Value > 9_007_199_254_740_991 {
			return false
		}
		prior = fact.Kind
	}
	return value.RatingDelta == nil ||
		*value.RatingDelta >= -9_007_199_254_740_991 && *value.RatingDelta <= 9_007_199_254_740_991
}

func cloneDescriptor(value Descriptor) Descriptor {
	value.Modes = append([]Mode(nil), value.Modes...)
	value.ErrorTaxonomy = append([]string(nil), value.ErrorTaxonomy...)
	destinations := make(map[string]DestinationClass, len(value.Destinations))
	for key, class := range value.Destinations {
		destinations[key] = class
	}
	value.Destinations = destinations
	return value
}

func cloneCreateInput(value CreateInput) CreateInput {
	value.ScalingInputs = cloneScaling(value.ScalingInputs)
	return value
}

func cloneScaling(value map[string]int64) map[string]int64 {
	result := make(map[string]int64, len(value))
	for key, amount := range value {
		result[key] = amount
	}
	return result
}

func cloneResult(value *Result) *Result {
	if value == nil {
		return nil
	}
	result := *value
	result.ScoreFacts = append([]ScoreFact(nil), value.ScoreFacts...)
	if value.RatingDelta != nil {
		delta := *value.RatingDelta
		result.RatingDelta = &delta
	}
	return &result
}
