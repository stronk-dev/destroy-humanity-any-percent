package gameserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/publicread"
)

// Uses a real queue-verified/archived run from the composed Exit witness. Catalogs
// are supplied from its pinned DB bundle until the public catalog reader exists;
// this is download/replay proof, not yet the complete third-party public loop.
func assertPublicVerifiedEvidence(t *testing.T, db *sql.DB, server *httptest.Server, stream string, genesis []byte, bundle production.CatalogBundle, privateValues map[string]string) {
	t.Helper()
	base := "/api/public/v1/runs/" + stream + "/1"
	download := func(path, contentType string) []byte {
		t.Helper()
		response, err := server.Client().Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(body))
		if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != contentType ||
			response.Header.Get("ETag") != `"`+digest+`"` || response.Header.Get("Cache-Control") != "public,max-age=31536000,immutable" || response.Header.Get("X-Request-ID") == "" {
			t.Fatalf("public evidence %s status=%d headers=%v", path, response.StatusCode, response.Header)
		}
		if path != base+"/verdict" && response.Header.Get(publicread.EvidenceHashHeader) != digest {
			t.Fatalf("raw evidence %s lacks its exact content digest", path)
		}
		request, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("If-None-Match", response.Header.Get("ETag"))
		cached, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		cachedBody, err := io.ReadAll(cached.Body)
		_ = cached.Body.Close()
		if err != nil || cached.StatusCode != http.StatusNotModified || len(cachedBody) != 0 ||
			cached.Header.Get("ETag") != response.Header.Get("ETag") || cached.Header.Get("Cache-Control") != response.Header.Get("Cache-Control") ||
			cached.Header.Get(publicread.EvidenceHashHeader) != response.Header.Get(publicread.EvidenceHashHeader) {
			t.Fatalf("public evidence conditional read %s: status=%d headers=%v err=%v", path, cached.StatusCode, cached.Header, err)
		}
		return body
	}
	manifestBytes := download(base+"/verdict", "application/json")
	var manifest struct {
		CatalogURL      string `json:"catalog_url"`
		ConstantsHash   string `json:"constants_hash"`
		EngineVersion   string `json:"engine_version"`
		GenesisSHA256   string `json:"genesis_sha256"`
		GenesisURL      string `json:"genesis_url"`
		ReplayLogSHA256 string `json:"replay_log_sha256"`
		ReplayLogURL    string `json:"replay_log_url"`
		RunID           string `json:"run_id"`
		Verdict         string `json:"verdict"`
	}
	decoder := json.NewDecoder(bytes.NewReader(manifestBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil || manifest.RunID != stream+":1" || manifest.Verdict != "verified" ||
		manifest.ConstantsHash != bundle.ConstantsHash || manifest.CatalogURL != "/api/public/v1/catalogs/"+bundle.ConstantsHash ||
		manifest.GenesisURL != base+"/genesis" || manifest.ReplayLogURL != base+"/replay-log" {
		t.Fatalf("public manifest differs from pinned identity/references: %+v err=%v", manifest, err)
	}
	servedGenesis := download(manifest.GenesisURL, "application/json")
	servedArchive := download(manifest.ReplayLogURL, "application/gzip")
	var storedArchive []byte
	var storedHash string
	if err := db.QueryRowContext(context.Background(), `SELECT bytes,sha256 FROM run_log_archive WHERE company_stream_id=$1 AND run_seq=1`, stream).Scan(&storedArchive, &storedHash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(servedGenesis, genesis) || !bytes.Equal(servedArchive, storedArchive) || manifest.ReplayLogSHA256 != storedHash ||
		manifest.GenesisSHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(servedGenesis)) || manifest.ReplayLogSHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(servedArchive)) {
		t.Fatal("downloaded evidence differs from immutable storage/manifest hashes")
	}
	reader, err := gzip.NewReader(bytes.NewReader(servedArchive))
	if err != nil {
		t.Fatal(err)
	}
	archiveBytes, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range privateValues {
		if value != "" && (bytes.Contains(archiveBytes, []byte(value)) || bytes.Contains(servedGenesis, []byte(value)) || bytes.Contains(manifestBytes, []byte(value))) {
			t.Fatalf("ranked public evidence leaked account/session value %s", name)
		}
	}
	var archive struct {
		CompanyStreamID string `json:"company_stream_id"`
		RunSeq          int64  `json:"run_seq"`
		Pin             struct {
			EngineVersion string `json:"engine_version"`
			ConstantsHash string `json:"constants_hash"`
		} `json:"pin"`
		Genesis struct {
			Version int             `json:"version"`
			State   json.RawMessage `json:"state"`
		} `json:"genesis"`
		Entries []struct {
			Sequence         int64           `json:"seq"`
			CanonicalPayload json.RawMessage `json:"canonical_payload"`
			ReplayInputs     json.RawMessage `json:"replay_inputs"`
			Receipt          json.RawMessage `json:"receipt"`
			Events           []struct {
				StreamID      string          `json:"stream_id"`
				Kind          string          `json:"kind"`
				SchemaVersion int             `json:"schema_version"`
				IntentID      string          `json:"intent_id"`
				Payload       json.RawMessage `json:"payload"`
			} `json:"events"`
		} `json:"entries"`
	}
	// run_genesis stores PostgreSQL's spaced canonical JSON bytes; the existing
	// gzip+json.v1 encoder embeds that JSON value compactly. Keep endpoint/storage
	// byte equality above, and compare the embedded value with exact compaction,
	// not a decode/re-encode that could change keys or numeric tokens.
	var compactGenesis bytes.Buffer
	if err := json.Compact(&compactGenesis, servedGenesis); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(archiveBytes, &archive); err != nil || archive.CompanyStreamID != stream || archive.RunSeq != 1 ||
		archive.Pin.EngineVersion != manifest.EngineVersion || archive.Pin.ConstantsHash != manifest.ConstantsHash || !bytes.Equal(archive.Genesis.State, compactGenesis.Bytes()) || len(archive.Entries) == 0 {
		t.Fatalf("downloaded archive binding: decode=%v stream=%t seq=%d engine=%q/%q constants=%q/%q genesis_bytes_equal=%t genesis_lengths=%d/%d entries=%d",
			err, archive.CompanyStreamID == stream, archive.RunSeq, archive.Pin.EngineVersion, manifest.EngineVersion, archive.Pin.ConstantsHash, manifest.ConstantsHash,
			bytes.Equal(archive.Genesis.State, compactGenesis.Bytes()), len(archive.Genesis.State), compactGenesis.Len(), len(archive.Entries))
	}
	entries := make([]production.ReplayLogEntry, len(archive.Entries))
	for index, entry := range archive.Entries {
		// The archive retains broader Founder history (including Fiscal harvests).
		// Match VerifyStoredRun's replay projection: this Company's events plus
		// founder_advanced, in the archive's existing total order. Byte preservation
		// above still covers the entire archive, not just this kernel projection.
		events := make([]map[string]any, 0, len(entry.Events))
		for _, event := range entry.Events {
			if event.StreamID == stream || event.Kind == "founder_advanced" {
				events = append(events, map[string]any{"kind": event.Kind, "schema_version": event.SchemaVersion, "intent_id": event.IntentID, "payload": event.Payload})
			}
		}
		eventBytes, err := json.Marshal(events)
		if err != nil {
			t.Fatal(err)
		}
		entries[index] = production.ReplayLogEntry{Sequence: entry.Sequence, CanonicalPayload: entry.CanonicalPayload, ReplayInputs: entry.ReplayInputs, ReceiptJSON: entry.Receipt, EventsJSON: eventBytes}
	}
	if verdict := production.VerifyReplayRun(servedGenesis, archive.Genesis.Version, bundle, entries, manifest.ConstantsHash, false); verdict != production.ReplayVerified {
		t.Fatalf("downloaded evidence did not reverify: %s", verdict)
	}
	entries[0].ReceiptJSON = []byte(`{}`)
	if verdict := production.VerifyReplayRun(servedGenesis, archive.Genesis.Version, bundle, entries, manifest.ConstantsHash, false); verdict == production.ReplayVerified {
		t.Fatal("corrupt downloaded receipt was accepted by the replay check")
	}
	t.Log("public verification downloads match immutable storage and reverify; corrupted receipt refuses; pinned catalogs supplied from DB")
}
