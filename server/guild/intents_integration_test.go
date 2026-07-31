package guild

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	return strings.HasPrefix(value, "small systems")
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
	if _, err := db.ExecContext(ctx, `INSERT INTO guild_applications(guild_id,account_id,created_at) VALUES($1,$2,$3)`, sweepGuild, accounts[1], now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO guild_invitations(guild_id,account_id,created_at) VALUES($1,$2,$3)`, sweepGuild, accounts[2], now); err != nil {
		t.Fatal(err)
	}
	if count, err := service.SweepDisbanded(ctx, now.Add(7*24*time.Hour-time.Millisecond), 10); err != nil || count != 0 {
		t.Fatalf("early sweep=%d err=%v", count, err)
	}
	if count, err := service.SweepDisbanded(ctx, now.Add(7*24*time.Hour), 10); err != nil || count != 1 {
		t.Fatalf("due sweep=%d err=%v", count, err)
	}
	if service.GuildMember(sweepAccount, sweepGuild) {
		t.Fatal("swept guild retained membership")
	}
	var pendingRequests int
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM guild_applications WHERE guild_id=$1 AND resolved_at IS NULL)+
		(SELECT count(*) FROM guild_invitations WHERE guild_id=$1 AND resolved_at IS NULL)`, sweepGuild).Scan(&pendingRequests); err != nil || pendingRequests != 0 {
		t.Fatalf("orphaned sweep requests=%d err=%v", pendingRequests, err)
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
	eventID, err := newGuildID(now.Add(3 * time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	event := save.EventRecord{EventID: eventID, StreamID: "018f0000-0000-7000-8000-000000000301", OwnerID: founderID,
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
	}, 100); err == nil {
		t.Fatal("clearing accepted a partial active-member snapshot")
	}
	if err := service.CommitClearingBoundary(ctx, guildID, 1, now.Add(time.Minute), []MemberStock{
		{AccountID: accounts[0], Produces: "libraries", Consumes: "carbon", AvailableUnits: 10},
		{AccountID: joinedAccount, Produces: "carbon", Consumes: "libraries", AvailableUnits: 10},
	}, 100); err != nil {
		t.Fatal(err)
	}
	committedSnapshot := []MemberStock{
		{AccountID: accounts[0], Produces: "libraries", Consumes: "carbon", AvailableUnits: 10},
		{AccountID: joinedAccount, Produces: "carbon", Consumes: "libraries", AvailableUnits: 10},
	}
	if err := service.CommitClearingBoundary(ctx, guildID, 1, now.Add(time.Minute), committedSnapshot, 100); err != nil {
		t.Fatalf("identical clearing retry: %v", err)
	}
	committedSnapshot[0].AvailableUnits++
	if err := service.CommitClearingBoundary(ctx, guildID, 1, now.Add(time.Minute), committedSnapshot, 100); err == nil {
		t.Fatal("different snapshot reused a committed boundary sequence")
	}
	pending, err := service.PendingSettlements(ctx, founderID, "", 0)
	if err != nil || len(pending.Settlements) != 1 || pending.Settlements[0].DebitUnits != 5 || pending.Settlements[0].CreditUnits != 5 {
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

	// Regression for F1: two already-closed histories plus one active history
	// must all anonymize when the account is deleted.
	deletionAccount := "018f0000-0000-4000-8000-000000000105"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',$2)`, deletionAccount, now); err != nil {
		t.Fatal(err)
	}
	deletionGuilds := []string{
		"018f0000-0000-7000-8000-000000000501",
		"018f0000-0000-7000-8000-000000000502",
		"018f0000-0000-7000-8000-000000000503",
	}
	for index, deletionGuild := range deletionGuilds {
		if _, err := db.ExecContext(ctx, `INSERT INTO guilds(guild_id,name,created_at,founder_account,join_policy,revision) VALUES($1,$2,$3,$4,'open',1)`, deletionGuild, fmt.Sprintf("delete fixture %d", index), now, deletionAccount); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO guild_members(guild_id,account_id,joined_at,role) VALUES($1,$2,$3,'leader')`, deletionGuild, deletionAccount, now.Add(time.Duration(index)*time.Second)); err != nil {
			t.Fatal(err)
		}
		if index < 2 {
			if _, err := db.ExecContext(ctx, `UPDATE guild_members SET left_at=$2 WHERE guild_id=$1 AND account_id=$3 AND left_at IS NULL`, deletionGuild, now.Add(time.Duration(index+1)*time.Minute), deletionAccount); err != nil {
				t.Fatal(err)
			}
		}
	}
	deleteTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.PrepareAccountDeletion(ctx, deleteTx, deletionAccount, now.Add(10*time.Minute)); err != nil {
		deleteTx.Rollback()
		t.Fatal(err)
	}
	if _, err := deleteTx.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, deletionAccount); err != nil {
		deleteTx.Rollback()
		t.Fatal(err)
	}
	if err := deleteTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=ANY($1::uuid[]) AND account_id IS NULL AND left_at IS NOT NULL`, deletionGuilds).Scan(&anonymized); err != nil || anonymized != 3 {
		t.Fatalf("closed-history anonymization=%d err=%v", anonymized, err)
	}
	var presenceAccountID *string
	var presenceRef string
	if err := db.QueryRowContext(ctx, `SELECT account_id,account_ref FROM guild_presence_outbox WHERE guild_id=$1 AND kind='left' ORDER BY outbox_id DESC LIMIT 1`, deletionGuilds[2]).Scan(&presenceAccountID, &presenceRef); err != nil || presenceAccountID != nil || presenceRef != deletionAccount {
		t.Fatalf("deletion presence account=%v ref=%q err=%v", presenceAccountID, presenceRef, err)
	}
}

func TestGuildActivityEvaluationDenominatorIntegration(t *testing.T) {
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
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	guildID := "018f0000-0000-7000-8000-000000000701"
	accounts := []string{
		"018f0000-0000-4000-8000-000000000601",
		"018f0000-0000-4000-8000-000000000602",
		"018f0000-0000-4000-8000-000000000603",
	}
	founders := []string{
		"018f0000-0000-4000-8000-000000000611",
		"018f0000-0000-4000-8000-000000000612",
		"018f0000-0000-4000-8000-000000000613",
	}
	for index, accountID := range accounts {
		if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',$2)`, accountID, now); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,created_at) VALUES($1,$2,$3)`, accountID, founders[index], now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO guilds(guild_id,name,created_at,founder_account,join_policy,revision) VALUES($1,'activity fixture',$2,$3,'open',1)`, guildID, now, accounts[0]); err != nil {
		t.Fatal(err)
	}
	for index, accountID := range accounts {
		role := "member"
		if index == 0 {
			role = "leader"
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO guild_members(guild_id,account_id,joined_at,role) VALUES($1,$2,$3,$4)`, guildID, accountID, now, role); err != nil {
			t.Fatal(err)
		}
	}
	projector, err := NewProjector(db, integrationWindow{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM guild_projection_events WHERE event_id=ANY($1::uuid[])`, []string{
		"018f0000-0000-7000-8000-000000000631",
		"018f0000-0000-7000-8000-000000000632",
		"018f0000-0000-7000-8000-000000000633",
	}); err != nil {
		t.Fatal(err)
	}
	events := make([]save.EventRecord, 0, 3)
	for index := range accounts {
		kind := save.EventGuildActivityEvaluated
		xp := int64(0)
		progress := int64(0)
		if index == 0 {
			kind = save.EventGuildTitheAccrued
			xp = 10
			progress = 500_000
		}
		streamID := fmt.Sprintf("018f0000-0000-7000-8000-00000000062%d", index+1)
		events = append(events, save.EventRecord{
			EventID: fmt.Sprintf("018f0000-0000-7000-8000-00000000063%d", index+1), StreamID: streamID,
			OwnerID: founders[index], Revision: 2, Kind: kind,
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", OccurredAt: now,
			Payload: json.RawMessage(fmt.Sprintf(`{"founder_id":"%s","run_id":{"company_stream_id":"%s","run_seq":1},"progress_delta_ppm":%d,"xp_delta":%d}`, founders[index], streamID, progress, xp)),
		})
	}
	if err := projector.Project(ctx, events); err != nil {
		t.Fatal(err)
	}
	var activeFounders, tithedXP int64
	if err := db.QueryRowContext(ctx, `SELECT active_founders,tithed_xp FROM guild_health_inputs WHERE guild_id=$1 AND window_start=$2`, guildID, now).Scan(&activeFounders, &tithedXP); err != nil {
		t.Fatal(err)
	}
	if activeFounders != 3 || tithedXP != 10 {
		t.Fatalf("health input active=%d xp=%d", activeFounders, tithedXP)
	}
}

func TestGuildIntentRejectionsAndLeadershipConcurrencyIntegration(t *testing.T) {
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
	accounts := make([]string, 9)
	for index := range accounts {
		accounts[index] = fmt.Sprintf("018f0000-0000-4000-8000-0000000008%02d", index+1)
		if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',clock_timestamp())`, accounts[index]); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := LoadCatalog([]byte(phase0Catalog))
	if err != nil {
		t.Fatal(err)
	}
	catalog.MaxMembers = 3
	now := time.Date(2026, 7, 31, 15, 0, 0, 0, time.UTC)
	service, err := NewService(db, catalog, integrationNames{}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	assertCategory := func(accountID, intentID, kind string, revision int, fields, expected string) {
		t.Helper()
		result, handleErr := service.Handle(ctx, accountID, []byte(guildIntent(intentID, kind, revision, fields)))
		if handleErr != nil {
			t.Fatalf("%s: %v", kind, handleErr)
		}
		if category := receiptCategory(t, result.Receipt); category != expected {
			t.Fatalf("%s category=%q want=%q receipt=%s", kind, category, expected, result.Receipt)
		}
	}
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000801", "create_guild", 1, `,"name":"no","join_policy":"open"`, "name_policy")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000802", "join_guild", 1, `,"guild_id":"018f0000-0000-7000-8000-000000009999"`, "unknown_id")

	create := func(accountID, intentID, name, policy string) string {
		t.Helper()
		result, createErr := service.Handle(ctx, accountID, []byte(guildIntent(intentID, "create_guild", 1, `,"name":"`+name+`","join_policy":"`+policy+`"`)))
		if createErr != nil || receiptOutcome(t, result.Receipt) != "applied" {
			t.Fatalf("create %s=%s err=%v", policy, result.Receipt, createErr)
		}
		return receiptGuildID(t, result.Receipt)
	}
	openGuild := create(accounts[0], "018f0000-0000-7000-8000-000000000803", "small systems ropen", "open")
	applyGuild := create(accounts[1], "018f0000-0000-7000-8000-000000000804", "small systems rapply", "apply")
	inviteGuild := create(accounts[2], "018f0000-0000-7000-8000-000000000805", "small systems rinvite", "invite")

	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000806", "join_guild", 1, `,"guild_id":"`+applyGuild+`"`, "not_open")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000807", "apply_guild", 1, `,"guild_id":"`+openGuild+`"`, "not_apply")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000808", "admit_member", 1, `,"account_id":"`+accounts[4]+`"`, "not_officer")
	assertCategory(accounts[1], "018f0000-0000-7000-8000-000000000809", "admit_member", 2, `,"account_id":"`+accounts[4]+`"`, "not_applicant")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000810", "accept_invite", 1, `,"guild_id":"`+inviteGuild+`"`, "not_invited")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000811", "leave_guild", 1, "", "not_member")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000812", "set_role", 1, `,"account_id":"`+accounts[4]+`","role":"officer"`, "not_leader")
	assertCategory(accounts[3], "018f0000-0000-7000-8000-000000000813", "disband_guild", 1, "", "not_leader")
	assertCategory(accounts[0], "018f0000-0000-7000-8000-000000000814", "leave_guild", 2, "", "leader_required")
	assertCategory(accounts[0], "018f0000-0000-7000-8000-000000000815", "set_role", 2, `,"account_id":"`+accounts[0]+`","role":"officer"`, "leadership_transfer_required")

	for index, accountID := range accounts[4:6] {
		result, joinErr := service.Handle(ctx, accountID, []byte(guildIntent(fmt.Sprintf("018f0000-0000-7000-8000-00000000082%d", index), "join_guild", 1, `,"guild_id":"`+openGuild+`"`)))
		if joinErr != nil || receiptOutcome(t, result.Receipt) != "applied" {
			t.Fatalf("join target %d=%s err=%v", index, result.Receipt, joinErr)
		}
	}
	assertCategory(accounts[0], "018f0000-0000-7000-8000-000000000816", "disband_guild", 2, "", "guild_not_empty")
	assertCategory(accounts[2], "018f0000-0000-7000-8000-000000000817", "invite_member", 2, `,"account_id":"`+accounts[4]+`"`, "already_member")

	// Two simultaneous transfers serialize on the leader's account revision.
	// Exactly one intent applies, and the partial unique leader invariant holds.
	var wait sync.WaitGroup
	transferOutcomes := make(chan string, 2)
	for index, target := range accounts[4:6] {
		wait.Add(1)
		go func(index int, target string) {
			defer wait.Done()
			result, transferErr := service.Handle(ctx, accounts[0], []byte(guildIntent(fmt.Sprintf("018f0000-0000-7000-8000-00000000083%d", index), "set_role", 2, `,"account_id":"`+target+`","role":"leader"`)))
			if transferErr != nil {
				transferOutcomes <- "error"
				return
			}
			if receiptOutcome(t, result.Receipt) == "rejected" {
				transferOutcomes <- receiptCategory(t, result.Receipt)
				return
			}
			transferOutcomes <- "applied"
		}(index, target)
	}
	wait.Wait()
	close(transferOutcomes)
	transferCounts := map[string]int{}
	for outcome := range transferOutcomes {
		transferCounts[outcome]++
	}
	if transferCounts["applied"] != 1 || transferCounts["revision_conflict"] != 1 {
		t.Fatalf("transfer outcomes=%v", transferCounts)
	}
	var leaders int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL AND role='leader'`, openGuild).Scan(&leaders); err != nil || leaders != 1 {
		t.Fatalf("leaders=%d err=%v", leaders, err)
	}

	// Cross-guild targets are rejected before their row locks, so opposite
	// requests have no target-guild AB-BA edge.
	deadlineCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	crossResults := make(chan error, 2)
	for index, pair := range [][2]string{{accounts[1], accounts[2]}, {accounts[2], accounts[1]}} {
		go func(index int, actor, target string) {
			result, callErr := service.Handle(deadlineCtx, actor, []byte(guildIntent(fmt.Sprintf("018f0000-0000-7000-8000-00000000084%d", index), "set_role", 2, `,"account_id":"`+target+`","role":"officer"`)))
			if callErr == nil && receiptCategory(t, result.Receipt) != "not_member" {
				callErr = fmt.Errorf("unexpected receipt %s", result.Receipt)
			}
			crossResults <- callErr
		}(index, pair[0], pair[1])
	}
	for range 2 {
		if callErr := <-crossResults; callErr != nil {
			t.Fatal(callErr)
		}
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

func receiptCategory(t *testing.T, data json.RawMessage) string {
	t.Helper()
	var receipt struct {
		Rejection struct {
			Category string `json:"category"`
		} `json:"rejection"`
	}
	if err := json.Unmarshal(data, &receipt); err != nil || receipt.Rejection.Category == "" {
		t.Fatalf("receipt rejection: data=%s err=%v", data, err)
	}
	return receipt.Rejection.Category
}
