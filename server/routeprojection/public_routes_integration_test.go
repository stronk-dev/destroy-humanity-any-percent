package routeprojection

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestPublicRoutePageIntegration(t *testing.T) {
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
	truncateRouteProjection(t, ctx, db)
	if _, err := db.ExecContext(ctx, `TRUNCATE account_founders,accounts CASCADE`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := routes.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const hash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	projector, err := New(db, testCatalogs{hash: catalog})
	if err != nil {
		t.Fatal(err)
	}
	const (
		active     = "33333333-3333-4333-8333-333333333333" // owned by a live account
		anonymized = "44444444-4444-4444-8444-444444444444" // account deleted: account_id NULL
		orphan     = "55555555-5555-4555-8555-555555555555" // no ownership row at all
		accountID  = "66666666-6666-4666-8666-666666666666"
	)
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, []any{accountID}},
		{`INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, []any{accountID, active}},
		{`INSERT INTO accounts(account_id,recovery_hash) VALUES('77777777-7777-4777-8777-777777777777','test')`, nil},
		{`INSERT INTO account_founders(account_id,founder_id) VALUES('77777777-7777-4777-8777-777777777777',$1)`, []any{anonymized}},
		{`DELETE FROM accounts WHERE account_id='77777777-7777-4777-8777-777777777777'`, nil},
	} {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	occurred := time.Date(2026, 7, 29, 14, 0, 0, 0, time.UTC)
	executions := []struct{ event, founder, route string }{
		{"10000000-0000-4000-8000-000000000001", active, "route.nonprofit_wrapper_zip"},
		{"20000000-0000-4000-8000-000000000002", anonymized, "route.acquihire_out_of_bounds"},
		{"30000000-0000-4000-8000-000000000003", orphan, "route.ipo_sequence_break"},
	}
	for index, execution := range executions {
		const gate = "gate.t4_to_t5" // all three routes discount the T4→T5 gate
		stream := insertStream(t, ctx, db, execution.founder, "company")
		record := insertExecution(t, ctx, db, execution.event, stream, execution.founder, 2, 1, execution.route, gate, hash, occurred.Add(time.Duration(index)*time.Minute))
		if err := projector.Project(ctx, []save.EventRecord{record}); err != nil {
			t.Fatal(err)
		}
	}
	// active publishes an approved name; anonymized submits one still pending.
	if err := projector.SubmitName(ctx, "route.nonprofit_wrapper_zip", active, "Approved Public Name", occurred.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := projector.ResolveName(ctx, "route.nonprofit_wrapper_zip", true); err != nil {
		t.Fatal(err)
	}
	if err := projector.SubmitName(ctx, "route.acquihire_out_of_bounds", anonymized, "Unmoderated Name", occurred.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	first, more, err := projector.PublicRoutePage(ctx, "", 2)
	if err != nil || !more || len(first) != 2 {
		t.Fatalf("first=%+v more=%v err=%v", first, more, err)
	}
	if first[0].RouteID != "route.acquihire_out_of_bounds" || first[1].RouteID != "route.ipo_sequence_break" {
		t.Fatalf("routes are not route_id-ordered: %s, %s", first[0].RouteID, first[1].RouteID)
	}
	if first[0].NamingStatus != "pending" || first[0].PublicName == "Unmoderated Name" || first[0].FirstExecutor != nil {
		t.Fatalf("pending name or anonymized executor leaked: %+v", first[0])
	}
	if first[1].FirstExecutor != nil {
		t.Fatalf("an orphaned executor must be withheld: %+v", first[1])
	}
	second, more, err := projector.PublicRoutePage(ctx, first[1].RouteID, 2)
	if err != nil || more || len(second) != 1 || second[0].RouteID != "route.nonprofit_wrapper_zip" {
		t.Fatalf("second=%+v more=%v err=%v", second, more, err)
	}
	if second[0].NamingStatus != "published" || second[0].PublicName != "Approved Public Name" || second[0].FirstExecutor == nil || *second[0].FirstExecutor != active ||
		second[0].AdoptionCount != 1 || !second[0].CreditedAt.Equal(occurred) || !second[0].NamingDeadline.Equal(occurred.Add(72*time.Hour)) {
		t.Fatalf("published route row: %+v", second[0])
	}
	if exact, more, err := projector.PublicRoutePage(ctx, "", 3); err != nil || more || len(exact) != 3 {
		t.Fatalf("an exactly-full page must not claim more: %d %v %v", len(exact), more, err)
	}
}
