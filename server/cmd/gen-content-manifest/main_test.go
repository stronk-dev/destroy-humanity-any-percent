package main

import (
	"strings"
	"testing"
)

func TestManifestBytesPinsBothIdentities(t *testing.T) {
	constantsHash := "sha256:" + strings.Repeat("a", 64)
	copyHash := "sha256:" + strings.Repeat("b", 64)
	data, err := manifestBytes(constantsHash, copyHash)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"schema_version\": 1,\n  \"constants_hash\": \"" + constantsHash + "\",\n  \"copy_hash\": \"" + copyHash + "\"\n}\n"
	if string(data) != want {
		t.Fatalf("manifest mismatch\nwant %s\n got %s", want, data)
	}
}

func TestManifestBytesRejectsNonCanonicalIdentity(t *testing.T) {
	if _, err := manifestBytes("not-a-hash", "sha256:"+strings.Repeat("b", 64)); err == nil {
		t.Fatal("non-canonical constants hash accepted")
	}
}
