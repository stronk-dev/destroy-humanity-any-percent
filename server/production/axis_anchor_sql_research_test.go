package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

// Test-only logical jsonb reader. This does not adopt a save or transport policy.
func anchorSQLResearchRestore(encoded []byte) (anchorResearchSnapshot, error) {
	var snapshot anchorResearchSnapshot
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return snapshot, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return snapshot, fmt.Errorf("invalid logical research framing")
	}
	canonical, err := json.Marshal(snapshot)
	if err != nil {
		return snapshot, err
	}
	return anchorResearchRestore(canonical)
}

type anchorSQLObservation struct {
	ID             string `json:"id"`
	NormalizedJSON string `json:"normalized_json"`
	StrictAccepted bool   `json:"strict_accepted"`
	LogicalExact   bool   `json:"logical_exact"`
}

type anchorSQLNegative struct {
	ID             string `json:"id"`
	SQLState       string `json:"sqlstate"`
	NormalizedJSON string `json:"normalized_json"`
	StrictAccepted bool   `json:"strict_accepted"`
	LogicalAccept  bool   `json:"logical_accepted"`
}

type anchorSQLReport struct {
	Version          int                    `json:"version"`
	Acceptance       string                 `json:"acceptance_status"`
	ServerVersion    string                 `json:"postgres_version"`
	Runtime          string                 `json:"runtime"`
	Sources          map[string]string      `json:"source_sha256"`
	Rows             []anchorSQLObservation `json:"valid_rows"`
	Negatives        []anchorSQLNegative    `json:"negative_rows"`
	StrictRefused    int                    `json:"strict_valid_refused"`
	LogicalPreserved int                    `json:"logical_valid_preserved"`
}

const anchorSQLResearchPath = "../../testdata/axis-stack/anchor-sql-research-v1.json"

// Environment is recorded provenance, not a golden outcome. All other fields
// remain exact; a different observation is never accepted as platform variance.
func anchorSQLComparable(report anchorSQLReport) (anchorSQLReport, error) {
	if !strings.HasPrefix(report.ServerVersion, "16.") ||
		(report.Runtime != "linux/arm64" && report.Runtime != "linux/amd64") {
		return report, fmt.Errorf("unsupported research environment")
	}
	report.ServerVersion, report.Runtime = "", ""
	return report, nil
}

func TestAxisAnchorSQLResearchEnvironmentComparison(t *testing.T) {
	local := anchorSQLReport{ServerVersion: "16.15", Runtime: "linux/arm64", LogicalPreserved: 1215}
	other := local
	other.ServerVersion, other.Runtime = "16.16", "linux/amd64"
	a, errA := anchorSQLComparable(local)
	b, errB := anchorSQLComparable(other)
	if errA != nil || errB != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("environment-only change contaminated semantic equality")
	}
	other.LogicalPreserved--
	b, errB = anchorSQLComparable(other)
	if errB != nil || reflect.DeepEqual(a, b) {
		t.Fatal("semantic change hidden by environment comparison")
	}
	for _, invalid := range []anchorSQLReport{
		{ServerVersion: "17.1", Runtime: "linux/arm64"},
		{ServerVersion: "16.15", Runtime: "darwin/arm64"},
		{ServerVersion: "16.15", Runtime: "linux/unknown"},
	} {
		if _, err := anchorSQLComparable(invalid); err == nil {
			t.Fatal("invalid environment admitted")
		}
	}
}

func anchorSQLResearchArtifact(t *testing.T, report anchorSQLReport) {
	t.Helper()
	if _, err := anchorSQLComparable(report); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("UPDATE_ANCHOR_SQL_RESEARCH") == "1" {
		sourceResearchArtifact(t, anchorSQLResearchPath, "UPDATE_ANCHOR_SQL_RESEARCH", report)
		return
	}
	encoded, err := os.ReadFile(anchorSQLResearchPath)
	if err != nil {
		t.Fatal(err)
	}
	var recorded anchorSQLReport
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&recorded); err != nil || decoder.Decode(new(any)) != io.EOF {
		t.Fatalf("invalid recorded SQL observation: %v", err)
	}
	if _, err := anchorSQLComparable(recorded); err != nil {
		t.Fatal(err)
	}
	// Explicitly compare at the recorded environment identity, not as a claim
	// that this replay executed there. The test log retains actual environment.
	report.ServerVersion, report.Runtime = recorded.ServerVersion, recorded.Runtime
	sourceResearchArtifact(t, anchorSQLResearchPath, "UPDATE_ANCHOR_SQL_RESEARCH", report)
}

