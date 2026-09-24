package releasepackage

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

const mitFixture = `MIT License

Copyright (c) Example

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND.`

const iscFixture = `Copyright (c) Example

Permission to use, copy, modify, and distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES.`

func TestReleaseMetadataIsSortedLicensedAndSPDX23(t *testing.T) {
	dependencies, err := SortDependencies([]Dependency{
		{Name: "browser-lib", Version: "2.0.0", Kind: "npm", License: "MIT", LicenseText: mitFixture, Download: "https://registry.npmjs.org/browser-lib/-/browser-lib-2.0.0.tgz", PackageURL: "pkg:npm/browser-lib@2.0.0"},
		{Name: "go.example/lib", Version: "v1.0.0", Kind: "go", License: "MIT", LicenseText: mitFixture, Download: "https://proxy.golang.org/go.example/lib/@v/v1.0.0.zip", PackageURL: "pkg:golang/go.example/lib@v1.0.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	notices, err := ThirdPartyNotices(dependencies)
	if err != nil || bytes.Index(notices, []byte("go.example/lib")) > bytes.Index(notices, []byte("browser-lib")) {
		t.Fatalf("notice ordering/error: %v\n%s", err, notices)
	}
	sbom, err := BuildSPDX("Cloud Clicker", "0.1.0", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC), dependencies)
	if err != nil {
		t.Fatal(err)
	}
	var document SPDXDocument
	if json.Unmarshal(sbom, &document) != nil || document.SPDXVersion != "SPDX-2.3" || len(document.Packages) != 3 || len(document.Relationships) != 3 {
		t.Fatalf("invalid SPDX document: %s", sbom)
	}
}

func TestReleaseMetadataRejectsUnknownLicenseDuplicateAndBadCommit(t *testing.T) {
	valid := Dependency{Name: "lib", Version: "1.0.0", Kind: "npm", License: "MIT", LicenseText: mitFixture,
		Download: "https://registry.npmjs.org/lib/-/lib-1.0.0.tgz", PackageURL: "pkg:npm/lib@1.0.0"}
	unknown := valid
	unknown.License = "NOASSERTION"
	if _, err := SortDependencies([]Dependency{unknown}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("unknown license accepted: %v", err)
	}
	if _, err := SortDependencies([]Dependency{valid, valid}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("duplicate dependency accepted: %v", err)
	}
	if _, err := BuildSPDX("Cloud Clicker", "0.1.0", "short", time.Now(), []Dependency{valid}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("bad commit accepted: %v", err)
	}
}

func TestDetectPermissiveLicenseRecognizesISC(t *testing.T) {
	license, err := DetectPermissiveLicense(iscFixture)
	if err != nil || license != "ISC" {
		t.Fatalf("ISC detection: license=%q err=%v", license, err)
	}
	dependency := Dependency{Name: "github.com/coder/websocket", Version: "v1.8.15", Kind: "go", License: license,
		LicenseText: iscFixture, Download: "https://proxy.golang.org/github.com/coder/websocket/@v/v1.8.15.zip", PackageURL: "pkg:golang/github.com/coder/websocket@v1.8.15"}
	if err := ValidateDependencies([]Dependency{dependency}); err != nil {
		t.Fatalf("valid ISC dependency rejected: %v", err)
	}
}

func TestValidateSPDXRejectsPlausibleButUnidentifiedJSON(t *testing.T) {
	if err := ValidateSPDX([]byte(`{"packages":[]}`)); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("unidentified JSON accepted as SPDX: %v", err)
	}
}

func TestDetectPermissiveLicenseSeparatesISCAndZeroClauseBSD(t *testing.T) {
	grant := "Permission to use, copy, modify, and/or distribute this software for any\npurpose with or without fee is hereby granted"
	disclaimer := "\n\nTHE SOFTWARE IS PROVIDED \"AS IS\" AND THE AUTHOR DISCLAIMS ALL WARRANTIES.\n"
	if license, err := DetectPermissiveLicense(grant + "." + disclaimer); err != nil || license != "0BSD" {
		t.Fatalf("0BSD text license=%q err=%v", license, err)
	}
	if license, err := DetectPermissiveLicense(grant + ", provided that the above copyright notice and this permission notice appear in all copies." + disclaimer); err != nil || license != "ISC" {
		t.Fatalf("ISC text license=%q err=%v", license, err)
	}
}
