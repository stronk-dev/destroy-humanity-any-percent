// Package replaycatalog composes the six immutable Phase-0 balance artifacts
// without weakening the production/Commons amplitude boundary.
package replaycatalog

import (
	"context"
	"database/sql"
	"sort"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

func LoadDatabase(ctx context.Context, db *sql.DB) (production.ReplayCatalogSet, error) {
	if db == nil {
		return nil, production.ErrInvalidReplayInputs
	}
	rows, err := db.QueryContext(ctx, `SELECT constants_hash,artifact_name,bytes FROM catalog_artifacts ORDER BY constants_hash,artifact_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	artifacts := map[string]map[string][]byte{}
	for rows.Next() {
		var hash, name string
		var data []byte
		if err := rows.Scan(&hash, &name, &data); err != nil {
			return nil, err
		}
		if artifacts[hash] == nil {
			artifacts[hash] = map[string][]byte{}
		}
		artifacts[hash][name] = append([]byte(nil), data...)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	set := make(production.ReplayCatalogSet, len(artifacts))
	for hash, values := range artifacts {
		bundle, err := Load(hash, values)
		if err != nil {
			return nil, err
		}
		set[hash] = bundle
	}
	if len(set) == 0 {
		return nil, production.ErrInvalidReplayInputs
	}
	return set, nil
}

func Load(constantsHash string, artifacts map[string][]byte) (production.CatalogBundle, error) {
	if constantsHash == "" || !validArtifactNames(artifacts) {
		return production.CatalogBundle{}, production.ErrInvalidReplayInputs
	}
	computed, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || computed != constantsHash {
		return production.CatalogBundle{}, production.ErrInvalidReplayInputs
	}
	economyCatalog, err := economy.LoadCatalog(artifacts["economy"])
	if err != nil {
		return production.CatalogBundle{}, err
	}
	routeCatalog, err := routes.LoadCatalog(artifacts["routes"])
	if err != nil {
		return production.CatalogBundle{}, err
	}
	commonsCatalog, err := commons.LoadCatalog(artifacts["commons"])
	if err != nil {
		return production.CatalogBundle{}, err
	}
	prestigePolicy, err := prestigecore.LoadPolicy(artifacts["prestige"])
	if err != nil {
		return production.CatalogBundle{}, err
	}
	factionCatalog, err := faction.LoadCatalog(artifacts["factions"], faction.CompactTitheBand{
		MinimumPPM: commonsCatalog.MinimumTithePPM, DefaultPPM: commonsCatalog.DefaultTithePPM, MaximumPPM: commonsCatalog.MaximumTithePPM,
	})
	if err != nil {
		return production.CatalogBundle{}, err
	}
	guildCatalog, err := guild.LoadCatalog(artifacts["guilds"])
	if err != nil {
		return production.CatalogBundle{}, err
	}
	gates := routeCatalog.Gates()
	gateIDs := make([]string, len(gates))
	for index, gate := range gates {
		gateIDs[index] = gate.ID
	}
	sort.Strings(gateIDs)
	if err := economyCatalog.ValidateGateReferences(gateIDs); err != nil {
		return production.CatalogBundle{}, err
	}
	if _, err := leaderboard.LoadCategoryCatalog(artifacts["categories"], gateIDs); err != nil {
		return production.CatalogBundle{}, err
	}
	frozen := make(map[string][]byte, len(artifacts))
	for name, data := range artifacts {
		frozen[name] = append([]byte(nil), data...)
	}
	bundle := production.CatalogBundle{ConstantsHash: constantsHash, Artifacts: frozen, Economy: economyCatalog, Routes: routeCatalog,
		Commons: commonsbinding.ReplayPolicy{Catalog: commonsCatalog}, Prestige: prestigePolicy,
		Faction: factionCatalog, Guild: guildCatalog}
	if _, active := artifacts["meters"]; active {
		meterCatalog, meterErr := meters.LoadCatalog(artifacts["meters"])
		if meterErr != nil {
			return production.CatalogBundle{}, meterErr
		}
		achievementCatalog, achievementErr := achievements.LoadCatalog(artifacts["achievements"], production.FoundationAchievementRegistry(economyCatalog))
		if achievementErr != nil {
			return production.CatalogBundle{}, achievementErr
		}
		bundle.Meters, bundle.Achievements = meterCatalog, achievementCatalog
	}
	if doctrineBytes, active := artifacts["doctrines"]; active {
		doctrineCatalog, doctrineErr := doctrine.LoadCatalog(doctrineBytes)
		if doctrineErr != nil {
			return production.CatalogBundle{}, doctrineErr
		}
		if doctrineErr := doctrineCatalog.ValidateRoutes(routeCatalog); doctrineErr != nil {
			return production.CatalogBundle{}, doctrineErr
		}
		bundle.Doctrines = doctrineCatalog
	}
	if minigameBytes, active := artifacts["minigames"]; active {
		minigameCatalog, minigameErr := minigame.LoadCatalog(minigameBytes)
		if minigameErr != nil {
			return production.CatalogBundle{}, minigameErr
		}
		bundle.Minigames = minigameCatalog
	}
	if petBytes, active := artifacts["pets"]; active {
		petCatalog, petErr := pet.LoadCatalog(petBytes)
		if petErr != nil {
			return production.CatalogBundle{}, petErr
		}
		bundle.Pets = petCatalog
	}
	if fiscalBytes, active := artifacts["fiscal"]; active {
		fiscalCatalog, fiscalErr := fiscal.LoadCatalog(fiscalBytes, economyCatalog)
		if fiscalErr != nil {
			return production.CatalogBundle{}, fiscalErr
		}
		bundle.Fiscal = fiscalCatalog
	}
	if soulBytes, active := artifacts["soul"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		soulCatalog, soulErr := soul.LoadCatalog(soulBytes, soul.Declarations{CopyKeys: keys, EpochSeeded: true,
			CatchupCeilingMS: prestigePolicy.CatchupCeilingMS})
		if soulErr != nil {
			return production.CatalogBundle{}, soulErr
		}
		bundle.Soul = soulCatalog
	}
	if pitchBytes, active := artifacts["pitch"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		pitchCatalog, pitchErr := pitch.LoadCatalog(pitchBytes, pitch.Declarations{CopyKeys: keys})
		if pitchErr != nil {
			return production.CatalogBundle{}, pitchErr
		}
		bundle.Pitch = pitchCatalog
	}
	if apiBytes, active := artifacts["minigame_api"]; active {
		apiCatalog, apiErr := minigameapi.LoadCatalog(apiBytes)
		if apiErr != nil || bundle.Minigames == nil || bundle.Pitch == nil {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
		definition, ok := bundle.Minigames.Definition("pitch")
		if !ok || !apiCatalog.SupportsTenant(definition.MinigameID, definition.EngineRef, definition.EngineVersion) {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
		bundle.MinigameAPI = apiCatalog
	}
	return bundle, nil
}

func validArtifactNames(artifacts map[string][]byte) bool {
	base := [...]string{"categories", "commons", "economy", "factions", "guilds", "prestige", "routes"}
	allowed := make(map[string]bool, len(base)+9)
	for _, name := range base {
		allowed[name] = true
		if len(artifacts[name]) == 0 {
			return false
		}
	}
	for _, name := range [...]string{"achievements", "doctrines", "fiscal", "meters", "minigame_api", "minigames", "pets", "pitch", "soul"} {
		allowed[name] = true
	}
	for name, data := range artifacts {
		if !allowed[name] || len(data) == 0 {
			return false
		}
	}
	_, meters := artifacts["meters"]
	_, achievements := artifacts["achievements"]
	_, doctrines := artifacts["doctrines"]
	_, minigames := artifacts["minigames"]
	_, pets := artifacts["pets"]
	_, fiscalActive := artifacts["fiscal"]
	_, soulActive := artifacts["soul"]
	_, pitchActive := artifacts["pitch"]
	_, minigameAPIActive := artifacts["minigame_api"]
	if meters != achievements || doctrines && !meters || minigames && !meters || pets && !minigames || fiscalActive && !pets ||
		soulActive && !fiscalActive || pitchActive && !soulActive || minigameAPIActive && !pitchActive {
		return false
	}
	want := len(base)
	if meters {
		want += 2
	}
	if doctrines {
		want++
	}
	if minigames {
		want++
	}
	if pets {
		want++
	}
	if fiscalActive {
		want++
	}
	if soulActive {
		want++
	}
	if pitchActive {
		want++
	}
	if minigameAPIActive {
		want++
	}
	return len(artifacts) == want
}
