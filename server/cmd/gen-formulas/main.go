package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/multiplier"
)

type formulaArtifact struct {
	SchemaVersion       int               `json:"schema_version"`
	ProductionRate      string            `json:"production_rate"`
	MultiplierSlotOrder []multiplier.Slot `json:"multiplier_slot_order"`
	WithinSlotOrder     string            `json:"within_slot_order"`
	SourceFingerprint   string            `json:"source_fingerprint"`
	Commons             commonsFormula    `json:"commons"`
}

type commonsFormula struct {
	Enclosure                string                 `json:"enclosure"`
	Compliance               string                 `json:"compliance"`
	Health                   string                 `json:"health"`
	EffectiveHealth          string                 `json:"effective_health"`
	Modifier                 string                 `json:"modifier"`
	Solidarity               string                 `json:"solidarity"`
	SourceWeights            []commons.SourceWeight `json:"source_weights"`
	DefaultTithePPM          int64                  `json:"default_tithe_ppm"`
	MinimumTithePPM          int64                  `json:"minimum_tithe_ppm"`
	MaximumTithePPM          int64                  `json:"maximum_tithe_ppm"`
	GuildHealthWeightPPM     int64                  `json:"guild_health_weight_ppm"`
	CohortHealthWeightPPM    int64                  `json:"cohort_health_weight_ppm"`
	ServerHealthWeightPPM    int64                  `json:"server_health_weight_ppm"`
	CollectiveWeightPPM      int64                  `json:"collective_weight_ppm"`
	CollapseHealthPPM        int64                  `json:"collapse_health_ppm"`
	HealthyHealthPPM         int64                  `json:"healthy_health_ppm"`
	MaximumBonus             string                 `json:"maximum_bonus"`
	HealthRecoveryPPMPerHour int64                  `json:"health_recovery_ppm_per_hour"`
	HealthDecayPPMPerHour    int64                  `json:"health_decay_ppm_per_hour"`
	SolidarityWindowMS       int64                  `json:"solidarity_window_ms"`
	CohortTargetSize         int                    `json:"cohort_target_size"`
	CohortMergeFloor         int                    `json:"cohort_merge_floor"`
	NPCPopulationFloor       int                    `json:"npc_population_floor"`
	NPCWeightPPM             int64                  `json:"npc_weight_ppm"`
	NPCCompliancePPM         int64                  `json:"npc_compliance_ppm"`
	PopulationTolerancePPM   int64                  `json:"population_tolerance_ppm"`
}

type authorityKind int

const (
	authorityFunction authorityKind = iota
	authorityValue
)

type authoritySpec struct {
	label  string
	path   string
	kind   authorityKind
	symbol string
}

var formulaAuthorities = []authoritySpec{
	{label: "production.Rates", path: "production/engine.go", kind: authorityFunction, symbol: "Rates"},
	{label: "multiplier.Order", path: "multiplier/contribution.go", kind: authorityValue, symbol: "Order"},
	{label: "multiplier.OrderedSourceIDs", path: "multiplier/contribution.go", kind: authorityFunction, symbol: "OrderedSourceIDs"},
	{label: "commons.EnclosureIndex", path: "commons/formula.go", kind: authorityFunction, symbol: "EnclosureIndex"},
	{label: "commons.Modifier", path: "commons/formula.go", kind: authorityFunction, symbol: "Modifier"},
	{label: "commons.AggregateHealth", path: "commons/health.go", kind: authorityFunction, symbol: "AggregateHealth"},
	{label: "commons.SmoothHealthPPM", path: "commons/health.go", kind: authorityFunction, symbol: "SmoothHealthPPM"},
}

