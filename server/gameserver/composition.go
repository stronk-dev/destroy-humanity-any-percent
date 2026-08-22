package gameserver

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/commonsprojection"
	"cloud-clicker/server/deploymentconfig"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/gameui"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/replayverify"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
	"cloud-clicker/server/transport"
)

var (
	ErrComposition    = errors.New("invalid gameserver composition")
	activityIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	uuidPattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

type CompositionConfig struct {
	DB                 *sql.DB
	RepositoryRoot     string
	ServerID           string
	ActivityBracket    string
	PublicOrigin       string
	TrustedProxyHops   int
	SigningKeys        account.SigningKeys
	BootstrapKeys      account.BootstrapReceiptKeys
	Clock              func() time.Time
	Random             io.Reader
	Logger             *slog.Logger
	GuildNameAdditions []byte
}

type Composition struct {
	CurrentHash          string
	Server               *Server
	Node                 *transport.Node
	Accounts             *account.Repository
	Production           *production.Service
	Guilds               *guild.Service
	Minigames            *minigame.Service
	GameUI               *gameui.Projector
	Commons              *commonsprojection.Projector
	Verification         *replayverify.Repository
	LeaderboardProjector *leaderboard.QueueProjector
	Clearing             *ClearingDriver
	Catalogs             *runtimeCatalogs
}

type runtimeCatalogs struct {
	replay  production.ReplayCatalogSet
	commons commons.CatalogSet
}

func newRuntimeCatalogs(replay production.ReplayCatalogSet) (*runtimeCatalogs, error) {
	if len(replay) == 0 {
		return nil, ErrComposition
	}
	result := &runtimeCatalogs{replay: replay, commons: commons.CatalogSet{}}
	for hash, bundle := range replay {
		policy, ok := bundle.Commons.(commonsbinding.ReplayPolicy)
		if !ok || policy.Catalog == nil || bundle.Economy == nil || bundle.Routes == nil || bundle.Prestige == nil || bundle.Faction == nil || bundle.Guild == nil {
			return nil, ErrComposition
		}
		result.commons[hash] = policy.Catalog
	}
	return result, nil
}

func (catalogs *runtimeCatalogs) bundle(hash string) (production.CatalogBundle, bool) {
	if catalogs == nil {
		return production.CatalogBundle{}, false
	}
	return catalogs.replay.ResolveReplayCatalogs(hash)
}
func (catalogs *runtimeCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	bundle, ok := catalogs.bundle(hash)
	return bundle.Economy, ok
}
func (catalogs *runtimeCatalogs) ValidateState(hash string, state *save.State) error {
	bundle, ok := catalogs.bundle(hash)
	if !ok {
		return ErrComposition
	}
	if err := bundle.Faction.ValidateState(state); err != nil {
		return err
	}
	return bundle.ValidateFoundationState(state)
}
func (catalogs *runtimeCatalogs) ResolveReplayCatalogs(hash string) (production.CatalogBundle, bool) {
	return catalogs.bundle(hash)
}
func (catalogs *runtimeCatalogs) ResolveTenantContent(constantsHash, engineRef, engineVersion string) (minigame.TenantContent, bool) {
	if catalogs == nil {
		return minigame.TenantContent{}, false
	}
	return catalogs.replay.ResolveTenantContent(constantsHash, engineRef, engineVersion)
}
func (catalogs *runtimeCatalogs) ResolvePrestige(hash string) (*prestigecore.Policy, bool) {
	bundle, ok := catalogs.bundle(hash)
	return bundle.Prestige, ok
}
func (catalogs *runtimeCatalogs) ResolveFaction(hash string) (*faction.Catalog, bool) {
	bundle, ok := catalogs.bundle(hash)
	return bundle.Faction, ok
}
func (catalogs *runtimeCatalogs) ResolveRoutes(hash string) (*routes.Catalog, bool) {
	bundle, ok := catalogs.bundle(hash)
	return bundle.Routes, ok
}
func (catalogs *runtimeCatalogs) ResolveGuild(hash string) (*guild.Catalog, bool) {
	bundle, ok := catalogs.bundle(hash)
	return bundle.Guild, ok
}
func (catalogs *runtimeCatalogs) ResolveCatchupCeilingMS(hash string) (int64, bool) {
	bundle, ok := catalogs.bundle(hash)
	if !ok || bundle.Prestige.CatchupCeilingMS <= 0 {
		return 0, false
	}
	return bundle.Prestige.CatchupCeilingMS, true
}
func (catalogs *runtimeCatalogs) ResolveCommons(hash string) (*commons.Catalog, bool) {
	return catalogs.commons.ResolveCommons(hash)
}
func (catalogs *runtimeCatalogs) CompactTitheBand(hash string) (int64, int64, bool) {
	return catalogs.commons.CompactTitheBand(hash)
}
func (catalogs *runtimeCatalogs) GuildHealthWindowMS(hash string) (int64, bool) {
	return catalogs.commons.GuildHealthWindowMS(hash)
}

type shardAssignment struct{ serverID, activityBracket string }

func (assignment shardAssignment) ResolveAssignment(string) (commonsprojection.AssignmentContext, bool) {
	return commonsprojection.AssignmentContext{ServerID: assignment.serverID, ActivityBracket: assignment.activityBracket}, true
}

type channelMemberships struct {
	guilds  *guild.Service
	commons *commonsprojection.Projector
}

func (memberships channelMemberships) GuildMember(accountID, guildID string) bool {
	return memberships.guilds.GuildMember(accountID, guildID)
}
func (memberships channelMemberships) CohortMember(founderID, cohortID string) bool {
	return memberships.commons.CohortMember(founderID, cohortID)
}
func (channelMemberships) MatchParticipant(string, string) bool { return false }

type logInvariantSink struct{ logger *slog.Logger }

func (sink logInvariantSink) ReportRelayInvariant(value transport.RelayInvariant) {
	sink.logger.Error("relay invariant", "kind", value.Kind, "founder_id", value.FounderID, "detail", value.Detail)
}
func (sink logInvariantSink) ReportVerificationInvariant(value replayverify.VerificationInvariant) {
	sink.logger.Error("verification invariant", "kind", value.Kind, "stream_id", value.StreamID, "run_seq", value.RunSeq, "detail", value.Detail)
}

func Compose(ctx context.Context, config CompositionConfig) (*Composition, error) {
	if config.DB == nil || config.RepositoryRoot == "" || !uuidPattern.MatchString(config.ServerID) || !activityIDPattern.MatchString(config.ActivityBracket) {
		return nil, ErrComposition
	}
	if config.Clock == nil {
		config.Clock = time.Now
	}
	if config.Logger == nil {
		config.Logger = slog.New(slog.DiscardHandler)
	}
	allowedOrigins, trustedProxyHops, err := deploymentBoundary(config)
	if err != nil {
		return nil, err
	}
	if err := save.Migrate(ctx, config.DB); err != nil {
		return nil, err
	}
	seed, err := epochseed.Load(config.RepositoryRoot)
	if err != nil {
		return nil, err
	}
	epochRepository, err := leaderboard.NewRepository(config.DB, config.RepositoryRoot)
	if err != nil {
		return nil, err
	}
	synchronizer, err := leaderboard.NewSeedSynchronizer(epochRepository, seed, config.Clock)
	if err != nil {
		return nil, err
	}
	if hash, err := synchronizer.Sync(ctx); err != nil || hash != seed.Hash {
		return nil, errors.Join(ErrComposition, err)
	}
	replaySet, err := replaycatalog.LoadDatabase(ctx, config.DB)
	if err != nil {
		return nil, err
	}
	catalogs, err := newRuntimeCatalogs(replaySet)
	if err != nil {
		return nil, err
	}
	current, ok := catalogs.bundle(seed.Hash)
	if !ok {
		return nil, ErrComposition
	}
	if err := commonsbinding.Validate(catalogs.commons[seed.Hash], current.Economy); err != nil {
		return nil, err
	}

	baseline, err := os.ReadFile(filepath.Join(config.RepositoryRoot, "moderation", "guild-names.txt"))
	if err != nil {
		return nil, err
	}
	names, err := guild.NewDenylistNameValidator(baseline, config.GuildNameAdditions)
	if err != nil {
		return nil, err
	}
	guildService, err := guild.NewService(config.DB, current.Guild, names, config.Clock)
	if err != nil {
		return nil, err
	}
	commonsProjector, err := commonsprojection.New(config.DB, shardAssignment{config.ServerID, config.ActivityBracket}, catalogs)
	if err != nil {
		return nil, err
	}
	if err := commonsProjector.AttachGuildHealth(guild.HealthReader{DB: config.DB, Catalogs: catalogs, Windows: catalogs, Clock: config.Clock}); err != nil {
		return nil, err
	}
	routeProjector, err := routeprojection.New(config.DB, catalogs)
	if err != nil {
		return nil, err
	}
	guildProjector, err := guild.NewProjector(config.DB, catalogs)
	if err != nil {
		return nil, err
	}
	store, err := save.NewStore(config.DB, catalogs, config.Logger)
	if err != nil {
		return nil, err
	}
	soulRecoveries, err := soul.NewRecoveryRepository(config.DB)
	if err != nil {
		return nil, err
	}
	minigameRepository, err := minigame.NewRepository(config.DB)
	if err != nil {
		return nil, err
	}
	minigameTenants, err := minigame.NewTenantRegistry(pitch.NewTenant())
	if err != nil {
		return nil, err
	}
	minigameService, err := minigame.NewService(minigameRepository, minigameTenants, catalogs)
	if err != nil {
		return nil, err
	}
	providers := production.CombinedContributionProviders{
		production.FrozenContributionProvider{DB: config.DB},
		commonsbinding.Provider{Catalogs: catalogs.commons, Snapshots: commonsProjector},
		faction.StockConsumptionProvider{Catalogs: catalogs, Members: guildService},
	}
	productionService, err := production.NewService(store, catalogs, providers, nil, config.Logger,
		production.WithRouteCatalogs(catalogs), production.WithRouteProjector(routeProjector),
		production.WithEventProjector(commonsProjector), production.WithEventProjector(guildProjector),
		production.WithCompactPolicies(catalogs), production.WithCommonsWeightResolver(commonsProjector),
		production.WithReplayCatalogs(catalogs), production.WithProgressionRuntime(catalogs),
		production.WithGuildRuntime(catalogs), production.WithGuildSettlements(guildService),
		production.WithMinigameActivity(minigameRepository),
		production.WithSoulRecovery(soulRecoveries),
		production.WithCurrentConstantsHash(seed.Hash))
	if err != nil {
		return nil, err
	}
	accounts, err := account.NewRepository(config.DB, catalogs, seed.Hash, config.SigningKeys, config.Clock, config.Random)
	if err != nil {
		return nil, err
	}
	if err := accounts.AttachFounderInitializer(production.FounderInitializer{Catalogs: catalogs}); err != nil {
		return nil, err
	}
	if err := accounts.AttachAccountDeletionParticipant(guildService); err != nil {
		return nil, err
	}
	apiConfig := account.Phase0APIConfig(config.BootstrapKeys)
	apiConfig.TrustedProxyHops = trustedProxyHops
	api, err := account.NewAPI(accounts, productionService, apiConfig)
	if err != nil {
		return nil, err
	}
	if err := api.AttachGuildIntents(guildService); err != nil {
		return nil, err
	}
	if err := api.AttachSoulRecoveries(productionService); err != nil {
		return nil, err
	}
	if err := api.AttachMinigames(minigameAPIAdapter{accounts: accounts, production: productionService, platform: minigameService}); err != nil {
		return nil, err
	}
	gameUIProjector, err := gameui.New(store, catalogs, providers)
	if err != nil {
		return nil, err
	}
	if err := api.AttachGameUI(gameUIProjector); err != nil {
		return nil, err
	}
	policyBytes, err := os.ReadFile(filepath.Join(config.RepositoryRoot, "balance", "transport", "phase0.json"))
	if err != nil {
		return nil, err
	}
	policy, err := transport.LoadPolicy(policyBytes)
	if err != nil {
		return nil, err
	}
	if allowedOrigins != nil {
		policy.AllowedOrigins = allowedOrigins
	}
	node, err := transport.NewNode(*policy, accounts, channelMemberships{guildService, commonsProjector})
	if err != nil {
		return nil, err
	}
	sink := logInvariantSink{logger: config.Logger}
	playerRelay, err := transport.NewPlayerRelay(store, node, sink)
	if err != nil {
		return nil, err
	}
	guildRelay, err := transport.NewGuildPresenceRelay(guildService, node, seed.Hash)
	if err != nil {
		return nil, err
	}
	verification, err := replayverify.NewRepository(config.DB, sink)
	if err != nil {
		return nil, err
	}
	boardProjector := leaderboard.NewQueueProjector()
	clearing, err := NewClearingDriver(config.DB, guildService, catalogs, config.Clock)
	if err != nil {
		return nil, err
	}
	worldSource := &databaseWorldSource{db: config.DB, serverID: config.ServerID, online: node, epochs: epochRepository}
	world, err := NewWorldAggregator(worldSource, node, seed.Hash, policy.WorldHz, config.Clock)
	if err != nil {
		return nil, err
	}
	server, err := New(config.DB, api.Router(), node, playerRelay, synchronizer, seed.Hash)
	if err != nil {
		return nil, err
	}

	verificationJob, err := NewPeriodicJob(100*time.Millisecond, func(ctx context.Context) error {
		for count := 0; count < 64; count++ {
			worked, err := verification.ProcessNext(ctx, boardProjector)
			if err != nil || !worked {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	presenceJob, err := NewPeriodicJob(25*time.Millisecond, func(ctx context.Context) error { _, err := guildRelay.Flush(ctx); return err })
	if err != nil {
		return nil, err
	}
	clearingJob, err := NewPeriodicJob(time.Duration(current.Guild.ClearingIntervalMS)*time.Millisecond, func(ctx context.Context) error { _, err := clearing.Tick(ctx); return err })
	if err != nil {
		return nil, err
	}
	sweepJob, err := NewPeriodicJob(time.Hour, func(ctx context.Context) error {
		_, err := guildService.SweepDisbanded(ctx, config.Clock(), 64)
		return err
	})
	if err != nil {
		return nil, err
	}
	sessionGCJob, err := NewPeriodicJob(time.Minute, func(ctx context.Context) error {
		return pruneExpiredCredentials(ctx, accounts, config.Clock(), 1_000)
	})
	if err != nil {
		return nil, err
	}
	if err := server.AttachJobs(world, verificationJob, presenceJob, clearingJob, sweepJob, sessionGCJob); err != nil {
		return nil, err
	}
	return &Composition{CurrentHash: seed.Hash, Server: server, Node: node, Accounts: accounts, Production: productionService, Guilds: guildService, Minigames: minigameService,
		GameUI: gameUIProjector, Commons: commonsProjector, Verification: verification, LeaderboardProjector: boardProjector, Clearing: clearing, Catalogs: catalogs}, nil
}

func deploymentBoundary(config CompositionConfig) ([]string, int, error) {
	if config.PublicOrigin == "" {
		if config.TrustedProxyHops != 0 {
			return nil, 0, ErrComposition
		}
		return nil, 0, nil
	}
	if config.TrustedProxyHops != 1 || !deploymentconfig.ValidProductionOrigin(config.PublicOrigin) {
		return nil, 0, ErrComposition
	}
	return []string{config.PublicOrigin}, 1, nil
}
