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

type integrationWindow struct{}

func (integrationWindow) GuildHealthWindowMS(string) (int64, bool) {
	return int64(30 * 24 * time.Hour / time.Millisecond), true
}

func (integrationNames) ValidateGuildName(value string) bool {
	return value == "small systems" || value == "small systems 2"
}

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
	createIntentID := "018f0000-0000-7000-8000-000000000101"
	created := guildIntent(createIntentID, "create_guild", 1, `,"name":"Small Systems","join_policy":"open"`)
	result, err := service.Handle(ctx, accounts[0], []byte(created))
	if err != nil || receiptOutcome(t, result.Receipt) != "applied" {
		t.Fatalf("create=%s err=%v", result.Receipt, err)
	}
	replay, err := service.Handle(ctx, accounts[0], []byte(created))
	if err != nil || !replay.Replay || string(replay.Receipt) != string(result.Receipt) {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
	guildID := receiptGuildID(t, result.Receipt)
	if guildID == createIntentID || !uuidV7Pattern.MatchString(guildID) {
		t.Fatalf("guild id was not independently server-generated: %q", guildID)
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

	sweepAccount := "018f0000-0000-4000-8000-000000000104"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',clock_timestamp()) ON CONFLICT DO NOTHING`, sweepAccount); err != nil {
		t.Fatal(err)
	}
	sweepCreateID := "018f0000-0000-7000-8000-000000000104"
	sweepCreated, err := service.Handle(ctx, sweepAccount, []byte(guildIntent(sweepCreateID, "create_guild", 1, `,"name":"Small Systems 2","join_policy":"open"`)))
	if err != nil || receiptOutcome(t, sweepCreated.Receipt) != "applied" {
		t.Fatalf("sweep create=%s err=%v", sweepCreated.Receipt, err)
	}
	sweepGuild := receiptGuildID(t, sweepCreated.Receipt)
	if count, err := service.SweepDisbanded(ctx, now.Add(7*24*time.Hour-time.Millisecond), 10); err != nil || count != 0 {
		t.Fatalf("early sweep=%d err=%v", count, err)
	}
	if count, err := service.SweepDisbanded(ctx, now.Add(7*24*time.Hour), 10); err != nil || count != 1 {
		t.Fatalf("due sweep=%d err=%v", count, err)
	}
	if service.GuildMember(sweepAccount, sweepGuild) {
		t.Fatal("swept guild retained membership")
	}
	var sweptLeaves int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_events WHERE guild_id=$1 AND kind='member_left'`, sweepGuild).Scan(&sweptLeaves); err != nil || sweptLeaves != 1 {
		t.Fatalf("swept member_left events=%d err=%v", sweptLeaves, err)
	}

	founderID := "018f0000-0000-7000-8000-000000000201"
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accounts[0], founderID, now); err != nil {
		t.Fatal(err)
	}
	projector, err := NewProjector(db, integrationWindow{})
	if err != nil {
		t.Fatal(err)
	}
	event := save.EventRecord{EventID: "018f0000-0000-4000-8000-000000000301", StreamID: "018f0000-0000-7000-8000-000000000301", OwnerID: founderID,
		Revision: 2, Kind: save.EventGuildTitheAccrued, IntentID: "018f0000-0000-7000-8000-000000000302", ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", OccurredAt: now,
		Payload: json.RawMessage(`{"founder_id":"` + founderID + `","run_id":{"company_stream_id":"018f0000-0000-7000-8000-000000000301","run_seq":1},"progress_delta_ppm":500000,"xp_delta":10}`)}
	if err := projector.Project(ctx, []save.EventRecord{event, event}); err != nil {
		t.Fatal(err)
	}
	var guildXP int64
	if err := db.QueryRowContext(ctx, `SELECT guild_xp FROM guilds WHERE guild_id=$1`, guildID).Scan(&guildXP); err != nil || guildXP != 10 {
		t.Fatalf("guild xp=%d err=%v", guildXP, err)
	}

	if err := service.CommitClearingBoundary(ctx, guildID, 1, now.Add(time.Minute), []MemberStock{
		{AccountID: accounts[0], Produces: "libraries", Consumes: "carbon", AvailableUnits: 10},
		{AccountID: joinedAccount, Produces: "carbon", Consumes: "libraries", AvailableUnits: 10},
	}, 100); err != nil {
		t.Fatal(err)
	}
	pending, err := service.PendingSettlements(ctx, founderID, 0)
	if err != nil || len(pending) != 1 || pending[0].DebitUnits != 5 || pending[0].CreditUnits != 5 {
		t.Fatalf("pending=%+v err=%v", pending, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareAccountDeletion(ctx, tx, accounts[0], now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, accounts[0]); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var successorRole string
	if err := db.QueryRowContext(ctx, `SELECT role FROM guild_members WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL`, guildID, joinedAccount).Scan(&successorRole); err != nil || successorRole != "leader" {
		t.Fatalf("successor role=%q err=%v", successorRole, err)
	}
	var anonymized int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND account_id IS NULL AND left_at IS NOT NULL`, guildID).Scan(&anonymized); err != nil || anonymized != 1 {
		t.Fatalf("anonymized memberships=%d err=%v", anonymized, err)
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

func receiptGuildID(t *testing.T, data json.RawMessage) string {
	t.Helper()
	var receipt struct {
		GuildID string `json:"guild_id"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil || receipt.GuildID == "" {
		t.Fatalf("receipt guild id: data=%s err=%v", data, err)
	}
	return receipt.GuildID
}
