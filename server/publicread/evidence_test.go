package publicread

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"cloud-clicker/server/leaderboard"
)

const evidenceStream = "01986666-0000-4000-8000-000000000001"

type fakeEvidence struct {
	value leaderboard.PublicRunEvidence
	err   error
}

func (fake *fakeEvidence) PublicRunEvidence(_ context.Context, stream string, seq int64) (leaderboard.PublicRunEvidence, error) {
	if fake.err != nil {
		return leaderboard.PublicRunEvidence{}, fake.err
	}
	if fake.value.RunID == "" || stream+":"+fmt.Sprint(seq) != fake.value.RunID {
		return leaderboard.PublicRunEvidence{}, leaderboard.ErrUnknownPublicRun
	}
	return fake.value, nil
}

func exampleEvidence() *fakeEvidence {
	genesis := []byte(`{"v":1,"nested":{"z":2,"a":1}}`)
	// Byte-preservation fixture; a genuine gzip/verifier archive is tested on Postgres.
	archive := []byte{0x1f, 0x8b, 0x08, 0x00, 0x01, 0x02, 0x03, 0x04}
	return &fakeEvidence{value: leaderboard.PublicRunEvidence{
		RunID: evidenceStream + ":1", EngineVersion: "0.1.0", ConstantsHash: hashOf("a"), GenesisVersion: 1,
		Genesis: genesis, GenesisSHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(genesis)),
		ReplayLog: archive, ReplayLogSHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(archive)),
	}}
}

func TestEvidenceHTTPPreservesRawBytesAndManifestReferences(t *testing.T) {
	fixture := newPublicCacheFixture(t)
	base := "/api/public/v1/runs/" + evidenceStream + "/1"
	response := publicCacheRequest(fixture.router, base+"/verdict", "", "192.0.2.7")
	var manifest publicRunVerdict
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &manifest) != nil {
		t.Fatalf("manifest %d %s", response.Code, response.Body.String())
	}
	expected := fixture.evidence.value
	want := publicRunVerdict{CatalogURL: "/api/public/v1/catalogs/" + expected.ConstantsHash, ConstantsHash: expected.ConstantsHash,
		EngineVersion: expected.EngineVersion, GenesisSHA256: expected.GenesisSHA256, GenesisURL: base + "/genesis",
		ReplayLogSHA256: expected.ReplayLogSHA256, ReplayLogURL: base + "/replay-log", RunID: expected.RunID, Verdict: "verified"}
	if manifest != want {
		t.Fatalf("manifest=%+v want=%+v", manifest, want)
	}
	for _, test := range []struct {
		path, hash, contentType string
		body                    []byte
	}{
		{manifest.GenesisURL, manifest.GenesisSHA256, "application/json", expected.Genesis},
		{manifest.ReplayLogURL, manifest.ReplayLogSHA256, "application/gzip", expected.ReplayLog},
	} {
		response := publicCacheRequest(fixture.router, test.path, "", "192.0.2.7")
		if response.Code != http.StatusOK || !bytes.Equal(response.Body.Bytes(), test.body) ||
			"sha256:"+response.Header().Get(EvidenceHashHeader) != test.hash || response.Header().Get("Content-Type") != test.contentType ||
			response.Header().Get("ETag") != `"`+test.hash[len("sha256:"):]+`"` {
			t.Fatalf("raw evidence %s: %d %v %q", test.path, response.Code, response.Header(), response.Body.Bytes())
		}
	}
}

func TestEvidenceHTTPRefusesUnknownPrivateAndCorruptEvidenceWithoutCaching(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
		body   []byte
	}{
		{"private or unknown run", leaderboard.ErrUnknownPublicRun, http.StatusNotFound, unknownRun},
		{"corrupt public evidence", leaderboard.ErrInvalidPublicRunEvidence, http.StatusInternalServerError, internalPublic},
		{"database error", errors.New("database unavailable"), http.StatusInternalServerError, internalPublic},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newPublicCacheFixture(t)
			fixture.evidence.err = test.err
			for _, suffix := range []string{"genesis", "replay-log", "verdict"} {
				response := publicCacheRequest(fixture.router, "/api/public/v1/runs/"+evidenceStream+"/1/"+suffix, `"cached-before-refusal"`, "192.0.2.7")
				if response.Code != test.status || !bytes.Equal(response.Body.Bytes(), test.body) || response.Header().Get("ETag") != "" ||
					response.Header().Get("Cache-Control") != "" || response.Header().Get(EvidenceHashHeader) != "" || response.Header().Get("X-Request-ID") != "public-cache-check" {
					t.Fatalf("refusal %s: %d %v %q", suffix, response.Code, response.Header(), response.Body.String())
				}
			}
		})
	}
	fixture := newPublicCacheFixture(t)
	for _, seq := range []string{"0", "01", "-1", "+1", "9007199254740992", "abc"} {
		for _, suffix := range []string{"genesis", "replay-log", "verdict"} {
			response := publicCacheRequest(fixture.router, "/api/public/v1/runs/"+evidenceStream+"/"+seq+"/"+suffix, "", "192.0.2.7")
			if response.Code != http.StatusNotFound || !bytes.Equal(response.Body.Bytes(), unknownRun) {
				t.Fatalf("invalid sequence %s/%s: %d %s", seq, suffix, response.Code, response.Body.String())
			}
		}
	}
}

func TestEvidenceHTTPRejectsRawHashDrift(t *testing.T) {
	for _, suffix := range []string{"genesis", "replay-log"} {
		t.Run(suffix, func(t *testing.T) {
			fixture := newPublicCacheFixture(t)
			fixture.evidence.value.Genesis[0] = 'X'
			fixture.evidence.value.ReplayLog[0] = 'X'
			response := publicCacheRequest(fixture.router, "/api/public/v1/runs/"+evidenceStream+"/1/"+suffix, "", "192.0.2.7")
			if response.Code != http.StatusInternalServerError || !bytes.Equal(response.Body.Bytes(), internalPublic) || response.Header().Get("ETag") != "" {
				t.Fatalf("corrupt bytes bypassed the raw response descriptor: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