func main() {
	output := flag.String("output", "", "output JSON filename")
	flag.Parse()
	if *output == "" {
		fmt.Fprintln(os.Stderr, "-output is required")
		os.Exit(2)
	}
	root, err := moduleRoot()
	if err != nil {
		panic(err)
	}
	fingerprint, err := sourceFingerprint(root)
	if err != nil {
		panic(err)
	}
	commonsBytes, err := os.ReadFile(filepath.Join(root, "..", "balance", "commons", "phase0.json"))
	if err != nil {
		panic(err)
	}
	commonsCatalog, err := commons.LoadCatalog(commonsBytes)
	if err != nil {
		panic(err)
	}
	artifact := formulaArtifact{
		SchemaVersion:       4,
		ProductionRate:      "sum_generators(count * base_rate * product(multiplier_slots))",
		MultiplierSlotOrder: append([]multiplier.Slot(nil), multiplier.Order[:]...),
		WithinSlotOrder:     multiplier.WithinSlotOrder,
		SourceFingerprint:   fingerprint,
		Commons: commonsFormula{
			Enclosure:                "clamp(1 - product(clean weighted factors) / product(all weighted factors), 0, 1)",
			Compliance:               "clamp(tithe_ppm / target_tithe_ppm, 0, 1) * (1 - enclosure)",
			Health:                   "sum(weight_ppm * compliance_ppm) / sum(weight_ppm)",
			EffectiveHealth:          "0.5 * guild_health + 0.3 * cohort_health + 0.2 * server_health; guildless substitutes cohort for guild",
			Modifier:                 "1 + maximum_bonus * (0.6 * max(0, ((health - 0.35) / 0.65)^1.5) + 0.4 * solidarity)",
			Solidarity:               "sum(hourly_compliance_ppm * covered_ms) / 2592000000",
			SourceWeights:            append([]commons.SourceWeight{}, commonsCatalog.SourceWeights...),
			DefaultTithePPM:          commonsCatalog.DefaultTithePPM,
			MinimumTithePPM:          commonsCatalog.MinimumTithePPM,
			MaximumTithePPM:          commonsCatalog.MaximumTithePPM,
			GuildHealthWeightPPM:     commonsCatalog.GuildHealthWeightPPM,
			CohortHealthWeightPPM:    commonsCatalog.CohortHealthWeightPPM,
			ServerHealthWeightPPM:    commonsCatalog.ServerHealthWeightPPM,
			CollectiveWeightPPM:      commonsCatalog.CollectiveWeightPPM,
			CollapseHealthPPM:        commonsCatalog.CollapseHealthPPM,
			HealthyHealthPPM:         commonsCatalog.HealthyHealthPPM,
			MaximumBonus:             commonsCatalog.MaximumBonus.String(),
			HealthRecoveryPPMPerHour: commonsCatalog.HealthRecoveryPPMPerHour,
			HealthDecayPPMPerHour:    commonsCatalog.HealthDecayPPMPerHour,
			SolidarityWindowMS:       commonsCatalog.SolidarityWindowMS,
			CohortTargetSize:         commonsCatalog.CohortTargetSize,
			CohortMergeFloor:         commonsCatalog.CohortMergeFloor,
			NPCPopulationFloor:       commonsCatalog.NPCPopulationFloor,
			NPCWeightPPM:             commonsCatalog.NPCWeightPPM,
			NPCCompliancePPM:         commonsCatalog.NPCCompliancePPM,
			PopulationTolerancePPM:   commonsCatalog.PopulationTolerancePPM,
		},
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(*output, data, 0o644); err != nil {
		panic(err)
	}
}

func moduleRoot() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("go.mod not found from working directory")
		}
		directory = parent
	}
}

func sourceFingerprint(root string) (string, error) {
	return sourceFingerprintFrom(func(path string) ([]byte, error) {
		return os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	})
}

func sourceFingerprintFrom(readSource func(string) ([]byte, error)) (string, error) {
	hash := sha256.New()
	for _, authority := range formulaAuthorities {
		source, err := readSource(authority.path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", authority.label, err)
		}
		canonical, err := canonicalAuthority(source, authority)
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(authority.label))
		_, _ = hash.Write([]byte{'\n'})
		_, _ = hash.Write(canonical)
		_, _ = hash.Write([]byte{'\n'})
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func canonicalAuthority(source []byte, authority authoritySpec) ([]byte, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, authority.path, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", authority.label, err)
	}
	var matches []ast.Node
	for _, declaration := range file.Decls {
		switch authority.kind {
		case authorityFunction:
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Recv == nil && function.Name.Name == authority.symbol {
				matches = append(matches, function)
			}
		case authorityValue:
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.VAR && general.Tok != token.CONST {
				continue
			}
			for _, raw := range general.Specs {
				value, ok := raw.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range value.Names {
					if name.Name == authority.symbol {
						matches = append(matches, value)
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported authority kind for %s", authority.label)
		}
	}
	if len(matches) != 1 {
		return nil, fmt.Errorf("%s: found %d declarations, want exactly 1", authority.label, len(matches))
	}
	var canonical bytes.Buffer
	if err := format.Node(&canonical, fileSet, matches[0]); err != nil {
		return nil, fmt.Errorf("format %s: %w", authority.label, err)
	}
	return []byte(strings.TrimSpace(canonical.String())), nil
}
