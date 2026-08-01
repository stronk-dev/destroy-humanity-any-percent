package transport

import (
	"encoding/json"
)

const maxWireInteger = int64(9_007_199_254_740_991)

type WorldPlanet struct {
	DepletionPPM int64 `json:"depletion_ppm"`
	HealthPPM    int64 `json:"health_ppm"`
}

type WorldCommons struct {
	ServerHealthPPM int64 `json:"server_health_ppm"`
	ActiveFounders  int64 `json:"active_founders"`
	CompactMembers  int64 `json:"compact_members"`
}

type WorldPopulation struct {
	Online        int64 `json:"online"`
	FoundersTotal int64 `json:"founders_total"`
}

type WorldMilestones struct {
	ActiveID    *string `json:"active_id"`
	ProgressPPM int64   `json:"progress_ppm"`
}

type WorldEpoch struct {
	EpochID int64  `json:"epoch_id"`
	Name    string `json:"name"`
}

type WorldSnapshot struct {
	Version    int             `json:"v"`
	WorldRev   int64           `json:"world_rev"`
	Planet     WorldPlanet     `json:"planet"`
	Commons    WorldCommons    `json:"commons"`
	Population WorldPopulation `json:"population"`
	Milestones WorldMilestones `json:"milestones"`
	Epoch      WorldEpoch      `json:"epoch"`
}

func ValidateWorldSnapshot(snapshot WorldSnapshot) error {
	ppm := func(value int64) bool { return value >= 0 && value <= 1_000_000 }
	count := func(value int64) bool { return value >= 0 && value <= maxWireInteger }
	if snapshot.Version != 1 || snapshot.WorldRev < 1 || snapshot.WorldRev > maxWireInteger ||
		!ppm(snapshot.Planet.DepletionPPM) || !ppm(snapshot.Planet.HealthPPM) ||
		!ppm(snapshot.Commons.ServerHealthPPM) || !count(snapshot.Commons.ActiveFounders) || !count(snapshot.Commons.CompactMembers) ||
		!count(snapshot.Population.Online) || !count(snapshot.Population.FoundersTotal) ||
		!ppm(snapshot.Milestones.ProgressPPM) || snapshot.Epoch.EpochID < 1 || snapshot.Epoch.EpochID > maxWireInteger || snapshot.Epoch.Name == "" {
		return ErrInvalidPolicy
	}
	if snapshot.Milestones.ActiveID == nil {
		if snapshot.Milestones.ProgressPPM != 0 {
			return ErrInvalidPolicy
		}
	} else if !eventKindPattern.MatchString(*snapshot.Milestones.ActiveID) {
		return ErrInvalidPolicy
	}
	return nil
}

func decodeWorldSnapshot(data json.RawMessage) (WorldSnapshot, error) {
	var snapshot WorldSnapshot
	if decodeExactPayload(data, &snapshot) != nil || ValidateWorldSnapshot(snapshot) != nil {
		return WorldSnapshot{}, ErrInvalidPolicy
	}
	return snapshot, nil
}
