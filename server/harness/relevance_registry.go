package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cloud-clicker/server/epochseed"
)

const relevanceRegistryPath = "testdata/harness/relevance/registry-v1.json"

type relevanceRegistry struct {
	SchemaVersion *int                     `json:"schema_version"`
	Entries       []relevanceRegistryEntry `json:"entries"`
}

type relevanceRegistryEntry struct {
	EconomyCatalog  *string `json:"economy_catalog"`
	Scenario        *string `json:"scenario"`
	RelevancePolicy *string `json:"relevance_policy"`
	GoldenReport    *string `json:"golden_report"`
}

func ValidateRelevanceRegistry(repositoryRoot string) error {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(relevanceRegistryPath)))
	if err != nil {
		return err
	}
	var registry relevanceRegistry
	if err := decodeRelevanceStrict(data, &registry); err != nil || registry.SchemaVersion == nil || *registry.SchemaVersion != 1 || registry.Entries == nil {
		return errors.New("invalid relevance scenario registry")
	}
	paths := map[string]bool{}
	prior := ""
	for index, entry := range registry.Entries {
		if entry.EconomyCatalog == nil || entry.Scenario == nil || entry.RelevancePolicy == nil || entry.GoldenReport == nil ||
			*entry.EconomyCatalog == "" || prior != "" && prior >= *entry.EconomyCatalog {
			return fmt.Errorf("invalid relevance registry entry %d", index)
		}
		prior = *entry.EconomyCatalog
		for _, path := range []string{*entry.EconomyCatalog, *entry.Scenario, *entry.RelevancePolicy, *entry.GoldenReport} {
			if filepath.ToSlash(filepath.Clean(path)) != path {
				return fmt.Errorf("invalid relevance registry path %q", path)
			}
			if _, err := os.Stat(filepath.Join(repositoryRoot, filepath.FromSlash(path))); err != nil {
				return fmt.Errorf("relevance registry path %q: %w", path, err)
			}
		}
		paths[*entry.EconomyCatalog] = true
	}
	bundle, err := epochseed.Load(repositoryRoot)
	if err != nil {
		return err
	}
	activeEconomy, ok := epochseed.ArtifactPath(bundle.Seed, "economy")
	if !ok {
		return errors.New("epoch seed has no economy artifact")
	}
	version, err := economySchemaVersion(filepath.Join(repositoryRoot, filepath.FromSlash(activeEconomy)))
	if err != nil {
		return err
	}
	if version >= 4 && !paths[activeEconomy] {
		return fmt.Errorf("active schema-v%d economy catalog %q has no relevance registry entry", version, activeEconomy)
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
