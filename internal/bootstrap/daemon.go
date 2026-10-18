// Package bootstrap composes process-level Envarium adapters and use cases.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	actionsAdapters "envarium.dev/envarium/internal/actions/adapters"
	actionsApplication "envarium.dev/envarium/internal/actions/application"
	agentsAdapters "envarium.dev/envarium/internal/agents/adapters"
	agentsApplication "envarium.dev/envarium/internal/agents/application"
	claudeProvider "envarium.dev/envarium/internal/agents/providers/claude"
	codexProvider "envarium.dev/envarium/internal/agents/providers/codex"
	openAIProvider "envarium.dev/envarium/internal/agents/providers/openai"
	catalog "envarium.dev/envarium/internal/catalog/application"
	diagnosticsAdapters "envarium.dev/envarium/internal/diagnostics/adapters"
	diagnosticsApplication "envarium.dev/envarium/internal/diagnostics/application"
	discoveryAdapters "envarium.dev/envarium/internal/discovery/adapters"
	environmentsAdapters "envarium.dev/envarium/internal/environments/adapters"
	environmentsApplication "envarium.dev/envarium/internal/environments/application"
	"envarium.dev/envarium/internal/foundation/buildinfo"
	observabilityAdapters "envarium.dev/envarium/internal/observability/adapters"
	observabilityApplication "envarium.dev/envarium/internal/observability/application"
	operations "envarium.dev/envarium/internal/operations/application"
	"envarium.dev/envarium/internal/operations/domain"
	overviewApplication "envarium.dev/envarium/internal/overview/application"
	"envarium.dev/envarium/internal/platform/daemonlog"
	hostPlatform "envarium.dev/envarium/internal/platform/host"
	"envarium.dev/envarium/internal/platform/sqlite"
	portsAdapters "envarium.dev/envarium/internal/ports/adapters"
	portsApplication "envarium.dev/envarium/internal/ports/application"
	routingAdapters "envarium.dev/envarium/internal/routing/adapters"
	routingApplication "envarium.dev/envarium/internal/routing/application"
	runtimeApplication "envarium.dev/envarium/internal/runtime/application"
	"envarium.dev/envarium/internal/runtime/compose"
	processRuntime "envarium.dev/envarium/internal/runtime/process"
	session "envarium.dev/envarium/internal/session/application"
	sourcecontrolAdapters "envarium.dev/envarium/internal/sourcecontrol/adapters"
	sourcecontrolApplication "envarium.dev/envarium/internal/sourcecontrol/application"
	"envarium.dev/envarium/internal/system/application"
	terminalAdapters "envarium.dev/envarium/internal/terminal/adapters"
	terminalApplication "envarium.dev/envarium/internal/terminal/application"
	terminalDomain "envarium.dev/envarium/internal/terminal/domain"
	"envarium.dev/envarium/internal/transport/httpapi"
	eventtransport "envarium.dev/envarium/internal/transport/websocket"
	workspaceApplication "envarium.dev/envarium/internal/workspace/application"
	"envarium.dev/envarium/web"
)

