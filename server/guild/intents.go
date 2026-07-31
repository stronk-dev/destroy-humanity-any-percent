package guild

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidIntent = errors.New("invalid guild intent")
	uuidPattern      = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuidV7Pattern    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

const maxSafeInteger = int64(9_007_199_254_740_991)

type NameValidator interface {
	ValidateGuildName(string) bool
}

type Service struct {
	db      *sql.DB
	catalog *Catalog
	names   NameValidator
	clock   func() time.Time
}

type HandleResult struct {
	Receipt json.RawMessage
	Replay  bool
}

type IntentRequest struct {
	IntentID         string
	Kind             string
	ExpectedRevision int64
	GuildID          string
	AccountID        string
	Name             string
	JoinPolicy       string
	Role             string
	RequestHash      string
}

func NewService(db *sql.DB, catalog *Catalog, names NameValidator, clock func() time.Time) (*Service, error) {
	if db == nil || catalog == nil || names == nil {
		return nil, ErrInvalidIntent
	}
	if clock == nil {
		clock = time.Now
	}
	return &Service{db: db, catalog: catalog, names: names, clock: clock}, nil
}

func ParseIntent(data []byte) (IntentRequest, error) {
	var root map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&root); err != nil || root == nil {
		return IntentRequest{}, ErrInvalidIntent
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return IntentRequest{}, ErrInvalidIntent
	}
	var request IntentRequest
	if json.Unmarshal(root["intent_id"], &request.IntentID) != nil || !uuidV7Pattern.MatchString(request.IntentID) ||
		json.Unmarshal(root["kind"], &request.Kind) != nil || !positiveInt(root["expected_revision"], &request.ExpectedRevision) {
		return IntentRequest{}, ErrInvalidIntent
	}
	keys := []string{"intent_id", "kind", "expected_revision"}
	switch request.Kind {
	case "create_guild":
		keys = append(keys, "name", "join_policy")
		_ = json.Unmarshal(root["name"], &request.Name)
		_ = json.Unmarshal(root["join_policy"], &request.JoinPolicy)
	case "join_guild", "apply_guild", "accept_invite":
		keys = append(keys, "guild_id")
		_ = json.Unmarshal(root["guild_id"], &request.GuildID)
	case "admit_member", "invite_member":
		keys = append(keys, "account_id")
		_ = json.Unmarshal(root["account_id"], &request.AccountID)
	case "set_role":
		keys = append(keys, "account_id", "role")
		_ = json.Unmarshal(root["account_id"], &request.AccountID)
		_ = json.Unmarshal(root["role"], &request.Role)
	case "leave_guild", "disband_guild":
	default:
		return IntentRequest{}, ErrInvalidIntent
	}
	if !exactKeys(root, keys...) || request.GuildID != "" && !uuidV7Pattern.MatchString(request.GuildID) ||
		request.AccountID != "" && !uuidPattern.MatchString(request.AccountID) ||
		request.JoinPolicy != "" && request.JoinPolicy != "open" && request.JoinPolicy != "invite" && request.JoinPolicy != "apply" ||
		request.Role != "" && request.Role != "leader" && request.Role != "officer" && request.Role != "member" {
		return IntentRequest{}, ErrInvalidIntent
	}
	canonical := make(map[string]any, len(root)-1)
	for key, raw := range root {
		if key == "intent_id" {
			continue
		}
		var value any
		valueDecoder := json.NewDecoder(bytes.NewReader(raw))
		valueDecoder.UseNumber()
		if valueDecoder.Decode(&value) != nil {
			return IntentRequest{}, ErrInvalidIntent
		}
		canonical[key] = value
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return IntentRequest{}, ErrInvalidIntent
	}
	digest := sha256.Sum256(encoded)
	request.RequestHash = "sha256:" + hex.EncodeToString(digest[:])
	return request, nil
}

func positiveInt(raw json.RawMessage, destination *int64) bool {
	return len(raw) > 0 && json.Unmarshal(raw, destination) == nil && *destination > 0 && *destination <= maxSafeInteger
}

func exactKeys(root map[string]json.RawMessage, keys ...string) bool {
	if len(root) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := root[key]; !ok {
			return false
		}
	}
	return true
}

