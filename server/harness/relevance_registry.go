package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud-clicker/server/epochseed"
)

const relevanceRegistryPath = "testdata/harness/relevance/registry-v1.json"

type RelevanceRegistryEntry struct {
	EconomyCatalog         string `json:"economy_catalog"`
	Scenario               string `json:"scenario"`
	RelevancePolicy        string `json:"relevance_policy"`
	GoldenReport           string `json:"golden_report"`
	BranchReport           string `json:"branch_report,omitempty"`
	JustificationChangelog string `json:"justification_changelog"`
	Active                 bool   `json:"-"`
	ConstantsHash          string `json:"-"`
}

type relevanceRegistryWireEntry struct {
	EconomyCatalog         *string `json:"economy_catalog"`
	Scenario               *string `json:"scenario"`
	RelevancePolicy        *string `json:"relevance_policy"`
	GoldenReport           *string `json:"golden_report"`
	BranchReport           *string `json:"branch_report"`
	JustificationChangelog *string `json:"justification_changelog"`
}

func LoadRelevanceRegistry(repositoryRoot string) ([]RelevanceRegistryEntry, error) {
	return loadRelevanceRegistry(repositoryRoot, false)
}

// LoadRelevanceRegistryForUpdate permits only generated evidence paths to be
// absent. All source artifacts remain mandatory, and the ordinary loader stays
// fail-closed so a check can never accept an unmaterialized registry entry.
func LoadRelevanceRegistryForUpdate(repositoryRoot string) ([]RelevanceRegistryEntry, error) {
	return loadRelevanceRegistry(repositoryRoot, true)
}

func loadRelevanceRegistry(repositoryRoot string, allowMissingEvidence bool) ([]RelevanceRegistryEntry, error) {
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
		entryPaths := []string{*entry.EconomyCatalog, *entry.Scenario, *entry.RelevancePolicy, *entry.GoldenReport, *entry.JustificationChangelog}
		evidencePaths := map[string]bool{*entry.GoldenReport: true}
		if entry.BranchReport != nil {
			entryPaths = append(entryPaths, *entry.BranchReport)
			evidencePaths[*entry.BranchReport] = true
		}
		for _, path := range entryPaths {
			if filepath.ToSlash(filepath.Clean(path)) != path {
				return nil, fmt.Errorf("invalid relevance registry path %q", path)
			}
			if ownedPaths[path] {
				return nil, fmt.Errorf("duplicate relevance registry path %q", path)
			}
			ownedPaths[path] = true
			if err := validateRelevanceRegistryPath(repositoryRoot, path, evidencePaths[path], allowMissingEvidence); err != nil {
				return nil, fmt.Errorf("relevance registry path %q: %w", path, err)
			}
		}
		paths[*entry.EconomyCatalog] = true
		branchReport := ""
		if entry.BranchReport != nil {
			branchReport = *entry.BranchReport
		}
		entries = append(entries, RelevanceRegistryEntry{EconomyCatalog: *entry.EconomyCatalog, Scenario: *entry.Scenario,
			RelevancePolicy: *entry.RelevancePolicy, GoldenReport: *entry.GoldenReport, BranchReport: branchReport, JustificationChangelog: *entry.JustificationChangelog,
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
		if entries[index].BranchReport == "" {
			return nil, fmt.Errorf("active relevance entry %q has no branch report", entries[index].Scenario)
		}
		if err := bindActiveRelevanceAuthority(&entries[index], suite.Scenario, bundle); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func validateRelevanceRegistryPath(repositoryRoot, path string, evidence, allowMissingEvidence bool) error {
	_, err := os.Stat(filepath.Join(repositoryRoot, filepath.FromSlash(path)))
	if allowMissingEvidence && evidence && errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
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
	activePolicy, policyPresent := epochseed.ArtifactPath(bundle.Seed, "relevance")
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
		return fmt.Errorf("active relevance gate failures for %q require branch evidence: %v", entry.Scenario, report.Failures)
	}
	return nil
}

func LoadRegisteredRelevanceBranchReport(repositoryRoot string, entry RelevanceRegistryEntry) (RelevanceBranchReport, error) {
	if entry.BranchReport == "" {
		return RelevanceBranchReport{}, errors.New("relevance registry entry has no branch report")
	}
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.BranchReport)))
	if err != nil {
		return RelevanceBranchReport{}, err
	}
	var report RelevanceBranchReport
	if err := decodeRelevanceStrict(data, &report); err != nil {
		return RelevanceBranchReport{}, err
	}
	if err := ValidateRelevanceBranchReport(report); err != nil {
		return RelevanceBranchReport{}, err
	}
	return report, nil
}

// ValidateActiveRelevanceEvidence composes the whole-path measurement with
// the branch-specific proofs accepted by T01-C23/C29. Active content may carry
// whole-path findings only when every affected purchasable has one passing
// branch proof under byte-identical scenario, constants, and policy identity.
func ValidateActiveRelevanceEvidence(entry RelevanceRegistryEntry, report RelevanceReport, branch RelevanceBranchReport) error {
	if !entry.Active {
		return nil
	}
	if entry.BranchReport == "" || report.ScenarioID != branch.ScenarioID || report.ScenarioHash != branch.ScenarioHash ||
		report.ConstantsHash != branch.ConstantsHash || report.RelevancePolicyHash != branch.RelevancePolicyHash ||
		report.ConstantsHash != entry.ConstantsHash {
		return fmt.Errorf("active relevance evidence identity mismatch for %q", entry.Scenario)
	}
	if err := ValidateRelevanceReport(report); err != nil {
		return err
	}
	if err := ValidateRelevanceBranchReport(branch); err != nil {
		return err
	}
	if len(branch.Failures) != 0 {
		return fmt.Errorf("active relevance branch failures for %q: %v", entry.Scenario, branch.Failures)
	}
	want := make([]string, 0)
	for _, item := range report.Items {
		if !item.RelevancePassed {
			want = append(want, item.PurchasableID)
		}
	}
	for _, failure := range report.Failures {
		covered := false
		for _, id := range want {
			if strings.HasSuffix(failure, ":"+id) {
				covered = true
				break
			}
		}
		if !covered {
			return fmt.Errorf("active relevance finding has no branch evidence for %q: %s", entry.Scenario, failure)
		}
	}
	got := make([]string, 0, len(branch.Proofs))
	for _, proof := range branch.Proofs {
		if proof.Passed {
			got = append(got, proof.PurchasableID)
		}
	}
	if fmt.Sprint(want) != fmt.Sprint(got) {
		return fmt.Errorf("active relevance branch coverage mismatch for %q: whole_path=%v branch=%v", entry.Scenario, want, got)
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