func TestAxisAnchorSQLIntegrationResearch(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		if os.Getenv("UPDATE_ANCHOR_SQL_RESEARCH") == "1" {
			t.Fatal("invalid SQL measurement: declared database unavailable")
		}
		t.Skip("NOT EXECUTED: requires declared Postgres16 service")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	db, err := save.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal("declared database connection failed") // Never disclose credentials.
	}
	defer db.Close()
	var report anchorSQLReport
	report.Version, report.Acceptance = 1, "NOT_PROVEN: production AC6 remains red"
	report.Runtime = runtime.GOOS + "/" + runtime.GOARCH
	if err := db.QueryRowContext(ctx, "SHOW server_version").Scan(&report.ServerVersion); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(report.ServerVersion, "16.") {
		t.Fatal("invalid SQL measurement: requires Postgres16")
	}
	report.Sources = sourceResearchHashes(t, []string{
		"server/production/axis_anchor_sql_research_test.go", "server/production/axis_rate_source_research_test.go",
		"server/production/axis_anchor_research_test.go", "testdata/axis-stack/anchor-research-v1.json",
		"server/production/accrual.go", "server/decimal/decimal.go", "server/decimal/canonical.go",
		"server/save/database.go", "compose.save-test.yml", "Makefile",
	})
	encoded, err := os.ReadFile(anchorResearchPath)
	if err != nil {
		t.Fatal(err)
	}
	var previous anchorResearchReport
	if err := json.Unmarshal(encoded, &previous); err != nil || previous.Version != 1 {
		t.Fatalf("invalid frozen population: %v", err)
	}
	type sample struct {
		id    string
		value anchorResearchSnapshot
	}
	var population []sample
	for _, row := range previous.Cases {
		if row.ExpectedError {
			if row.One != nil || row.Split != nil {
				t.Fatal("refused population contains snapshots")
			}
			continue
		}
		if row.One == nil || row.Split == nil {
			t.Fatal("missing frozen snapshots")
		}
		population = append(population, sample{row.ID + "/one", *row.One}, sample{row.ID + "/split", *row.Split})
	}
	for _, row := range previous.Boundaries {
		population = append(population, sample{row.ID + "/before", row.Before}, sample{row.ID + "/after", row.After})
	}
	for i, value := range previous.NearCapAfter {
		population = append(population, sample{fmt.Sprintf("near-cap/%d/post-debit", i), value})
	}
	if len(population) != 1215 || len(previous.Negatives) != 16 {
		t.Fatal("incomplete frozen SQL population")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() // Transaction-local table only; no live tables or migrations.
	if _, err := tx.ExecContext(ctx, "CREATE TEMP TABLE axis_anchor_research (ordinal integer PRIMARY KEY, snapshot jsonb NOT NULL) ON COMMIT DROP"); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for ordinal, row := range population {
		if seen[row.id] {
			t.Fatal("duplicate SQL sample identity")
		}
		seen[row.id] = true
		payload, err := json.Marshal(row.value)
		if err != nil {
			t.Fatal(err)
		}
		if original, err := anchorResearchRestore(payload); err != nil || !reflect.DeepEqual(original, row.value) {
			t.Fatalf("%s original snapshot invalid: %v", row.id, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO axis_anchor_research VALUES ($1,$2::jsonb)", ordinal, string(payload)); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := tx.QueryContext(ctx, "SELECT ordinal,snapshot::text FROM axis_anchor_research ORDER BY ordinal")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var ordinal int
		var normalized string
		if err := rows.Scan(&ordinal, &normalized); err != nil {
			t.Fatal(err)
		}
		if ordinal != len(report.Rows) || ordinal >= len(population) {
			t.Fatal("SQL ordinal population drift")
		}
		row := population[ordinal]
		_, strictErr := anchorResearchRestore([]byte(normalized))
		logical, logicalErr := anchorSQLResearchRestore([]byte(normalized))
		exact := logicalErr == nil && reflect.DeepEqual(logical, row.value)
		if !exact {
			t.Fatalf("%s logical SQL state changed: %v", row.id, logicalErr)
		}
		if strictErr != nil {
			report.StrictRefused++
		}
		report.LogicalPreserved++
		report.Rows = append(report.Rows, anchorSQLObservation{row.id, normalized, strictErr == nil, exact})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(report.Rows) != 1215 {
		t.Fatal("truncated SQL population")
	}
	var names []string
	for name := range previous.Negatives {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := anchorResearchRestore([]byte(previous.Negatives[name])); err == nil {
			t.Fatalf("%s ceased being a strict negative", name)
		}
		if _, err := tx.ExecContext(ctx, "SAVEPOINT negative_payload"); err != nil {
			t.Fatal(err)
		}
		row := anchorSQLNegative{ID: name}
		err := tx.QueryRowContext(ctx, "SELECT $1::jsonb::text", previous.Negatives[name]).Scan(&row.NormalizedJSON)
		if err != nil {
			var sqlError interface{ SQLState() string }
			if !errors.As(err, &sqlError) || sqlError.SQLState() != "22P02" || name != "trailing JSON" {
				t.Fatalf("%s invalid SQL observation: %v", name, err)
			}
			row.SQLState = sqlError.SQLState()
			if _, err := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT negative_payload"); err != nil {
				t.Fatal(err)
			}
		} else {
			row.SQLState = "00000"
			_, strictErr := anchorResearchRestore([]byte(row.NormalizedJSON))
			_, logicalErr := anchorSQLResearchRestore([]byte(row.NormalizedJSON))
			row.StrictAccepted, row.LogicalAccept = strictErr == nil, logicalErr == nil
			// SQL has discarded these representation-only differences, not validated transport.
			representationOnly := name == "duplicate wire" || name == "noncanonical field order" || name == "length513"
			if row.LogicalAccept != representationOnly || name == "trailing JSON" {
				t.Fatalf("%s unexpected logical SQL admission: %v", name, logicalErr)
			}
		}
		if _, err := tx.ExecContext(ctx, "RELEASE SAVEPOINT negative_payload"); err != nil {
			t.Fatal(err)
		}
		report.Negatives = append(report.Negatives, row)
	}
	if len(report.Negatives) != 16 {
		t.Fatal("truncated SQL negatives")
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	anchorSQLResearchArtifact(t, report)
	t.Logf("R-012 real %s %s:1215 complete; strict refused%d, logical exact%d;16 SQL negatives complete; AC6 NOT_PROVEN", report.ServerVersion, report.Runtime, report.StrictRefused, report.LogicalPreserved)
}