func (service *Service) Handle(ctx context.Context, actorAccount string, data []byte) (HandleResult, error) {
	if !uuidPattern.MatchString(actorAccount) {
		return HandleResult{}, ErrInvalidIntent
	}
	request, err := ParseIntent(data)
	if err != nil {
		return HandleResult{}, err
	}
	now := service.clock().UTC().Truncate(time.Millisecond)
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return HandleResult{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_account_revisions(account_id) VALUES($1) ON CONFLICT DO NOTHING`, actorAccount); err != nil {
		return HandleResult{}, err
	}
	var revision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM guild_account_revisions WHERE account_id=$1 FOR UPDATE`, actorAccount).Scan(&revision); err != nil {
		return HandleResult{}, err
	}
	var recordedHash, recordedOutcome string
	var recordedReceipt []byte
	err = tx.QueryRowContext(ctx, `SELECT request_hash,outcome,receipt FROM guild_intent_records WHERE account_id=$1 AND intent_id=$2`, actorAccount, request.IntentID).Scan(&recordedHash, &recordedOutcome, &recordedReceipt)
	if err == nil {
		if recordedHash != request.RequestHash {
			return HandleResult{}, ErrInvalidIntent
		}
		var normalized any
		decoder := json.NewDecoder(bytes.NewReader(recordedReceipt))
		decoder.UseNumber()
		if err := decoder.Decode(&normalized); err != nil {
			return HandleResult{}, err
		}
		recordedReceipt, err = json.Marshal(normalized)
		if err != nil {
			return HandleResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return HandleResult{}, err
		}
		return HandleResult{Receipt: recordedReceipt, Replay: true}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return HandleResult{}, err
	}
	if request.ExpectedRevision != revision {
		return service.finish(ctx, tx, actorAccount, request, revision, "rejected", "revision_conflict", "expected_revision", "", now)
	}
	if _, err := tx.ExecContext(ctx, `SAVEPOINT guild_mutation`); err != nil {
		return HandleResult{}, err
	}
	result, err := service.apply(ctx, tx, actorAccount, request, revision, now)
	if err != nil {
		if constraint, ok := uniqueConstraint(err); ok {
			if _, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT guild_mutation`); rollbackErr != nil {
				return HandleResult{}, errors.Join(err, rollbackErr)
			}
			category := "conflict"
			if constraint == "guilds_active_name_idx" {
				category = "name_taken"
			} else if constraint == "guild_members_one_active_account_idx" {
				category = "already_member"
			} else if constraint == "guild_members_one_active_leader_idx" {
				category = "leadership_conflict"
			}
			return service.finish(ctx, tx, actorAccount, request, revision, "rejected", category, constraint, "", now)
		}
		return HandleResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT guild_mutation`); err != nil {
		return HandleResult{}, err
	}
	return service.finish(ctx, tx, actorAccount, request, revision, result.outcome, result.category, result.detail, result.guildID, now)
}

func (service *Service) HandleGuild(ctx context.Context, actorAccount string, data []byte) (json.RawMessage, bool, error) {
	result, err := service.Handle(ctx, actorAccount, data)
	return result.Receipt, result.Replay, err
}

func (service *Service) IsInvalidGuildIntent(err error) bool { return errors.Is(err, ErrInvalidIntent) }

type mutationResult struct{ outcome, category, detail, guildID string }

func applied(guildID string) mutationResult {
	return mutationResult{outcome: "applied", guildID: guildID}
}
func rejected(category, detail string) mutationResult {
	return mutationResult{outcome: "rejected", category: category, detail: detail}
}

func (service *Service) apply(ctx context.Context, tx *sql.Tx, actor string, request IntentRequest, accountRevision int64, now time.Time) (mutationResult, error) {
	switch request.Kind {
	case "create_guild":
		normalizedName, ok := NormalizeGuildName(request.Name)
		if !ok || !service.names.ValidateGuildName(normalizedName) {
			return rejected("name_policy", "name"), nil
		}
		request.Name = normalizedName
		if active, _, _, err := activeMembership(ctx, tx, actor, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		guildID, err := newGuildID(now)
		if err != nil {
			return mutationResult{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO guilds(guild_id,name,created_at,founder_account,join_policy,revision) VALUES($1,$2,$3,$4,$5,1)`, guildID, request.Name, now, actor, request.JoinPolicy)
		if err != nil {
			return mutationResult{}, err
		}
		if err := service.join(ctx, tx, guildID, actor, "leader", 1, request.IntentID, now); err != nil {
			return mutationResult{}, err
		}
		if err := insertGuildEvent(ctx, tx, guildID, 1, "guild_created", actor, actor, request.IntentID, map[string]any{"name": request.Name, "join_policy": request.JoinPolicy}); err != nil {
			return mutationResult{}, err
		}
		return applied(guildID), nil
	case "join_guild":
		guildRevision, policy, err := lockGuild(ctx, tx, request.GuildID)
		if err != nil {
			return mapGuildLookup(err), nil
		}
		if policy != "open" {
			return rejected("not_open", request.GuildID), nil
		}
		if active, _, _, err := activeMembership(ctx, tx, actor, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		if full, err := guildFull(ctx, tx, request.GuildID, service.catalog.MaxMembers); err != nil {
			return mutationResult{}, err
		} else if full {
			return rejected("guild_full", request.GuildID), nil
		}
		guildRevision++
		if err := setGuildRevision(ctx, tx, request.GuildID, guildRevision); err != nil {
			return mutationResult{}, err
		}
		if err := service.join(ctx, tx, request.GuildID, actor, "member", guildRevision, request.IntentID, now); err != nil {
			return mutationResult{}, err
		}
		return applied(request.GuildID), nil
	case "apply_guild":
		_, policy, err := lockGuild(ctx, tx, request.GuildID)
		if err != nil {
			return mapGuildLookup(err), nil
		}
		if policy != "apply" {
			return rejected("not_apply", request.GuildID), nil
		}
		if active, _, _, err := activeMembership(ctx, tx, actor, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO guild_applications(guild_id,account_id,created_at) VALUES($1,$2,$3) ON CONFLICT (guild_id,account_id) WHERE resolved_at IS NULL DO NOTHING`, request.GuildID, actor, now)
		return applied(request.GuildID), err
	case "admit_member":
		guildID, _, _, err := activeMembership(ctx, tx, actor, false)
		if err != nil {
			return mutationResult{}, err
		}
		if guildID == "" {
			return rejected("not_officer", request.AccountID), nil
		}
		guildRevision, _, err := lockGuild(ctx, tx, guildID)
		if err != nil {
			return mutationResult{}, err
		}
		lockedGuild, role, _, err := activeMembership(ctx, tx, actor, true)
		if err != nil {
			return mutationResult{}, err
		}
		if lockedGuild != guildID || role == "member" {
			return rejected("not_officer", request.AccountID), nil
		}
		var exists bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM guild_applications WHERE guild_id=$1 AND account_id=$2 AND resolved_at IS NULL)`, guildID, request.AccountID).Scan(&exists)
		if err != nil {
			return mutationResult{}, err
		}
		if !exists {
			return rejected("not_applicant", request.AccountID), nil
		}
		if active, _, _, err := activeMembership(ctx, tx, request.AccountID, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		if full, err := guildFull(ctx, tx, guildID, service.catalog.MaxMembers); err != nil {
			return mutationResult{}, err
		} else if full {
			return rejected("guild_full", guildID), nil
		}
		guildRevision++
		if err := setGuildRevision(ctx, tx, guildID, guildRevision); err != nil {
			return mutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE guild_applications SET resolved_at=$3,admitted=true WHERE guild_id=$1 AND account_id=$2 AND resolved_at IS NULL`, guildID, request.AccountID, now); err != nil {
			return mutationResult{}, err
		}
		if err := service.join(ctx, tx, guildID, request.AccountID, "member", guildRevision, request.IntentID, now); err != nil {
			return mutationResult{}, err
		}
		return applied(guildID), nil
	case "invite_member":
		guildID, _, _, err := activeMembership(ctx, tx, actor, false)
		if err != nil {
			return mutationResult{}, err
		}
		if guildID == "" {
			return rejected("not_officer", request.AccountID), nil
		}
		if _, _, err := lockGuild(ctx, tx, guildID); err != nil {
			return mutationResult{}, err
		}
		lockedGuild, role, _, err := activeMembership(ctx, tx, actor, true)
		if err != nil {
			return mutationResult{}, err
		}
		if lockedGuild != guildID || role == "member" {
			return rejected("not_officer", request.AccountID), nil
		}
		if active, _, _, err := activeMembership(ctx, tx, request.AccountID, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO guild_invitations(guild_id,account_id,created_at) VALUES($1,$2,$3) ON CONFLICT (guild_id,account_id) WHERE resolved_at IS NULL DO NOTHING`, guildID, request.AccountID, now)
		return applied(guildID), err
	case "accept_invite":
		guildRevision, _, err := lockGuild(ctx, tx, request.GuildID)
		if err != nil {
			return mapGuildLookup(err), nil
		}
		var exists bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM guild_invitations WHERE guild_id=$1 AND account_id=$2 AND resolved_at IS NULL)`, request.GuildID, actor).Scan(&exists)
		if err != nil {
			return mutationResult{}, err
		}
		if !exists {
			return rejected("not_invited", request.GuildID), nil
		}
		if active, _, _, err := activeMembership(ctx, tx, actor, false); err != nil {
			return mutationResult{}, err
		} else if active != "" {
			return rejected("already_member", active), nil
		}
		if full, err := guildFull(ctx, tx, request.GuildID, service.catalog.MaxMembers); err != nil {
			return mutationResult{}, err
		} else if full {
			return rejected("guild_full", request.GuildID), nil
		}
		guildRevision++
		if err := setGuildRevision(ctx, tx, request.GuildID, guildRevision); err != nil {
			return mutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE guild_invitations SET resolved_at=$3,accepted=true WHERE guild_id=$1 AND account_id=$2 AND resolved_at IS NULL`, request.GuildID, actor, now); err != nil {
			return mutationResult{}, err
		}
		if err := service.join(ctx, tx, request.GuildID, actor, "member", guildRevision, request.IntentID, now); err != nil {
			return mutationResult{}, err
		}
		return applied(request.GuildID), nil
	case "leave_guild":
		guildID, _, _, err := activeMembership(ctx, tx, actor, false)
		if err != nil {
			return mutationResult{}, err
		}
		if guildID == "" {
			return rejected("not_member", actor), nil
		}
		guildRevision, _, err := lockGuild(ctx, tx, guildID)
		if err != nil {
			return mutationResult{}, err
		}
		lockedGuild, role, membershipID, err := activeMembership(ctx, tx, actor, true)
		if err != nil {
			return mutationResult{}, err
		}
		if lockedGuild != guildID {
			return rejected("not_member", actor), nil
		}
		if role == "leader" {
			return rejected("leader_required", guildID), nil
		}
		guildRevision++
		if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET left_at=$2 WHERE membership_id=$1`, membershipID, now); err != nil {
			return mutationResult{}, err
		}
		if err := setGuildRevision(ctx, tx, guildID, guildRevision); err != nil {
			return mutationResult{}, err
		}
		if err := insertGuildEvent(ctx, tx, guildID, guildRevision, "member_left", actor, actor, request.IntentID, map[string]any{}); err != nil {
			return mutationResult{}, err
		}
		if err := insertPresence(ctx, tx, guildID, actor, "left", guildRevision, now); err != nil {
			return mutationResult{}, err
		}
		if err := service.refreshFloorState(ctx, tx, guildID, now); err != nil {
			return mutationResult{}, err
		}
		return applied(guildID), nil
	case "set_role":
		guildID, _, _, err := activeMembership(ctx, tx, actor, false)
		if err != nil {
			return mutationResult{}, err
		}
		if guildID == "" {
			return rejected("not_leader", request.AccountID), nil
		}
		guildRevision, _, err := lockGuild(ctx, tx, guildID)
		if err != nil {
			return mutationResult{}, err
		}
		lockedGuild, role, _, err := activeMembership(ctx, tx, actor, true)
		if err != nil {
			return mutationResult{}, err
		}
		if lockedGuild != guildID || role != "leader" {
			return rejected("not_leader", request.AccountID), nil
		}
		targetGuild, targetRole, _, err := activeMembership(ctx, tx, request.AccountID, true)
		if err != nil {
			return mutationResult{}, err
		}
		if targetGuild != guildID {
			return rejected("not_member", request.AccountID), nil
		}
		if request.AccountID == actor && request.Role != "leader" {
			return rejected("leadership_transfer_required", actor), nil
		}
		guildRevision++
		if request.Role == "leader" && request.AccountID != actor {
			if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET role='officer' WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL`, guildID, actor); err != nil {
				return mutationResult{}, err
			}
			if err := insertGuildEvent(ctx, tx, guildID, guildRevision, "role_changed", actor, actor, request.IntentID, map[string]any{"role": "officer"}); err != nil {
				return mutationResult{}, err
			}
		}
		if targetRole != request.Role {
			if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET role=$3 WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL`, guildID, request.AccountID, request.Role); err != nil {
				return mutationResult{}, err
			}
		}
		if err := setGuildRevision(ctx, tx, guildID, guildRevision); err != nil {
			return mutationResult{}, err
		}
		if err := insertGuildEvent(ctx, tx, guildID, guildRevision, "role_changed", actor, request.AccountID, request.IntentID, map[string]any{"role": request.Role}); err != nil {
			return mutationResult{}, err
		}
		return applied(guildID), nil
	case "disband_guild":
		guildID, _, _, err := activeMembership(ctx, tx, actor, false)
		if err != nil {
			return mutationResult{}, err
		}
		if guildID == "" {
			return rejected("not_leader", actor), nil
		}
		guildRevision, _, err := lockGuild(ctx, tx, guildID)
		if err != nil {
			return mutationResult{}, err
		}
		lockedGuild, role, membershipID, err := activeMembership(ctx, tx, actor, true)
		if err != nil {
			return mutationResult{}, err
		}
		if lockedGuild != guildID || role != "leader" {
			return rejected("not_leader", actor), nil
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL`, guildID).Scan(&count); err != nil {
			return mutationResult{}, err
		}
		if count != 1 {
			return rejected("guild_not_empty", guildID), nil
		}
		guildRevision++
		if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET left_at=$2 WHERE membership_id=$1`, membershipID, now); err != nil {
			return mutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE guilds SET revision=$2,disbanded_at=$3 WHERE guild_id=$1`, guildID, guildRevision, now); err != nil {
			return mutationResult{}, err
		}
		if err := insertGuildEvent(ctx, tx, guildID, guildRevision, "member_left", actor, actor, request.IntentID, map[string]any{}); err != nil {
			return mutationResult{}, err
		}
		if err := insertGuildEvent(ctx, tx, guildID, guildRevision, "guild_disbanded", actor, actor, request.IntentID, map[string]any{}); err != nil {
			return mutationResult{}, err
		}
		if err := insertPresence(ctx, tx, guildID, actor, "left", guildRevision, now); err != nil {
			return mutationResult{}, err
		}
		return applied(guildID), nil
	}
	return mutationResult{}, ErrInvalidIntent
}

func (service *Service) join(ctx context.Context, tx *sql.Tx, guildID, accountID, role string, revision int64, intentID string, now time.Time) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_members(guild_id,account_id,joined_at,role) VALUES($1,$2,$3,$4)`, guildID, accountID, now, role); err != nil {
		return err
	}
	if err := insertGuildEvent(ctx, tx, guildID, revision, "member_joined", accountID, accountID, intentID, map[string]any{"role": role}); err != nil {
		return err
	}
	if err := insertPresence(ctx, tx, guildID, accountID, "joined", revision, now); err != nil {
		return err
	}
	return service.refreshFloorState(ctx, tx, guildID, now)
}

func (service *Service) refreshFloorState(ctx context.Context, tx *sql.Tx, guildID string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE guilds SET below_floor_since=CASE
		WHEN (SELECT count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL) < $2
		THEN COALESCE(below_floor_since,$3) ELSE NULL END WHERE guild_id=$1`, guildID, service.catalog.MinMembers, now)
	return err
}

func (service *Service) finish(ctx context.Context, tx *sql.Tx, actor string, request IntentRequest, revision int64, outcome, category, detail, guildID string, now time.Time) (HandleResult, error) {
	newRevision := revision
	if outcome == "applied" {
		newRevision++
	}
	receipt := map[string]any{"intent_id": request.IntentID, "outcome": outcome}
	if outcome == "applied" {
		receipt["new_revision"] = newRevision
		receipt["guild_id"] = guildID
	} else {
		receipt["current_revision"] = revision
		receipt["rejection"] = map[string]string{"category": category, "detail": detail}
	}
	encoded, _ := json.Marshal(receipt)
	if outcome == "applied" {
		if _, err := tx.ExecContext(ctx, `UPDATE guild_account_revisions SET revision=$2 WHERE account_id=$1`, actor, newRevision); err != nil {
			return HandleResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO guild_intent_records(account_id,intent_id,request_hash,outcome,receipt,created_at) VALUES($1,$2,$3,$4,$5,$6)`, actor, request.IntentID, request.RequestHash, outcome, encoded, now); err != nil {
		return HandleResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: encoded}, nil
}

