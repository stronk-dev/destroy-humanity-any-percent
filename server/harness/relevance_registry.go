package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cloud-clicker/server/epochseed"
)

const relevanceRegistryPath = "testdata/harness/relevance/registry-v1.json"

type RelevanceRegistryEntry struct {
	EconomyCatalog         string `json:"economy_catalog"`
	Scenario               string `json:"scenario"`
	RelevancePolicy        string `json:"relevance_policy"`
	GoldenReport           string `json:"golden_report"`
	JustificationChangelog string `json:"justification_changelog"`
	Active                 bool   `json:"-"`
	ConstantsHash          string `json:"-"`
}

type relevanceRegistryWireEntry struct {
	EconomyCatalog         *string `json:"economy_catalog"`
	Scenario               *string `json:"scenario"`
	RelevancePolicy        *string `json:"relevance_policy"`
	GoldenReport           *string `json:"golden_report"`
	JustificationChangelog *string `json:"justification_changelog"`
}

func LoadRelevanceRegistry(repositoryRoot string) ([]RelevanceRegistryEntry, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(relevanceRegistryPath)))
	if err != nil {
		return nil, err
	}
	var registry struct {
		SchemaVersion *int                         `json:"schema_version"`
		Entries       []relevanceRegistryWireEntry `json:"entries"`
	}
	if err := decodeRelevanceStrict(data, &registry); err != nil || registry.SchemaVersion == nil || *registry.SchemaVersion != 1 || registry.Entries == nil {
		return nil, errors.New("invalid relevance scenario registry")
	}
	bundle, err := epochseed.Load(repositoryRoot)
	if err != nil {
		return nil, err
	}
	activeEconomy, ok := epochseed.ArtifactPath(bundle.Seed, "economy")
	if !ok {
		return nil, errors.New("epoch seed has no economy artifact")
	}
	paths := map[string]bool{}
	ownedPaths := map[string]bool{}
	entries := make([]RelevanceRegistryEntry, 0, len(registry.Entries))
	prior := ""
	for index, entry := range registry.Entries {
		if entry.EconomyCatalog == nil || entry.Scenario == nil || entry.RelevancePolicy == nil || entry.GoldenReport == nil || entry.JustificationChangelog == nil ||
			*entry.EconomyCatalog == "" || prior != "" && prior >= *entry.EconomyCatalog {
			return nil, fmt.Errorf("invalid relevance registry entry %d", index)
		}
		prior = *entry.EconomyCatalog
		for _, path := range []string{*entry.EconomyCatalog, *entry.Scenario, *entry.RelevancePolicy, *entry.GoldenReport, *entry.JustificationChangelog} {
			if filepath.ToSlash(filepath.Clean(path)) != path {
				return nil, fmt.Errorf("invalid relevance registry path %q", path)
			}
			if ownedPaths[path] {
				return nil, fmt.Errorf("duplicate relevance registry path %q", path)
			}
			ownedPaths[path] = true
			if _, err := os.Stat(filepath.Join(repositoryRoot, filepath.FromSlash(path))); err != nil {
				return nil, fmt.Errorf("relevance registry path %q: %w", path, err)
			}
		}
		paths[*entry.EconomyCatalog] = true
		entries = append(entries, RelevanceRegistryEntry{EconomyCatalog: *entry.EconomyCatalog, Scenario: *entry.Scenario,
			RelevancePolicy: *entry.RelevancePolicy, GoldenReport: *entry.GoldenReport, JustificationChangelog: *entry.JustificationChangelog,
			Active: *entry.EconomyCatalog == activeEconomy})
	}
	version, err := economySchemaVersion(filepath.Join(repositoryRoot, filepath.FromSlash(activeEconomy)))
	if err != nil {
		return nil, err
	}
	if version >= 4 && !paths[activeEconomy] {
		return nil, fmt.Errorf("active schema-v%d economy catalog %q has no relevance registry entry", version, activeEconomy)
	}
	for index := range entries {
		suite, err := LoadRelevanceSuite(repositoryRoot, entries[index].Scenario)
		if err != nil {
			return nil, err
		}
		if suite.Scenario.Catalog != entries[index].EconomyCatalog || suite.Scenario.Policy != entries[index].RelevancePolicy {
			return nil, fmt.Errorf("relevance registry/scenario artifact mismatch for %q", entries[index].Scenario)
		}
		if err := validateTrapJustifications(repositoryRoot, entries[index], suite.Policy); err != nil {
			return nil, err
		}
		if !entries[index].Active || version < 4 {
			continue
		}
		if err := bindActiveRelevanceAuthority(&entries[index], suite.Scenario, bundle); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func validateTrapJustifications(repositoryRoot string, entry RelevanceRegistryEntry, policy *RelevancePolicy) error {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.JustificationChangelog)))
	if err != nil {
		return err
	}
	for _, item := range policy.Items {
		if !item.TrapExempt || item.JustificationKey == nil {
			continue
		}
		if !bytes.Contains(data, []byte("`"+*item.JustificationKey+"`")) {
			return fmt.Errorf("trap exemption %q has no changelog evidence in %q", *item.JustificationKey, entry.JustificationChangelog)
		}
	}
	return nil
}

func bindActiveRelevanceAuthority(entry *RelevanceRegistryEntry, scenario RelevanceScenario, bundle epochseed.Bundle) error {
	activeRoutes, routesPresent := epochseed.ArtifactPath(bundle.Seed, "routes")
	activePolicy, policyPresent := epochseed.ArtifactPath(bundle.Seed, "relevance_policy")
	if !routesPresent || !policyPresent || scenario.RoutesCatalog != activeRoutes || entry.RelevancePolicy != activePolicy {
		return fmt.Errorf("active relevance entry %q is not owned by the epoch artifact manifest", entry.Scenario)
	}
	current := epochseed.Current(bundle.Seed)
	if !epochseed.Accepts(current, bundle.Hash) {
		return fmt.Errorf("active relevance constants hash %s is not accepted by epoch %d", bundle.Hash, current.ID)
	}
	if entry.JustificationChangelog != current.ChangelogRef {
		return fmt.Errorf("active relevance entry %q does not use current epoch changelog", entry.Scenario)
	}
	entry.ConstantsHash = bundle.Hash
	return nil
}

func LoadRegisteredRelevanceSuite(repositoryRoot string, entry RelevanceRegistryEntry) (*RelevanceSuite, error) {
	suite, err := LoadRelevanceSuite(repositoryRoot, entry.Scenario)
	if err != nil {
		return nil, err
	}
	if suite.Scenario.Catalog != entry.EconomyCatalog || suite.Scenario.Policy != entry.RelevancePolicy {
		return nil, fmt.Errorf("relevance registry/scenario artifact mismatch for %q", entry.Scenario)
	}
	if entry.Active {
		if !relevanceHashPattern.MatchString(entry.ConstantsHash) {
			return nil, errors.New("active relevance entry has no epoch constants hash")
		}
		suite.ConstantsHash = entry.ConstantsHash
	}
	return suite, nil
}

func ValidateRelevanceRegistry(repositoryRoot string) error {
	_, err := LoadRelevanceRegistry(repositoryRoot)
	return err
}

func ValidateActiveRelevanceReport(entry RelevanceRegistryEntry, report RelevanceReport) error {
	if entry.Active && len(report.Failures) > 0 {
		return fmt.Errorf("active relevance gate failures for %q: %v", entry.Scenario, report.Failures)
	}
	return nil
}

func economySchemaVersion(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var envelope struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.SchemaVersion < 1 {
		return 0, errors.New("invalid economy schema version")
	}
	return envelope.SchemaVersion, nil
}
