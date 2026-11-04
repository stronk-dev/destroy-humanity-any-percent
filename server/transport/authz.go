package transport

import (
	"context"
	"strings"
)

type Identity struct {
	AccountID string
	FounderID string
}

type Memberships interface {
	GuildMember(ctx context.Context, accountID, guildID string) bool
	CohortMember(ctx context.Context, founderID, cohortID string) bool
	MatchParticipant(ctx context.Context, founderID, matchID string) bool
}

func Authorized(ctx context.Context, identity Identity, channel string, memberships Memberships) bool {
	if ctx.Err() != nil || identity.AccountID == "" || identity.FounderID == "" || channel == "" {
		return false
	}
	if channel == "world" || channel == "feed" || channel == "player:"+identity.FounderID {
		return true
	}
	prefix, id, ok := strings.Cut(channel, ":")
	if !ok || id == "" || strings.Contains(id, ":") || memberships == nil {
		return false
	}
	switch prefix {
	case "guild":
		return memberships.GuildMember(ctx, identity.AccountID, id)
	case "cohort":
		return memberships.CohortMember(ctx, identity.FounderID, id)
	case "match":
		return memberships.MatchParticipant(ctx, identity.FounderID, id)
	default:
		return false
	}
}

func isPlayerChannel(channel string) bool { return strings.HasPrefix(channel, "player:") }
