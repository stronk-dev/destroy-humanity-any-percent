package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/releasepackage"
)

func TestNoticeAtRetainsSeparateAttributionWithoutChangingLicense(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "LICENSE"), mitText)
	writeFixture(t, filepath.Join(root, "NOTICE"), "Copyright separate author\r\nApache License version 2.0 is mentioned here.\r\n")
	writeFixture(t, filepath.Join(root, "NOTICE.extra"), "Additional attribution\n")
	text, license, err := licenseAt(root)
	if err != nil || license != "MIT" {
		t.Fatalf("NOTICE changed license classification: %q %v", license, err)
	}
	notice, err := noticeAt(root)
	if err != nil || !strings.Contains(notice, "Copyright separate author") || !strings.Contains(notice, "Additional attribution") || strings.Contains(notice, "\r") {
		t.Fatalf("notice omitted or not normalized: %v %q", err, notice)
	}
	output, err := releasepackage.ThirdPartyNotices([]releasepackage.Dependency{{Name: "fixture", Version: "1", Kind: "go", License: license, LicenseText: text, NoticeText: notice, Download: "https://example.org/fixture", PackageURL: "pkg:golang/fixture@1"}})
	if err != nil || !strings.Contains(string(output), notice) {
		t.Fatalf("separate NOTICE not delivered: %v %s", err, output)
	}
}

func TestNoticeAtRefusesInvalidPresentNotices(t *testing.T) {
	for _, variant := range []string{"empty", "directory", "symlink"} {
		t.Run(variant, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "NOTICE")
			switch variant {
			case "empty":
				writeFixture(t, path, " \n")
			case "directory":
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				outside := filepath.Join(t.TempDir(), "notice")
				writeFixture(t, outside, "outside attribution")
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := noticeAt(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
				t.Fatalf("invalid present NOTICE accepted: %v", err)
			}
		})
	}
	if text, err := noticeAt(t.TempDir()); err != nil || text != "" {
		t.Fatalf("missing optional NOTICE refused: %v %q", err, text)
	}
}

func TestNoticeCannotSupplyMissingLicense(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, filepath.Join(root, "NOTICE"), mitText)
	if _, _, err := licenseAt(root); !errors.Is(err, releasepackage.ErrInvalidContent) {
		t.Fatalf("NOTICE substituted for missing LICENSE: %v", err)
	}
}

func TestGoInventoryPreservesLinkedModuleNotice(t *testing.T) {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "../../.."))
	const module = "github.com/prometheus/client_golang"
	command := exec.Command("go", "list", "-m", "-json", module)
	command.Dir = filepath.Join(root, "server")
	data, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	var source struct{ Dir, Version string }
	if json.Unmarshal(data, &source) != nil || source.Dir == "" || source.Version == "" {
		t.Fatal("actual module source identity missing")
	}
	required, err := os.ReadFile(filepath.Join(source.Dir, "NOTICE"))
	if err != nil || len(strings.TrimSpace(string(required))) == 0 {
		t.Fatal("actual linked module NOTICE missing", err)
	}
	dependencies, err := discoverGoDependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, dependency := range dependencies {
		if dependency.Name != module {
			continue
		}
		if dependency.Version != source.Version {
			t.Fatal("inventory and actual source versions disagree")
		}
		output, err := releasepackage.ThirdPartyNotices([]releasepackage.Dependency{dependency})
		if err != nil {
			t.Fatal(err)
		}
		text := strings.TrimSpace(strings.ReplaceAll(string(required), "\r\n", "\n"))
		if !strings.Contains(string(output), text) {
			t.Fatalf("required NOTICE omitted for actually linked %s@%s", module, source.Version)
		}
		t.Logf("delivered actual %s@%s NOTICE", module, source.Version)
		return
	}
	t.Fatal("selected module is not actually linked into a shipped Go command")
}
