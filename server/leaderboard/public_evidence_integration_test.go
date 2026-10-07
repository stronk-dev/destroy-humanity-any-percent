package leaderboard

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

// This is a repository/authorization witness with synthetic immutable evidence,
// not a replay-verifier verdict or a public HTTP download acceptance test.
func TestPublicRunEvidenceIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE verification_projection_events,epochs,catalog_sets,accounts,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	const streamID = "01986666-0000-4000-8000-000000000001"
	const founderID = "01986666-0000-4000-8000-000000000002"
	hash := "sha256:" + strings.Repeat("a", 64)
	foreignHash := "sha256:" + strings.Repeat("b", 64)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'company')`, streamID, founderID)
	exec(`INSERT INTO catalog_sets(constants_hash) VALUES($1),($2)`, hash, foreignHash)
	exec(`INSERT INTO epochs(name,started_at,changelog_ref) VALUES('evidence','2026-10-07T00:00:00Z','changelog/epoch-1.md')`)
	exec(`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES(1,$1)`, hash)
	genesis := []byte(`{"v":1,"label":"raw \\ bytes","nested":{"z":2,"a":1}}`)
	var compressed bytes.Buffer
	zipper := gzip.NewWriter(&compressed)
	// Nondefault gzip metadata makes accidental recompression observable.
	zipper.Name = "original-replay.json"
	if _, err := zipper.Write([]byte(`{"schema_version":1,"entries":[]}`)); err != nil {
		t.Fatal(err)
	}
	if err := zipper.Close(); err != nil {
		t.Fatal(err)
	}
	archive := compressed.Bytes()
	archiveHash := fmt.Sprintf("sha256:%x", sha256.Sum256(archive))
	for seq := int64(1); seq <= 7; seq++ {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		seedPublicEvidencePin(t, tx, streamID, seq, hash, foreignHash, genesis)
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		if seq != 4 { // authorized run with missing archive
			storedHash := archiveHash
			if seq == 5 {
				storedHash = "sha256:" + strings.Repeat("f", 64)
			}
			exec(`INSERT INTO run_log_archive(run_id,company_stream_id,run_seq,terminal_seq,encoding,bytes,sha256)
VALUES($1,$2,$3,1,'gzip+json.v1',$4,$5)`, streamID+":"+fmt.Sprint(seq), streamID, seq, archive, storedHash)
		}
		if seq != 2 && seq != 3 {
			seedPublicEvidenceBoard(t, db, streamID, founderID, seq)
		}
	}
	// Queue-only verification must remain private, even with complete archive bytes.
	exec(`INSERT INTO verification_queue(company_stream_id,run_seq,status,verdict) VALUES($1,2,'verified','verified')`, streamID)
	// A board row whose pin is missing is corrupt public evidence, not private/unknown.
	seedPublicEvidenceBoard(t, db, streamID, founderID, 8)
	// Multiple categories must still retrieve a single evidence result.
	exec(`INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,verified_at)
SELECT run_id,event_id,founder_id,'ethical_percent',variables,epoch_id,mandate_level,key_ms,verified_at FROM verified_runs WHERE run_id=$1`, streamID+":1")
	repository, err := NewRepository(db, "../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("exact stored bytes and pinned identity", func(t *testing.T) {
		evidence, err := repository.PublicRunEvidence(ctx, streamID, 1)
		if err != nil || evidence.RunID != streamID+":1" || evidence.EngineVersion != "0.1.0" || evidence.ConstantsHash != hash || evidence.GenesisVersion != 1 ||
			!bytes.Equal(evidence.Genesis, genesis) || !bytes.Equal(evidence.ReplayLog, archive) ||
			evidence.GenesisSHA256 != fmt.Sprintf("sha256:%x", sha256.Sum256(genesis)) || evidence.ReplayLogSHA256 != archiveHash {
			t.Fatalf("public evidence differs from exact stored bytes/pin: %+v err=%v", evidence, err)
		}
		evidence.Genesis[0], evidence.ReplayLog[0] = 'X', 'X'
		again, err := repository.PublicRunEvidence(ctx, streamID, 1)
		if err != nil || !bytes.Equal(again.Genesis, genesis) || !bytes.Equal(again.ReplayLog, archive) {
			t.Fatal("caller mutation changed stored evidence")
		}
	})
	for _, test := range []struct {
		name string
		seq  int64
		want error
	}{
		{"queue verified without a board record", 2, ErrUnknownPublicRun},
		{"archive without a board record", 3, ErrUnknownPublicRun},
		{"public run without an archive", 4, ErrInvalidPublicRunEvidence},
		{"public run with corrupt archive hash", 5, ErrInvalidPublicRunEvidence},
		{"genesis differs from pinned constants", 6, ErrInvalidPublicRunEvidence},
		{"public run with invalid genesis JSON", 7, ErrInvalidPublicRunEvidence},
		{"public record with missing pin", 8, ErrInvalidPublicRunEvidence},
		{"unknown run", 9, ErrUnknownPublicRun},
		{"invalid sequence", 0, ErrUnknownPublicRun},
		{"inexact sequence", decimal.MaxExactInteger + 1, ErrUnknownPublicRun},
	} {
		t.Run(test.name, func(t *testing.T) {
			evidence, err := repository.PublicRunEvidence(ctx, streamID, test.seq)
			if !errors.Is(err, test.want) || evidence.RunID != "" || len(evidence.Genesis) != 0 || len(evidence.ReplayLog) != 0 {
				t.Fatalf("rejected read leaked evidence or wrong error: %+v err=%v want=%v", evidence, err, test.want)
			}
		})
	}
	t.Run("database cancellation remains an operational error", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		evidence, err := repository.PublicRunEvidence(cancelled, streamID, 1)
		if !errors.Is(err, context.Canceled) || evidence.RunID != "" {
			t.Fatalf("cancelled read must not become an unknown run/verdict: %+v %v", evidence, err)
		}
	})
}

func seedPublicEvidencePin(t *testing.T, tx *sql.Tx, streamID string, seq int64, hash, foreignHash string, genesis []byte) {
	t.Helper()
	// The transaction is caller-owned; test cleanup also rolls back on a failure.
	t.Cleanup(func() { _ = tx.Rollback() })
	if _, err := tx.Exec(`INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
VALUES($1,$2,1,$3,'0.1.0','evidence-test','1')`, streamID, seq, hash); err != nil {
		t.Fatal(err)
	}
	genesisHash := hash
	if seq == 6 {
		genesisHash = foreignHash
	}
	if seq == 7 {
		genesis = []byte("not JSON")
	}
	if _, err := tx.Exec(`INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash) VALUES($1,$2,$3,1,$4)`, streamID, seq, genesis, genesisHash); err != nil {
		t.Fatal(err)
	}
}

func seedPublicEvidenceBoard(t *testing.T, db *sql.DB, streamID, founderID string, seq int64) {
	t.Helper()
	eventID := fmt.Sprintf("01986666-0000-7000-8000-%012d", seq)
	if _, err := db.Exec(`INSERT INTO verification_projection_events(event_id) VALUES($1)`, eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,verified_at)
VALUES($1,$2,$3,'any_percent','{"commons":false,"advisor":false,"glitched":false,"faction":null}',1,0,1000,now())`, streamID+":"+fmt.Sprint(seq), eventID, founderID); err != nil {
		t.Fatal(err)
	}
}
