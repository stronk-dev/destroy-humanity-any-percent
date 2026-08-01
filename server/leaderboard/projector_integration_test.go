package leaderboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/kernel"
	"cloud-clicker/server/save"
)

const projectorHash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func TestQueueProjectorCategoriesVariablesPreTimerAndRetryIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE verification_projection_events,accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	categoryBytes, err := os.ReadFile("../../testdata/leaderboards/l7a-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCategoryCatalog(categoryBytes, []string{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"})
	if err != nil {
		t.Fatal(err)
	}
	projector, err := NewQueueProjector(catalog)
	if err != nil {
		t.Fatal(err)
	}
	accountID := "41111111-1111-4111-8111-111111111111"
	founderID := "42222222-2222-4222-8222-222222222222"
	streamID := "43333333-3333-4333-8333-333333333333"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'integration')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'company')`, streamID, founderID); err != nil {
		t.Fatal(err)
	}
	seedProjectorEpoch(t, db, streamID, 1)
	intentID := "01989999-1001-7000-8000-000000000001"
	if _, err := db.ExecContext(ctx, `INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,applied_revision,server_ts_ms)
		VALUES($1,1,1,$2,'{}','{}','{}',1,100)`, streamID, intentID); err != nil {
		t.Fatal(err)
	}
	runID := map[string]any{"company_stream_id": streamID, "run_seq": 1}
	incorporated, _ := json.Marshal(map[string]any{"run_id": runID, "faction_id": "open_source"})
	route, _ := json.Marshal(map[string]any{"run_id": runID, "route_id": "route.example"})
	ended := projectorRunEndedPayload(streamID, founderID, 1, false, []string{"route.example"}, "open_source")
	for _, event := range []struct {
		kind   string
		schema int
		body   []byte
	}{{"incorporated", 1, incorporated}, {"route_executed", 1, route}, {"run_ended", 2, ended}} {
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload)
			VALUES($1,1,$2,$3,$4,$5,'2026-08-01T12:00:00Z',$6)`, streamID, event.schema, event.kind, intentID, projectorHash, event.body); err != nil {
			t.Fatal(err)
		}
	}
	projectInTransaction(t, db, projector, streamID, 1)
	projectInTransaction(t, db, projector, streamID, 1)

	rows, err := db.QueryContext(ctx, `SELECT category_id,key_ms,variables,world_first FROM verified_runs WHERE run_id=$1 ORDER BY category_id`, streamID+":1")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var category, variables string
		var key int64
		var worldFirst bool
		if err := rows.Scan(&category, &key, &variables, &worldFirst); err != nil {
			t.Fatal(err)
		}
		if !worldFirst || category == "ethical_percent" && key != 90 || category != "ethical_percent" && key != 100 ||
			variables != `{"advisor": false, "commons": true, "faction": "open_source", "glitched": true}` {
			t.Fatalf("category=%s key=%d variables=%s world_first=%v", category, key, variables, worldFirst)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !sameStrings(categories, []string{"any_percent", "ethical_percent", "hundred_percent", "low_percent"}) {
		t.Fatalf("categories=%v", categories)
	}

	seedProjectorEpoch(t, db, streamID, 2)
	preTimerIntent := "01989999-1002-7000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,applied_revision,server_ts_ms)
		VALUES($1,2,1,$2,'{}','{}','{}',1,200)`, streamID, preTimerIntent); err != nil {
		t.Fatal(err)
	}
	preTimerEnded := projectorRunEndedPayload(streamID, founderID, 2, true, nil, "")
	if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload)
		VALUES($1,2,2,'run_ended',$2,$3,'2026-08-01T12:01:00Z',$4)`, streamID, preTimerIntent, projectorHash, preTimerEnded); err != nil {
		t.Fatal(err)
	}
	projectInTransaction(t, db, projector, streamID, 2)
	var boardRows, claims int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, streamID+":2").Scan(&boardRows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM verification_projection_events`).Scan(&claims); err != nil {
		t.Fatal(err)
	}
	if boardRows != 0 || claims != 2 {
		t.Fatalf("pre-timer rows=%d claims=%d", boardRows, claims)
	}

	seedProjectorEpoch(t, db, streamID, 3)
	driftIntent := "01989999-1003-7000-8000-000000000003"
	if _, err := db.ExecContext(ctx, `INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,applied_revision,server_ts_ms)
		VALUES($1,3,1,$2,'{}','{}','{}',1,300)`, streamID, driftIntent); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO run_version_drift(company_stream_id,run_seq,observed_version) VALUES($1,3,'0.0.1')`, streamID); err != nil {
		t.Fatal(err)
	}
	driftEnded := projectorRunEndedPayload(streamID, founderID, 3, false, nil, "")
	if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload)
		VALUES($1,3,2,'run_ended',$2,$3,'2026-08-01T12:02:00Z',$4)`, streamID, driftIntent, projectorHash, driftEnded); err != nil {
		t.Fatal(err)
	}
	projectInTransaction(t, db, projector, streamID, 3)
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM verified_runs WHERE run_id=$1`, streamID+":3").Scan(&boardRows); err != nil || boardRows != 0 {
		t.Fatalf("drifted run rows=%d err=%v", boardRows, err)
	}
}

func seedProjectorEpoch(t *testing.T, db *sql.DB, streamID string, runSeq int64) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, projectorHash); err != nil {
		t.Fatal(err)
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id FROM epochs WHERE ended_at IS NULL`).Scan(&epochID); err == sql.ErrNoRows {
		if err := tx.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('projector','2026-08-01T00:00:00Z','changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2) ON CONFLICT DO NOTHING`, epochID, projectorHash); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
		VALUES($1,$2::bigint,$3,$4,$5,'integration',$2::bigint::text)`, streamID, runSeq, epochID, projectorHash, kernel.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash) VALUES($1,$2,'{}',1,$3)`, streamID, runSeq, projectorHash); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func projectorRunEndedPayload(streamID, founderID string, runSeq int64, preTimer bool, routes []string, faction string) []byte {
	var factionValue any
	if faction != "" {
		factionValue = faction
	}
	payload, _ := json.Marshal(map[string]any{
		"founder_id": founderID, "run_id": map[string]any{"company_stream_id": streamID, "run_seq": runSeq}, "exit_type": "collapse",
		"started_at_ms": 100, "ended_at_ms": 200, "rta_ms": 100, "attended_ms": 90, "pre_timer": preTimer, "terminal_seq": 1,
		"payout": map[string]any{}, "tier": 3, "lifetime_value": "1e3", "ledger_fact_kinds": []string{"fact.completion"},
		"executed_routes": routes, "gates_crossed": []string{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"},
		"generators_purchased_total": 10, "assisted": map[string]bool{"commons": true, "advisor": false}, "faction": factionValue,
	})
	return payload
}

func projectInTransaction(t *testing.T, db *sql.DB, projector *QueueProjector, streamID string, runSeq int64) {
	t.Helper()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := projector.ProjectVerifiedRun(context.Background(), tx, streamID, runSeq); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}
