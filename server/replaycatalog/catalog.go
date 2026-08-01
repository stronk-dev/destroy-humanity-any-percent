// Package replaycatalog composes the six immutable Phase-0 balance artifacts
// without weakening the production/Commons amplitude boundary.
package replaycatalog

import (
	"context"
	"database/sql"
	"sort"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/leaderboard"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
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
	if constantsHash == "" || len(artifacts) != 7 {
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
	if _, err := leaderboard.LoadCategoryCatalog(artifacts["categories"], gateIDs); err != nil {
		return production.CatalogBundle{}, err
	}
	frozen := make(map[string][]byte, len(artifacts))
	for name, data := range artifacts {
		frozen[name] = append([]byte(nil), data...)
	}
	return production.CatalogBundle{ConstantsHash: constantsHash, Artifacts: frozen, Economy: economyCatalog, Routes: routeCatalog,
		Commons: commonsbinding.ReplayPolicy{Catalog: commonsCatalog}, Prestige: prestigePolicy,
		Faction: factionCatalog, Guild: guildCatalog}, nil
}
