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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/production"
	"cloud-clicker/server/publicread"
	"cloud-clicker/server/replaycatalog"
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
		etag := response.Header.Get("ETag")
		for _, condition := range [][]string{{etag}, {"W/" + etag}, {`"other", W/` + etag}, {`"other"`, etag}, {"*"}} {
			request, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("X-Request-ID", "evidence-conditional-check")
			for _, value := range condition {
				request.Header.Add("If-None-Match", value)
			}
			cached, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			cachedBody, err := io.ReadAll(cached.Body)
			_ = cached.Body.Close()
			if err != nil || cached.StatusCode != http.StatusNotModified || len(cachedBody) != 0 ||
				cached.Header.Get("ETag") != etag || cached.Header.Get("Cache-Control") != response.Header.Get("Cache-Control") ||
				cached.Header.Get(publicread.EvidenceHashHeader) != response.Header.Get(publicread.EvidenceHashHeader) ||
				cached.Header.Get("X-Request-ID") != "evidence-conditional-check" ||
				cached.Header.Get("Content-Type") != "" || cached.Header.Get("Content-Length") != "" {
				t.Fatalf("public evidence conditional read %s (%q): status=%d headers=%v err=%v", path, condition, cached.StatusCode, cached.Header, err)
			}
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
	assertTypeScriptPublicReplay(t, server.URL, stream, servedGenesis, servedArchive, bundle)
	entries[0].ReceiptJSON = []byte(`{}`)
	if verdict := production.VerifyReplayRun(servedGenesis, archive.Genesis.Version, bundle, entries, manifest.ConstantsHash, false); verdict == production.ReplayVerified {
		t.Fatal("corrupt downloaded receipt was accepted by the replay check")
	}
	t.Log("public verification downloads match immutable storage and reverify; corrupted receipt refuses; pinned catalogs supplied from DB")
}

// Advance only this disposable database through the real epoch writer. The
// running composition is not restarted and no deployable content is minted.
// Historical evidence must remain pinned even when the active epoch excludes it.
func assertHistoricalPublicVerifiedEvidence(t *testing.T, db *sql.DB, server *httptest.Server, stream string, genesis []byte, bundle production.CatalogBundle, privateValues map[string]string) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	repository, err := leaderboard.NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}
	previous, err := repository.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	nextID := previous.ID + 1
	reference := fmt.Sprintf("changelog/epoch-%d.md", nextID)
	if err := os.Mkdir(filepath.Join(root, "changelog"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, reference), []byte("Test-only historical evidence rollover; no product mint.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	artifacts := make([]leaderboard.Artifact, 0, len(bundle.Artifacts))
	for name, raw := range bundle.Artifacts {
		data := bytes.Clone(raw)
		if name == "economy" {
			// Whitespace changes original-byte identity, not gameplay policy.
			data = append(data, '\n')
		}
		artifacts = append(artifacts, leaderboard.Artifact{Name: name, Bytes: data})
	}
	next, err := repository.MintEpoch(ctx, "Historical public evidence fixture", previous.StartedAt.Add(time.Hour), reference, artifacts)
	if err != nil || next.ID != nextID || len(next.Hashes) != 1 || next.Hashes[0] == bundle.ConstantsHash {
		t.Fatalf("rollover did not create a distinct fixture identity: %+v err=%v", next, err)
	}
	current, err := repository.Current(ctx)
	if err != nil || current.ID != next.ID || len(current.Hashes) != 1 || current.Hashes[0] != next.Hashes[0] {
		t.Fatalf("active database epoch did not change: %+v err=%v", current, err)
	}
	var pinnedEpoch int64
	var pinnedHash string
	var closed bool
	if err := db.QueryRowContext(ctx, `SELECT p.epoch_id,p.constants_hash,e.ended_at IS NOT NULL
FROM run_epochs p JOIN epochs e ON e.epoch_id=p.epoch_id
WHERE p.company_stream_id=$1 AND p.run_seq=1`, stream).Scan(&pinnedEpoch, &pinnedHash, &closed); err != nil ||
		pinnedEpoch >= current.ID || pinnedHash != bundle.ConstantsHash || !closed {
		t.Fatalf("verified run is not pinned to a closed historical epoch: epoch=%d hash=%s closed=%t err=%v", pinnedEpoch, pinnedHash, closed, err)
	}
	// Resolve again from storage after rollover, not the pre-rollover in-memory bundle.
	stored, err := replaycatalog.LoadDatabase(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	historical, ok := stored.ResolveReplayCatalogs(bundle.ConstantsHash)
	if !ok || historical.ConstantsHash != bundle.ConstantsHash || len(historical.Artifacts) != len(bundle.Artifacts) {
		t.Fatal("historical pinned replay bundle unavailable after rollover")
	}
	for name, raw := range bundle.Artifacts {
		if !bytes.Equal(historical.Artifacts[name], raw) {
			t.Fatalf("rollover replaced historical artifact %s", name)
		}
	}
	assertPublicVerifiedEvidence(t, db, server, stream, genesis, historical, privateValues)
	t.Log("historical verified HTTP evidence survives distinct active epoch/hash; DB-pinned replay and corruption refusal: PASS")
}

// The root browser-composition driver enables this seam and requires its PASS
// marker. Ordinary Go/DB lanes do not acquire a Node dependency or claim TS proof.
func assertTypeScriptPublicReplay(t *testing.T, baseURL, stream string, genesis, archive []byte, bundle production.CatalogBundle) {
	t.Helper()
	if os.Getenv("TEST_PUBLIC_REPLAY_CLIENT") != "1" {
		t.Log("TypeScript public download/replay seam not requested (root composed lane owns it)")
		return
	}
	input, err := json.Marshal(struct {
		BaseURL       string            `json:"base_url"`
		Stream        string            `json:"stream"`
		ConstantsHash string            `json:"constants_hash"`
		Genesis       []byte            `json:"genesis"`
		Archive       []byte            `json:"archive"`
		Artifacts     map[string][]byte `json:"artifacts"`
	}{baseURL, stream, bundle.ConstantsHash, genesis, archive, bundle.Artifacts})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "client/tools/public-replay-composed.mjs")
	command.Dir = filepathRoot(t)
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	const passed = "public SDK HTTP downloads → TypeScript replay: verified + receipt/event/gap/constants refusal; DB-pinned catalogs supplied: PASS"
	if err != nil || !bytes.Contains(output, []byte(passed)) {
		t.Fatalf("TypeScript public replay seam failed: %v\n%s", err, output)
	}
	t.Log(passed)
}