// RunDaemon starts the migrated local control plane and blocks until shutdown.
func RunDaemon(ctx context.Context, config Config) error {
	config.Logger = daemonLogger(config.Logger)
	if err := validateDaemonConfig(config); err != nil {
		return err
	}
	if err := os.MkdirAll(config.DataDir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	lock, err := acquireLock(filepath.Join(config.DataDir, "daemon.lock"))
	if err != nil {
		return err
	}
	redactor, internalLog, logger, err := prepareDaemonSupport(config)
	if err != nil {
		_ = lock.release()
		return err
	}
	config.Logger = logger
	defer closeInternalLog(config.Logger, internalLog)
	defer releaseDaemonLock(config.Logger, lock)

	database, settingsService, config, err := openConfiguredDatabase(ctx, config)
	if err != nil {
		return err
	}
	defer closeDatabase(config.Logger, database)

	journal := sqlite.NewJournal(database)
	catalogService := catalog.NewService(sqlite.NewCatalogRepository(database), discoveryAdapters.Defaults(), settingsService)
	gitAdapter := sourcecontrolAdapters.NewGit()
	gitService := sourcecontrolApplication.NewService(sourcecontrolApplication.NewCatalogSource(catalogService), gitAdapter)
	environmentService := environmentsApplication.NewService(
		environmentsAdapters.NewCatalogSource(catalogService), environmentsAdapters.NewSourceControlSource(gitService),
		sqlite.NewEnvironmentRepository(database),
	)
	runtimeSource := runtimeApplication.NewSchedulerSource(runtimeApplication.NewCatalogSource(catalogService, environmentsAdapters.NewRuntimeSource(environmentService)))
	composeDriver := compose.NewDriverWithArtifacts(filepath.Join(config.DataDir, "runtime", "compose"))
	defer closeComposeDriver(config.Logger, composeDriver)
	runtimeService := runtimeApplication.NewService(
		runtimeSource,
		composeDriver,
		processRuntime.NewDriverWithLogCapacity(ctx, sqlite.NewRunRepository(database), config.LogRingCapacity, redactor),
	)
	branchService := sourcecontrolApplication.NewBranchService(gitService, gitAdapter, runtimeService)
	logStore, err := sqlite.NewLogStore(database, sqlite.LogStoreConfig{
		Directory:    filepath.Join(config.DataDir, "logs"),
		SegmentBytes: config.LogSegmentBytes, RetentionAge: config.LogRetentionAge, RetentionBytes: config.LogRetentionBytes,
	}, redactor)
	if err != nil {
		return err
	}
	defer closeLogStore(config.Logger, logStore)
	logService := observabilityApplication.NewLogService(runtimeService, logStore)
	healthService := observabilityApplication.NewHealthService(runtimeSource, runtimeService, sqlite.NewHealthRepository(database), observabilityAdapters.NewHealthEvaluator())
	resourceService, err := observabilityApplication.NewResourceService(
		observabilityAdapters.NewResourceCatalogSource(catalogService),
		observabilityAdapters.NewResourceRuntimeSource(runtimeService, healthService),
		database,
		observabilityAdapters.NewDockerStorage(),
		observabilityApplication.ResourceConfig{
			SampleInterval: config.MetricSampleInterval, RawRetention: config.MetricRawRetention,
			MinuteRetention: config.MetricMinuteRetention, QuarterHourRetention: config.MetricQuarterHourRetention,
			MaximumHistoryPoints: config.MetricMaximumHistoryPoints,
			LogRetentionAge:      config.LogRetentionAge, LogRetentionBytes: config.LogRetentionBytes,
		},
	)
	if err != nil {
		return err
	}
	portDeclarations := portsApplication.NewCatalogSource(catalogService)
	portService := portsApplication.NewService(
		portDeclarations,
		portsAdapters.NewRuntimeBindings(catalogService, runtimeService),
		portsAdapters.NewOSListeners(),
		sqlite.NewPortReservationRepository(database),
		environmentsAdapters.NewPortLeaseSource(environmentService),
	)
	environmentRegistration := environmentsApplication.NewRegistrationCoordinator(
		environmentService, environmentsAdapters.NewPortDeclarations(portDeclarations), environmentsAdapters.NewPortAllocator(portService),
	)
	routeService := routingApplication.NewService(config.RoutingAddr != "")
	routeRegistry := routingAdapters.NewRegistry(routeService, routingAdapters.NewEnvironmentSource(environmentService))
	if _, err := routeRegistry.Refresh(ctx); err != nil {
		return err
	}
	environmentRuntime := environmentsAdapters.NewLifecycle(environmentService, runtimeService, func(refreshCtx context.Context) error {
		_, refreshErr := routeRegistry.Refresh(refreshCtx)
		return refreshErr
	})
	actionAudits := sqlite.NewActionAuditRepository(database)
	if err := actionAudits.Recover(ctx, time.Now().UTC()); err != nil {
		return err
	}
	actionCatalog := actionsApplication.NewCatalogSource(catalogService, settingsService)
	actionService := actionsApplication.NewService(
		actionCatalog,
		actionsAdapters.NewRunner(actionsAdapters.NewLauncher()),
		actionAudits,
	)
	providerRegistry, err := agentsApplication.NewRegistry(
		codexProvider.NewProposalProvider(codexProvider.ProposalConfig{Disabled: config.AICodexExecutable == "", Executable: config.AICodexExecutable, Model: config.AICodexModel, Redactor: redactor}),
		claudeProvider.NewProposalProvider(claudeProvider.ProposalConfig{Disabled: config.AIClaudeExecutable == "", Executable: config.AIClaudeExecutable, Model: config.AIClaudeModel, Redactor: redactor}),
		openAIProvider.NewProposalProvider(openAIProvider.ProposalConfig{Endpoint: config.AIOpenAIEndpoint, Model: config.AIOpenAIModel, APIKey: os.Getenv(config.AIOpenAIAPIKeyEnv), Redactor: redactor}),
	)
	if err != nil {
		return err
	}
	terminalService, err := terminalApplication.NewService(
		ctx,
		sqlite.NewTerminalSessionRepository(database),
		terminalAdapters.NewResolver(catalogService, actionCatalog, environmentService, config.AICodexExecutable, config.AIClaudeExecutable),
		terminalAdapters.NewPTY(),
		terminalApplication.DefaultConfig(),
	)
	if err != nil {
		return err
	}
	defer closeTerminalSessions(config.Logger, terminalService)
	aiService, err := agentsApplication.NewEnhancementService(
		catalogService, sqlite.NewAgentRunRepository(database), agentsAdapters.RepositoryReader{}, redactor, agentsAdapters.ManifestValidator{}, providerRegistry,
	)
	if err != nil {
		return err
	}
	build := buildinfo.Current()
	extensions, err := newConfiguredExtensions(ctx, database, catalogService, redactor, config.DataDir, build.Version, config.Logger)
	if err != nil {
		return err
	}
	operationRepository := sqlite.NewOperationRepository(database)
	diagnosticRepository := sqlite.NewDiagnosticRepository(database)
	diagnosticCollector := diagnosticsAdapters.NewCollector(
		catalogService, runtimeService, healthService, logService, portService, gitService, actionService, resourceService, operationRepository,
	)
	diagnosticService, err := diagnosticsApplication.NewService(
		diagnosticCollector, diagnosticsAdapters.NewStructuredProvider(providerRegistry), diagnosticRepository,
	)
	if err != nil {
		return err
	}
	launcher := actionsAdapters.NewLauncher()
	var workspaceService *workspaceApplication.Service
	coordinator := operations.NewCoordinator(ctx, operationRepository, journal, branchAwareExecutor{branches: branchService, fallback: operations.ExecutorFunc(
		func(operationCtx context.Context, operation domain.Operation, progress operations.Progress) error {
			return executeOperation(
				operationCtx, runtimeService, healthService, actionService, aiService, extensions.plugin, workspaceService,
				workspaceRecipeRunner{projects: runtimeSource, launcher: launcher}, environmentRuntime, operation, progress,
			)
		},
	)}, extensions.shared.telemetry)
	fleetService, remoteHandler, err := newFleetExpansion(
		config, database, extensions.shared.team, catalogService, runtimeService, environmentService, coordinator, build.Version,
	)
	if err != nil {
		return err
	}
	workspaceService = workspaceApplication.NewService(
		sqlite.NewWorkspaceRepository(database), &workspaceProjectOperator{operations: coordinator},
		workspaceHealthGate{health: healthService}, workspaceMemberValidator{runtime: runtimeSource},
	)
	automationService, err := diagnosticsApplication.NewAutomationService(
		diagnosticRepository, diagnosticCollector, diagnosticService, diagnosticCollector, diagnosticsAdapters.NewActionSubmitter(coordinator),
	)
	if err != nil {
		return err
	}
	if err := workspaceService.Recover(ctx); err != nil {
		return err
	}
	if err := coordinator.Recover(ctx); err != nil {
		return err
	}
	sessions := session.NewManager()
	system := application.NewQuery(database, buildinfo.Current(), time.Now())
	host := application.NewHostQuery(hostPlatform.NewObserverWithStorage(func() (int64, int64, bool) {
		summary, available := resourceService.CachedStorageSummary()
		return summary.Bytes, summary.ReclaimableBytes, available
	}))
	dependencies := httpapi.Dependencies{
		System: system, Host: host, Operations: coordinator, Sessions: sessions, Catalog: catalogService, Runtime: runtimeService,
		Health: healthService, LogService: logService, Events: eventtransport.NewEvents(journal), Logs: eventtransport.NewLogs(logService),
		Ports: portService, Git: gitService, Branches: branchService, Actions: actionService, Overview: overviewApplication.NewService(runtimeService, gitService, actionService, healthService),
		AI: aiService, Resources: resourceService, Plugins: extensions.plugin, Diagnostics: diagnosticService, Automations: automationService,
		Workspaces: workspaceService, Environments: environmentService, EnvironmentRegistration: environmentRegistration, Routes: routeRegistry,
		Fleet: fleetService, Team: extensions.shared.team, Telemetry: extensions.shared.telemetry, Settings: settingsService,
		Terminals: terminalService, Terminal: eventtransport.NewTerminal(terminalService, func(actorContext context.Context) terminalDomain.Owner {
			actorType, actorID := httpapi.RequestActor(actorContext)
			return terminalDomain.Owner{Type: actorType, ID: actorID}
		}),
		Web: web.Handler(), Logger: config.Logger,
	}
	servers, err := newLocalServers(config, dependencies, routingAdapters.NewProxy(routeRegistry), remoteHandler)
	if err != nil {
		return err
	}
	reconcileSink := runtimeReconciliationSink{runtime: runtimeService, journal: journal}
	var background sync.WaitGroup
	background.Add(8)
	go func() { defer background.Done(); runtimeSource.Run(ctx) }()
	go func() {
		defer background.Done()
		runtimeService.WatchAll(ctx, reconcileSink, func(projectID string, watchErr error) {
			config.Logger.Warn("runtime event watcher unavailable", "component", "runtime", "project_id", projectID, "error", watchErr)
		})
	}()
	extensions.shared.telemetry.Record(ctx, "daemon.started")
	go func() {
		defer background.Done()
		extensions.shared.telemetry.Run(ctx, 24*time.Hour)
	}()
	go func() {
		defer background.Done()
		healthService.Run(ctx, func(projectID string, healthErr error) {
			config.Logger.Warn("health observer unavailable", "component", "observability", "project_id", projectID, "error", healthErr)
		})
	}()
	go func() {
		defer background.Done()
		logService.Run(ctx, func(projectID string, logErr error) {
			config.Logger.Warn("log collector unavailable", "component", "observability", "project_id", projectID, "error", logErr)
		})
	}()
	go func() {
		defer background.Done()
		resourceService.Run(ctx, func(resourceErr error) {
			config.Logger.Warn("resource sampler unavailable", "component", "observability", "error", resourceErr)
		})
	}()
	go func() {
		defer background.Done()
		extensions.plugin.RunHealth(ctx, 30*time.Second)
	}()
	go func() {
		defer background.Done()
		automationService.Run(ctx, 5*time.Minute, func(projectID string, automationErr error) {
			config.Logger.Warn("automation evaluation unavailable", "component", "diagnostics", "project_id", projectID, "error", automationErr)
		})
	}()
	config.Logger.Info(
		"envarium daemon ready",
		"component", "bootstrap",
		"address", servers.browserAddress(),
		"ipc_address", servers.ipcAddress,
		"routing_address", config.RoutingAddr,
		"remote_address", servers.remoteAddress,
		"data_dir", config.DataDir,
	)
	runErr := servers.run(ctx, coordinator.Shutdown)
	background.Wait()
	return runErr
}

func daemonLogger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.Default()
	}
	return logger
}

