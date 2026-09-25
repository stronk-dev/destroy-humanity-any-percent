// Package replaycatalog composes immutable, hash-pinned balance artifacts
// without weakening the production/Commons amplitude boundary.
package replaycatalog

import (
	"context"
	"database/sql"
	"sort"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/arcade"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/relevancepolicy"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
	"cloud-clicker/server/typer"
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
		resourceIDs := make([]string, 0, len(economyCatalog.Resources()))
		for _, resource := range economyCatalog.Resources() {
			resourceIDs = append(resourceIDs, resource.ID)
		}
		if meterErr := meterCatalog.ValidateResourceSeparation(resourceIDs); meterErr != nil {
			return production.CatalogBundle{}, meterErr
		}
		achievementCatalog, achievementErr := achievements.LoadCatalog(artifacts["achievements"], production.FoundationAchievementRegistry(economyCatalog))
		if achievementErr != nil {
			return production.CatalogBundle{}, achievementErr
		}
		bundle.Meters, bundle.Achievements = meterCatalog, achievementCatalog
	}
	var maximumAttainment, maximumScore int64
	if bundle.Achievements != nil {
		var scoreErr error
		if maximumAttainment, maximumScore, scoreErr = bundle.Achievements.MaximumScores(); scoreErr != nil {
			return production.CatalogBundle{}, scoreErr
		}
	}
	if err := economyCatalog.ValidateAxisInputs(bundle.Achievements != nil, maximumAttainment, maximumScore); err != nil {
		return production.CatalogBundle{}, err
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
	if typerBytes, active := artifacts["typer"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		typerCatalog, typerErr := typer.LoadCatalog(typerBytes, typer.Declarations{CopyKeys: keys})
		if typerErr != nil {
			return production.CatalogBundle{}, typerErr
		}
		bundle.Typer = typerCatalog
	}
	if arcadeBytes, active := artifacts["arcade"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		arcadeCatalog, arcadeErr := arcade.LoadCatalog(arcadeBytes, arcade.Declarations{CopyKeys: keys})
		if arcadeErr != nil {
			return production.CatalogBundle{}, arcadeErr
		}
		bundle.Arcade = arcadeCatalog
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
		// TT-PA3 loader chain: a typer definition row, the pinned typer artifact,
		// and the minigame_api typer tenant exist together or not at all.
		typerDefinition, hasTyperDefinition := bundle.Minigames.Definition(typer.EngineRef)
		typerTenant := apiCatalog.SupportsTenant(typer.EngineRef, typer.EngineRef, typer.EngineVersion)
		if hasTyperDefinition != (bundle.Typer != nil) || typerTenant != (bundle.Typer != nil) ||
			hasTyperDefinition && (typerDefinition.EngineRef != typer.EngineRef || typerDefinition.EngineVersion != typer.EngineVersion) {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
		if !validArcadeChain(bundle, apiCatalog) {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
		bundle.MinigameAPI = apiCatalog
	}
	// A typer definition row without its pinned artifact fails even without
	// minigame_api (the artifact itself already requires minigame_api).
	if bundle.Minigames != nil {
		if _, hasTyperDefinition := bundle.Minigames.Definition(typer.EngineRef); hasTyperDefinition != (bundle.Typer != nil) {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
		if !validArcadeChain(bundle, nil) {
			return production.CatalogBundle{}, minigameapi.ErrInvalidCatalog
		}
	}
	if opportunityBytes, active := artifacts["opportunities"]; active {
		opportunityCatalog, opportunityErr := activeplay.LoadCatalog(opportunityBytes, economyCatalog)
		if opportunityErr != nil {
			return production.CatalogBundle{}, opportunityErr
		}
		bundle.Opportunities = opportunityCatalog
	}
	if relevanceBytes, active := artifacts["relevance"]; active {
		// Epoch relevance artifacts may deliberately cover a gate-bounded catalog
		// slice (the cumulative T1 policy excludes later Legal Department content).
		// The strict parser still validates every present row and cross-reference;
		// scenario loaders own window completeness for the slice they measure.
		relevanceCatalog, relevanceErr := relevancepolicy.Load(relevanceBytes, economyCatalog, routeCatalog, false)
		if relevanceErr != nil {
			return production.CatalogBundle{}, relevanceErr
		}
		bundle.Relevance = relevanceCatalog
	}
	if curriculumBytes, active := artifacts["curriculum"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		gateIDs := make(map[string]struct{}, len(gates))
		for _, gate := range gates {
			gateIDs[gate.ID] = struct{}{}
		}
		curriculumCatalog, curriculumErr := curriculum.Load(curriculumBytes, curriculum.Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: gateIDs})
		if curriculumErr != nil {
			return production.CatalogBundle{}, curriculumErr
		}
		bundle.Curriculum = curriculumCatalog
	}
	if treeBytes, active := artifacts["reputation_tree"]; active {
		keys := make(map[string]struct{})
		for _, key := range copykeys.All() {
			keys[key] = struct{}{}
		}
		tree, treeErr := reputation.LoadTree(treeBytes, reputation.Declarations{Economy: economyCatalog, Curriculum: bundle.Curriculum, CopyKeys: keys})
		if treeErr != nil {
			return production.CatalogBundle{}, treeErr
		}
		bundle.ReputationTree = tree
	}
	if (bundle.ReputationTree != nil) != production.ReputationDeclared(economyCatalog) {
		return production.CatalogBundle{}, reputation.ErrInvalidTree
	}
	if speciesBytes, active := artifacts["pet_species"]; active {
		declarations := pet.SpeciesDeclarations{CopyKeys: map[string]struct{}{}, CompanionKeys: map[string]struct{}{}}
		for _, key := range copykeys.All() {
			declarations.CopyKeys[key] = struct{}{}
		}
		for _, key := range copykeys.CompanionKeys() {
			declarations.CompanionKeys[key] = struct{}{}
		}
		species, speciesErr := pet.LoadSpeciesCatalog(speciesBytes, declarations)
		if speciesErr != nil {
			return production.CatalogBundle{}, speciesErr
		}
		bundle.PetSpecies = species
	}
	if cosmeticsBytes, active := artifacts["cosmetics"]; active {
		cosmetics, cosmeticsErr := cosmetic.Load(cosmeticsBytes)
		if cosmeticsErr != nil {
			return production.CatalogBundle{}, cosmeticsErr
		}
		bundle.Cosmetics = cosmetics
	}
	if gardenBytes, active := artifacts[garden.ArtifactName]; active {
		// Server Garden SG1 rules 9 and 10: payout resources, Fiscal unlock and
		// host rows, and copy keys resolve against the pinned bundle.
		declarations := garden.Declarations{CopyKeys: map[string]struct{}{}, ResourceIDs: map[string]struct{}{},
			FiscalUnlockIDs: map[string]struct{}{}, FiscalGeneratorIDs: map[string]struct{}{}}
		for _, key := range copykeys.All() {
			declarations.CopyKeys[key] = struct{}{}
		}
		for _, resource := range economyCatalog.Resources() {
			declarations.ResourceIDs[resource.ID] = struct{}{}
		}
		if bundle.Fiscal == nil {
			return production.CatalogBundle{}, garden.ErrInvalidCatalog
		}
		for _, row := range bundle.Fiscal.UnlockRows() {
			declarations.FiscalUnlockIDs[row.UnlockID] = struct{}{}
		}
		for _, row := range bundle.Fiscal.GeneratorLevelRows() {
			declarations.FiscalGeneratorIDs[row.GeneratorID] = struct{}{}
		}
		gardenCatalog, gardenErr := garden.LoadCatalog(gardenBytes, declarations)
		if gardenErr != nil {
			return production.CatalogBundle{}, gardenErr
		}
		bundle.Garden = gardenCatalog
	}
	return bundle, nil
}

func validArtifactNames(artifacts map[string][]byte) bool {
	base := [...]string{"categories", "commons", "economy", "factions", "guilds", "prestige", "routes"}
	allowed := make(map[string]bool, len(base)+12)
	for _, name := range base {
		allowed[name] = true
		if len(artifacts[name]) == 0 {
			return false
		}
	}
	for _, name := range [...]string{"achievements", "cosmetics", "curriculum", "doctrines", "fiscal", "meters", "minigame_api", "minigames", "opportunities", "pets", "pitch", "pet_species", "relevance", "reputation_tree", "soul", "typer", "arcade", garden.ArtifactName} {
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
	_, typerActive := artifacts["typer"]
	_, arcadeActive := artifacts["arcade"]
	_, opportunitiesActive := artifacts["opportunities"]
	_, relevanceActive := artifacts["relevance"]
	_, curriculumActive := artifacts["curriculum"]
	_, reputationActive := artifacts["reputation_tree"]
	_, petSpeciesActive := artifacts["pet_species"]
	_, cosmeticsActive := artifacts["cosmetics"]
	_, gardenActive := artifacts[garden.ArtifactName]
	if meters != achievements || doctrines && !meters || minigames && !meters || pets && !minigames || fiscalActive && !pets ||
		soulActive && !fiscalActive || pitchActive && !soulActive || minigameAPIActive && !pitchActive || typerActive && !minigameAPIActive || arcadeActive && !minigameAPIActive ||
		opportunitiesActive && !doctrines || relevanceActive && !opportunitiesActive || curriculumActive && !relevanceActive || reputationActive && !minigameAPIActive ||
		petSpeciesActive && (!reputationActive || !pets) || cosmeticsActive && !petSpeciesActive ||
		gardenActive && (!cosmeticsActive || !fiscalActive) {
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
	if typerActive {
		want++
	}
	if arcadeActive {
		want++
	}
	if opportunitiesActive {
		want++
	}
	if relevanceActive {
		want++
	}
	if curriculumActive {
		want++
	}
	if reputationActive {
		want++
	}
	if petSpeciesActive {
		want++
	}
	if cosmeticsActive {
		want++
	}
	if gardenActive {
		want++
	}
	return len(artifacts) == want
}

// validArcadeChain is AR-P1's composition check: arcade-engine definition
// rows exist exactly when the arcade artifact does; every stage toy resolves
// to such a definition at the pinned engine version; and, when minigame_api
// is present, each arcade definition is exactly one of its tenants.
func validArcadeChain(bundle production.CatalogBundle, api *minigameapi.Catalog) bool {
	arcadeDefinitions := map[string]bool{}
	for _, id := range bundle.Minigames.MinigameIDs() {
		definition, _ := bundle.Minigames.Definition(id)
		if definition.EngineRef != arcade.MineGridEngineRef && definition.EngineRef != arcade.SnakeEngineRef {
			continue
		}
		if bundle.Arcade == nil || definition.EngineVersion != arcade.EngineVersion ||
			api != nil && !api.SupportsTenant(definition.MinigameID, definition.EngineRef, definition.EngineVersion) {
			return false
		}
		arcadeDefinitions[id] = true
	}
	if bundle.Arcade == nil {
		return true
	}
	if len(arcadeDefinitions) == 0 {
		return false
	}
	for _, stage := range bundle.Arcade.Container.Stages {
		for _, toy := range stage.Toys {
			if !arcadeDefinitions[toy] {
				return false
			}
		}
	}
	return true
}
