package transport

import "strings"

type Identity struct {
	AccountID string
	FounderID string
}

type Memberships interface {
	GuildMember(accountID, guildID string) bool
	CohortMember(founderID, cohortID string) bool
	MatchParticipant(founderID, matchID string) bool
}

func Authorized(identity Identity, channel string, memberships Memberships) bool {
	if identity.AccountID == "" || identity.FounderID == "" || channel == "" {
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
		return memberships.GuildMember(identity.AccountID, id)
	case "cohort":
		return memberships.CohortMember(identity.FounderID, id)
	case "match":
		return memberships.MatchParticipant(identity.FounderID, id)
	default:
		return false
	}
}

func isPlayerChannel(channel string) bool { return strings.HasPrefix(channel, "player:") }