func releaseDaemonLock(logger *slog.Logger, lock *lockFile) {
	if err := lock.release(); err != nil {
		logger.Error("release daemon lock", "component", "bootstrap", "error", err)
	}
}

func closeInternalLog(logger *slog.Logger, handler *daemonlog.FileHandler) {
	if err := handler.Close(); err != nil {
		logger.Error("close internal daemon log", "component", "bootstrap", "error", err)
	}
}

func closeDatabase(logger *slog.Logger, database *sqlite.Database) {
	if err := database.Close(); err != nil {
		logger.Error("close database", "component", "bootstrap", "error", err)
	}
}

func closeComposeDriver(logger *slog.Logger, driver *compose.Driver) {
	if err := driver.Close(); err != nil {
		logger.Warn("close Compose Engine clients", "error", err)
	}
}

func closeLogStore(logger *slog.Logger, store *sqlite.LogStore) {
	if err := store.Close(context.Background()); err != nil {
		logger.Error("close log store", "component", "observability", "error", err)
	}
}

func closeTerminalSessions(logger *slog.Logger, service *terminalApplication.Service) {
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		logger.Error("close terminal sessions", "component", "terminal", "error", err)
	}
}

func validateLoopbackAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("parse daemon address: %w", err)
	}
	if port == "" {
		return errors.New("daemon address requires a port")
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("daemon address must use a loopback host: %s", address)
	}
	return nil
}

func validateDaemonConfig(config Config) error {
	if err := validateLoopbackAddress(config.HTTPAddr); err != nil {
		return err
	}
	if config.RoutingAddr != "" {
		if err := validateLoopbackAddress(config.RoutingAddr); err != nil {
			return fmt.Errorf("validate local routing address: %w", err)
		}
	}
	if config.RemoteAddr != "" {
		if _, _, err := net.SplitHostPort(config.RemoteAddr); err != nil {
			return fmt.Errorf("validate remote agent address: %w", err)
		}
	}
	return validateRemoteConfig(config)
}
