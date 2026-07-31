// Package replaycatalog composes the six immutable Phase-0 balance artifacts
// without weakening the production/Commons amplitude boundary.
package replaycatalog

import (
	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
)

func Load(constantsHash string, artifacts map[string][]byte) (production.CatalogBundle, error) {
	if constantsHash == "" || len(artifacts) != 6 {
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
	return production.CatalogBundle{ConstantsHash: constantsHash, Economy: economyCatalog, Routes: routeCatalog,
		Commons: commonsbinding.ReplayPolicy{Catalog: commonsCatalog}, Prestige: prestigePolicy,
		Faction: factionCatalog, Guild: guildCatalog}, nil
}
