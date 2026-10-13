package operations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
	"cloud-clicker/server/transport"
)

func TestDatabaseCollectorIntegrationCountsActualTransportDeadLetters(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := "operations_dead_letters_" + strings.ToLower(rand.Text())
	if len(name) > 63 || !regexp.MustCompile(`^operations_dead_letters_[a-z0-9]+$`).MatchString(name) {
		t.Fatal("invalid generated fixture database name")
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.ExecContext(context.Background(), "DROP DATABASE "+name); err != nil {
			t.Error(err)
		}
	}()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + name
	database, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	fixture := newOperationsDeadLetterFixture(t, ctx, database, registry)
	observe := func(want int) {
		t.Helper()
		text := gatheredText(t, registry)
		for _, expected := range []string{
			fmt.Sprintf("cloud_clicker_dead_letters{queue=\"transport\"} %d\n", want),
			"cloud_clicker_dead_letters{queue=\"verification\"} 0\n",
			"cloud_clicker_dead_letters{queue=\"verification_poison\"} 0\n",
			"cloud_clicker_outbox_pending 0\n",
			"cloud_clicker_database_collection_success 1\n",
		} {
			if !strings.Contains(text, expected) {
				t.Fatalf("actual outbox population not reported: want %q in:\n%s", expected, text)
			}
		}
		for _, private := range []string{fixture.founder, fixture.stream, "operations.dead_letter_fixture"} {
			if strings.Contains(text, private) {
				t.Fatalf("private outbox material exposed in metrics: %q", private)
			}
		}
	}
	fixture.publishHealthy(t, ctx)
	observe(0)
	for want := 1; want <= 3; want++ {
		fixture.failOversized(t, ctx)
		observe(want)
	}
	// A subsequent valid event still publishes, without erasing dead letters.
	fixture.publishHealthy(t, ctx)
	observe(3)
}

// Seed only a stream and compensation events. There is no gameplay intent,
// authenticated client or save-state construction in this operations fixture.
// The current event trigger, Store claims, real Node encoder, relay's five-
// attempt policy and production SQL collector are the subject under test.
type operationsDeadLetterFixture struct {
	database *sql.DB
	relay    *transport.PlayerRelay
	registry *Registry
	founder  string
	stream   string
	hash     string
}

func newOperationsDeadLetterFixture(t *testing.T, ctx context.Context, database *sql.DB, registry *Registry) *operationsDeadLetterFixture {
	t.Helper()
	var pending, dead int
	if err := database.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM transport_player_outbox WHERE published_at IS NULL AND dead_lettered_at IS NULL),
		(SELECT count(*) FROM transport_player_outbox WHERE dead_lettered_at IS NOT NULL)`).Scan(&pending, &dead); err != nil || pending != 0 || dead != 0 {
		t.Fatalf("dead-letter fixture requires an isolated empty outbox: pending=%d dead=%d error=%v", pending, dead, err)
	}
	fixture := &operationsDeadLetterFixture{database: database, registry: registry, hash: save.ConstantsHash([]byte("operations.dead_letter_fixture"))}
	if err := database.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope)
		VALUES('founder',gen_random_uuid(),'company') RETURNING id,owner_id`).Scan(&fixture.stream, &fixture.founder); err != nil {
		t.Fatal(err)
	}
	store, err := save.NewStore(database, operationsUnusedCatalogs{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	path := os.Getenv("TEST_TRANSPORT_POLICY_PATH")
	if path == "" {
		path = "../../balance/transport/phase0.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := transport.LoadPolicy(data)
	if err != nil {
		t.Fatal(err)
	}
	node, err := transport.NewNodeWithMetrics(*policy, operationsUnusedAuth{}, nil, registry.Prometheus())
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := node.Shutdown(shutdown); err != nil {
			t.Error(err)
		}
	})
	fixture.relay, err = transport.NewPlayerRelay(store, node, fixture)
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func (fixture *operationsDeadLetterFixture) ReportRelayInvariant(invariant transport.RelayInvariant) {
	fixture.registry.Increment(invariant.Kind)
}

func (fixture *operationsDeadLetterFixture) enqueue(t *testing.T, ctx context.Context, oversized bool) string {
	t.Helper()
	reason := "operations.dead_letter_fixture"
	if oversized {
		reason += strings.Repeat("x", 1<<20) // greater than every admitted Node message cap
	}
	payload, err := json.Marshal(map[string]string{"compensates_event_id": "01985555-1001-7000-8000-000000000001", "reason_key": reason})
	if err != nil {
		t.Fatal(err)
	}
	var eventID string
	if err := fixture.database.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,constants_hash,payload)
		VALUES($1,1,1,'compensation',$2,$3) RETURNING event_id`, fixture.stream, fixture.hash, payload).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	var retained bool
	if err := fixture.database.QueryRowContext(ctx, `SELECT payload->'payload'=$2::jsonb AND payload->>'cursor_effect'='historical'
		FROM transport_player_outbox WHERE message_kind='event' AND source_id=$1`, eventID, payload).Scan(&retained); err != nil || !retained {
		t.Fatalf("current trigger did not preserve authoritative event: %v", err)
	}
	return eventID
}

func (fixture *operationsDeadLetterFixture) publishHealthy(t *testing.T, ctx context.Context) {
	t.Helper()
	eventID := fixture.enqueue(t, ctx, false)
	if published, err := fixture.relay.Flush(ctx); err != nil || published != 1 {
		t.Fatalf("intact event delivery failed: published=%d error=%v", published, err)
	}
	var published bool
	if err := fixture.database.QueryRowContext(ctx, `SELECT published_at IS NOT NULL AND dead_lettered_at IS NULL AND attempt_count=0
		FROM transport_player_outbox WHERE source_id=$1`, eventID).Scan(&published); err != nil || !published {
		t.Fatalf("valid event not published without failure attempts: %v", err)
	}
}

func (fixture *operationsDeadLetterFixture) failOversized(t *testing.T, ctx context.Context) {
	t.Helper()
	eventID := fixture.enqueue(t, ctx, true)
	for attempt := 1; attempt <= 5; attempt++ {
		if published, err := fixture.relay.Flush(ctx); published != 0 || !errors.Is(err, transport.ErrInvalidPolicy) {
			t.Fatalf("oversized event bypassed real encoder refusal: published=%d error=%v", published, err)
		}
		var attempts int
		var dead, published, claimed, retained bool
		if err := fixture.database.QueryRowContext(ctx, `SELECT o.attempt_count,o.dead_lettered_at IS NOT NULL,
			o.published_at IS NOT NULL,o.claim_token IS NOT NULL,o.payload->'payload'=e.payload
			FROM transport_player_outbox o JOIN events e ON e.event_id=o.source_id WHERE o.source_id=$1`, eventID).
			Scan(&attempts, &dead, &published, &claimed, &retained); err != nil || attempts != attempt || dead != (attempt == 5) || published || claimed || !retained {
			t.Fatalf("actual retry/dead-letter transition wrong: attempt=%d persisted=%d dead=%t published=%t claimed=%t retained=%t error=%v",
				attempt, attempts, dead, published, claimed, retained, err)
		}
	}
}

// Neither dependency is exercised by a claim-only Store / publication-only
// Node. Reject any accidental expansion into state writes or authentication.
type operationsUnusedCatalogs struct{}

func (operationsUnusedCatalogs) Resolve(string) (*economy.Catalog, bool) { return nil, false }

type operationsUnusedAuth struct{}

func (operationsUnusedAuth) Authenticate(context.Context, string) (account.Claims, error) {
	return account.Claims{}, transport.ErrInvalidNode
}
