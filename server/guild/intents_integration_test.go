package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type integrationNames struct{}

func (integrationNames) ValidateGuildName(value string) bool { return value == "Small Systems" }

func TestGuildLifecycleConcurrencyAndHistoryIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}

	accounts := []string{
		"018f0000-0000-4000-8000-000000000101",
		"018f0000-0000-4000-8000-000000000102",
		"018f0000-0000-4000-8000-000000000103",
	}
	for _, account := range accounts {
		if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',clock_timestamp()) ON CONFLICT DO NOTHING`, account); err != nil {
			t.Fatal(err)
		}
	}
	catalog, _ := LoadCatalog([]byte(phase0Catalog))
	catalog.MaxMembers = 2
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	service, err := NewService(db, catalog, integrationNames{}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	guildID := "018f0000-0000-7000-8000-000000000101"
	created := guildIntent(guildID, "create_guild", 1, `,"name":"Small Systems","join_policy":"open"`)
	result, err := service.Handle(ctx, accounts[0], []byte(created))
	if err != nil || receiptOutcome(t, result.Receipt) != "applied" {
		t.Fatalf("create=%s err=%v", result.Receipt, err)
	}
	replay, err := service.Handle(ctx, accounts[0], []byte(created))
	if err != nil || !replay.Replay || string(replay.Receipt) != string(result.Receipt) {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}

	var wait sync.WaitGroup
	outcomes := make(chan string, 2)
	for index := 1; index < 3; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			intentID := fmt.Sprintf("018f0000-0000-7000-8000-00000000010%d", index+1)
			joined, joinErr := service.Handle(ctx, accounts[index], []byte(guildIntent(intentID, "join_guild", 1, `,"guild_id":"`+guildID+`"`)))
			if joinErr != nil {
				outcomes <- "error:" + joinErr.Error()
				return
			}
			outcomes <- receiptOutcome(t, joined.Receipt)
		}(index)
	}
	wait.Wait()
	close(outcomes)
	counts := map[string]int{}
	for outcome := range outcomes {
		counts[outcome]++
	}
	if counts["applied"] != 1 || counts["rejected"] != 1 {
		t.Fatalf("outcomes=%v", counts)
	}
	var active int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL`, guildID).Scan(&active); err != nil || active != 2 {
		t.Fatalf("active=%d err=%v", active, err)
	}
	if !service.GuildMember(accounts[0], guildID) {
		t.Fatal("leader is not authorized")
	}

	var joinedAccount string
	if err := db.QueryRowContext(ctx, `SELECT account_id FROM guild_members WHERE guild_id=$1 AND role='member' AND left_at IS NULL`, guildID).Scan(&joinedAccount); err != nil {
		t.Fatal(err)
	}
	leaveID := "018f0000-0000-7000-8000-000000000110"
	left, err := service.Handle(ctx, joinedAccount, []byte(guildIntent(leaveID, "leave_guild", 2, "")))
	if err != nil || receiptOutcome(t, left.Receipt) != "applied" || service.GuildMember(joinedAccount, guildID) {
		t.Fatalf("left=%s err=%v", left.Receipt, err)
	}
	rejoinID := "018f0000-0000-7000-8000-000000000111"
	rejoined, err := service.Handle(ctx, joinedAccount, []byte(guildIntent(rejoinID, "join_guild", 3, `,"guild_id":"`+guildID+`"`)))
	if err != nil || receiptOutcome(t, rejoined.Receipt) != "applied" {
		t.Fatalf("rejoin=%s err=%v", rejoined.Receipt, err)
	}
	var history int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND account_id=$2`, guildID, joinedAccount).Scan(&history); err != nil || history != 2 {
		t.Fatalf("history=%d err=%v", history, err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM guild_members WHERE guild_id=$1 AND account_id=$2`, guildID, joinedAccount); err == nil {
		t.Fatal("membership history deletion succeeded")
	}
}

func guildIntent(intentID, kind string, revision int, fields string) string {
	return fmt.Sprintf(`{"intent_id":"%s","kind":"%s","expected_revision":%d%s}`, intentID, kind, revision, fields)
}

func receiptOutcome(t *testing.T, data json.RawMessage) string {
	t.Helper()
	var receipt struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt.Outcome
}