func activeMembership(ctx context.Context, tx *sql.Tx, accountID string, lock bool) (guildID, role, membershipID string, err error) {
	query := `SELECT guild_id,role,membership_id FROM guild_members WHERE account_id=$1 AND left_at IS NULL`
	if lock {
		query += ` FOR UPDATE`
	}
	err = tx.QueryRowContext(ctx, query, accountID).Scan(&guildID, &role, &membershipID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", nil
	}
	return
}

var errGuildMissing = errors.New("guild missing")

func lockGuild(ctx context.Context, tx *sql.Tx, guildID string) (int64, string, error) {
	var revision int64
	var policy string
	var disbanded sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT revision,join_policy,disbanded_at FROM guilds WHERE guild_id=$1 FOR UPDATE`, guildID).Scan(&revision, &policy, &disbanded)
	if errors.Is(err, sql.ErrNoRows) || disbanded.Valid {
		return 0, "", errGuildMissing
	}
	return revision, policy, err
}
func mapGuildLookup(err error) mutationResult {
	if errors.Is(err, errGuildMissing) {
		return rejected("unknown_id", "guild")
	}
	return rejected("internal_invariant", "guild")
}
func guildFull(ctx context.Context, tx *sql.Tx, guildID string, max int) (bool, error) {
	var count int
	err := tx.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL`, guildID).Scan(&count)
	return count >= max, err
}
func setGuildRevision(ctx context.Context, tx *sql.Tx, guildID string, revision int64) error {
	_, err := tx.ExecContext(ctx, `UPDATE guilds SET revision=$2 WHERE guild_id=$1`, guildID, revision)
	return err
}
func insertPresence(ctx context.Context, tx *sql.Tx, guildID, accountID, kind string, revision int64, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO guild_presence_outbox(guild_id,account_id,kind,guild_revision,occurred_at,active_count)
		SELECT $1,$2,$3,$4,$5,count(*) FROM guild_members WHERE guild_id=$1 AND left_at IS NULL`, guildID, accountID, kind, revision, now)
	return err
}
func insertGuildEvent(ctx context.Context, tx *sql.Tx, guildID string, revision int64, kind, actor, subject, intentID string, payload map[string]any) error {
	encoded, _ := json.Marshal(payload)
	_, err := tx.ExecContext(ctx, `INSERT INTO guild_events(guild_id,revision,kind,actor_account,subject_account,intent_id,payload) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		guildID, revision, kind, nullableUUID(actor), nullableUUID(subject), nullableUUID(intentID), encoded)
	return err
}

func nullableUUID(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func uniqueConstraint(err error) (string, bool) {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return pg.ConstraintName, true
	}
	return "", false
}

func (service *Service) GuildMember(accountID, guildID string) bool {
	if service == nil || !uuidPattern.MatchString(accountID) || !uuidV7Pattern.MatchString(guildID) {
		return false
	}
	var exists bool
	return service.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM guild_members WHERE account_id=$1 AND guild_id=$2 AND left_at IS NULL)`, accountID, guildID).Scan(&exists) == nil && exists
}

func (service *Service) FounderGuildMember(ctx context.Context, founderID string) (bool, error) {
	if service == nil || !uuidPattern.MatchString(founderID) {
		return false, ErrInvalidIntent
	}
	var exists bool
	err := service.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM account_founders f JOIN guild_members m ON m.account_id=f.account_id WHERE f.founder_id=$1 AND m.left_at IS NULL)`, founderID).Scan(&exists)
	return exists, err
}

func (service *Service) CohortMember(_, _ string) bool     { return false }
func (service *Service) MatchParticipant(_, _ string) bool { return false }
