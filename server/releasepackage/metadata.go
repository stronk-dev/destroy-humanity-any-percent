package releasepackage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Dependency struct {
	Name        string
	Version     string
	Kind        string
	License     string
	LicenseText string
	Download    string
	PackageURL  string
}

type SPDXDocument struct {
	SPDXVersion       string             `json:"spdxVersion"`
	DataLicense       string             `json:"dataLicense"`
	SPDXID            string             `json:"SPDXID"`
	Name              string             `json:"name"`
	DocumentNamespace string             `json:"documentNamespace"`
	CreationInfo      CreationInfo       `json:"creationInfo"`
	Packages          []SPDXPackage      `json:"packages"`
	Relationships     []SPDXRelationship `json:"relationships"`
}

type CreationInfo struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
}

type SPDXPackage struct {
	Name             string            `json:"name"`
	SPDXID           string            `json:"SPDXID"`
	VersionInfo      string            `json:"versionInfo"`
	DownloadLocation string            `json:"downloadLocation"`
	FilesAnalyzed    bool              `json:"filesAnalyzed"`
	LicenseConcluded string            `json:"licenseConcluded"`
	LicenseDeclared  string            `json:"licenseDeclared"`
	CopyrightText    string            `json:"copyrightText"`
	ExternalRefs     []SPDXExternalRef `json:"externalRefs,omitempty"`
}

type SPDXExternalRef struct {
	ReferenceCategory string `json:"referenceCategory"`
	ReferenceType     string `json:"referenceType"`
	ReferenceLocator  string `json:"referenceLocator"`
}

type SPDXRelationship struct {
	SPDXElementID      string `json:"spdxElementId"`
	RelationshipType   string `json:"relationshipType"`
	RelatedSPDXElement string `json:"relatedSpdxElement"`
}

var spdxLicense = regexp.MustCompile(`^(Apache-2\.0|BSD-2-Clause|BSD-3-Clause|MIT)( AND (Apache-2\.0|BSD-2-Clause|BSD-3-Clause|MIT))*$`)

func DetectPermissiveLicense(text string) (string, error) {
	normalized := strings.ToLower(strings.ReplaceAll(text, "\r\n", "\n"))
	licenses := []string{}
	if strings.Contains(normalized, "apache license") && strings.Contains(normalized, "version 2.0") {
		licenses = append(licenses, "Apache-2.0")
	}
	if strings.Contains(normalized, "permission is hereby granted, free of charge") && strings.Contains(normalized, "the software is provided \"as is\"") {
		licenses = append(licenses, "MIT")
	}
	if strings.Contains(normalized, "redistribution and use in source and binary forms") {
		if strings.Contains(normalized, "neither the name") {
			licenses = append(licenses, "BSD-3-Clause")
		} else if strings.Contains(normalized, "this list of conditions") {
			licenses = append(licenses, "BSD-2-Clause")
		}
	}
	sort.Strings(licenses)
	licenses = uniqueStrings(licenses)
	if len(licenses) == 0 {
		return "", ErrInvalidContent
	}
	return strings.Join(licenses, " AND "), nil
}

func ValidateDependencies(dependencies []Dependency) error {
	if len(dependencies) == 0 {
		return ErrInvalidContent
	}
	prior := ""
	for _, dependency := range dependencies {
		identity := dependency.Kind + "\x00" + dependency.Name + "\x00" + dependency.Version
		if dependency.Name == "" || dependency.Version == "" || dependency.Kind == "" || dependency.Download == "" ||
			dependency.PackageURL == "" || !spdxLicense.MatchString(dependency.License) || strings.TrimSpace(dependency.LicenseText) == "" || identity <= prior {
			return ErrInvalidContent
		}
		detected, err := DetectPermissiveLicense(dependency.LicenseText)
		if err != nil || detected != dependency.License {
			return ErrInvalidContent
		}
		prior = identity
	}
	return nil
}

func ThirdPartyNotices(dependencies []Dependency) ([]byte, error) {
	if err := ValidateDependencies(dependencies); err != nil {
		return nil, err
	}
	var output strings.Builder
	output.WriteString("Cloud Clicker third-party licenses\n")
	output.WriteString("Generated from the dependencies linked into the gameserver and bundled into the browser client.\n\n")
	for _, dependency := range dependencies {
		fmt.Fprintf(&output, "================================================================================\n%s %s (%s)\n%s\n%s\n\n",
			dependency.Name, dependency.Version, dependency.Kind, dependency.License, strings.TrimSpace(strings.ReplaceAll(dependency.LicenseText, "\r\n", "\n")))
	}
	return []byte(output.String()), nil
}

func BuildSPDX(name, version, commit string, created time.Time, dependencies []Dependency) ([]byte, error) {
	if name == "" || version == "" || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(commit) || created.IsZero() || ValidateDependencies(dependencies) != nil {
		return nil, ErrInvalidContent
	}
	identityBytes, _ := json.Marshal(dependencies)
	namespaceHash := sha256.Sum256(append([]byte(name+"\x00"+version+"\x00"+commit+"\x00"), identityBytes...))
	document := SPDXDocument{
		SPDXVersion: "SPDX-2.3", DataLicense: "CC0-1.0", SPDXID: "SPDXRef-DOCUMENT", Name: name + " " + version,
		DocumentNamespace: "https://github.com/stronk-dev/destroy-humanity-any-percent/sbom/" + hex.EncodeToString(namespaceHash[:]),
		CreationInfo:      CreationInfo{Created: created.UTC().Format(time.RFC3339), Creators: []string{"Tool: cloud-clicker/gen-release-metadata"}},
		Packages: []SPDXPackage{{Name: name, SPDXID: "SPDXRef-Package-cloud-clicker", VersionInfo: version,
			DownloadLocation: "https://github.com/stronk-dev/destroy-humanity-any-percent/commit/" + commit,
			FilesAnalyzed:    false, LicenseConcluded: "MIT", LicenseDeclared: "MIT", CopyrightText: "Copyright (c) 2026 Marco van Dijk"}},
		Relationships: []SPDXRelationship{{SPDXElementID: "SPDXRef-DOCUMENT", RelationshipType: "DESCRIBES", RelatedSPDXElement: "SPDXRef-Package-cloud-clicker"}},
	}
	for _, dependency := range dependencies {
		idHash := sha256.Sum256([]byte(dependency.Kind + "\x00" + dependency.Name + "\x00" + dependency.Version))
		packageID := "SPDXRef-Package-" + hex.EncodeToString(idHash[:8])
		document.Packages = append(document.Packages, SPDXPackage{
			Name: dependency.Name, SPDXID: packageID, VersionInfo: dependency.Version,
			DownloadLocation: dependency.Download, FilesAnalyzed: false, LicenseConcluded: dependency.License,
			LicenseDeclared: dependency.License, CopyrightText: "NOASSERTION",
			ExternalRefs: []SPDXExternalRef{{ReferenceCategory: "PACKAGE-MANAGER", ReferenceType: "purl", ReferenceLocator: dependency.PackageURL}},
		})
		document.Relationships = append(document.Relationships, SPDXRelationship{SPDXElementID: "SPDXRef-Package-cloud-clicker", RelationshipType: "DEPENDS_ON", RelatedSPDXElement: packageID})
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

func SortDependencies(dependencies []Dependency) ([]Dependency, error) {
	result := append([]Dependency(nil), dependencies...)
	sort.Slice(result, func(left, right int) bool {
		return result[left].Kind+"\x00"+result[left].Name+"\x00"+result[left].Version < result[right].Kind+"\x00"+result[right].Name+"\x00"+result[right].Version
	})
	for index := 1; index < len(result); index++ {
		if result[index-1].Kind == result[index].Kind && result[index-1].Name == result[index].Name && result[index-1].Version == result[index].Version {
			return nil, errors.Join(ErrInvalidContent, fmt.Errorf("duplicate dependency %s", result[index].Name))
		}
	}
	if err := ValidateDependencies(result); err != nil {
		return nil, err
	}
	return result, nil
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
