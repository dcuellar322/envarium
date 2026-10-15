package bootstrap

import (
	"context"
	"log/slog"
	"net/http"
	"runtime"

	catalogApplication "envarium.dev/envarium/internal/catalog/application"
	environmentsApplication "envarium.dev/envarium/internal/environments/application"
	fleetAdapters "envarium.dev/envarium/internal/fleet/adapters"
	fleetApplication "envarium.dev/envarium/internal/fleet/application"
	fleetDomain "envarium.dev/envarium/internal/fleet/domain"
	observabilityAdapters "envarium.dev/envarium/internal/observability/adapters"
	operationsApplication "envarium.dev/envarium/internal/operations/application"
	"envarium.dev/envarium/internal/platform/sqlite"
	pluginsApplication "envarium.dev/envarium/internal/plugins/application"
	runtimeApplication "envarium.dev/envarium/internal/runtime/application"
	teamAdapters "envarium.dev/envarium/internal/team/adapters"
	teamApplication "envarium.dev/envarium/internal/team/application"
	telemetryAdapters "envarium.dev/envarium/internal/telemetry/adapters"
	telemetryApplication "envarium.dev/envarium/internal/telemetry/application"
)

type sharedConfiguration struct {
	team      *teamApplication.Service
	telemetry *telemetryApplication.Service
}

type configuredExtensions struct {
	shared sharedConfiguration
	plugin *pluginsApplication.Service
}

func newConfiguredExtensions(
	ctx context.Context,
	database *sqlite.Database,
	catalogService *catalogApplication.Service,
	redactor *observabilityAdapters.Redactor,
	dataDir, version string,
	logger *slog.Logger,
) (configuredExtensions, error) {
	shared, err := newSharedConfiguration(database, version)
	if err != nil {
		return configuredExtensions{}, err
	}
	pluginService, err := newPluginService(ctx, database, catalogService, redactor, dataDir, version, logger)
	if err != nil {
		return configuredExtensions{}, err
	}
	return configuredExtensions{shared: shared, plugin: pluginService}, nil
}

func newSharedConfiguration(database *sqlite.Database, version string) (sharedConfiguration, error) {
	teamService, err := teamApplication.NewService(sqlite.NewTeamRepository(database), teamAdapters.ManifestValidator{})
	if err != nil {
		return sharedConfiguration{}, err
	}
	telemetryService, err := telemetryApplication.NewService(
		sqlite.NewTelemetryRepository(database), telemetryAdapters.NewHTTPSender(),
		telemetryApplication.Build{Version: version}, teamService,
	)
	if err != nil {
		return sharedConfiguration{}, err
	}
	return sharedConfiguration{team: teamService, telemetry: telemetryService}, nil
}

func newFleetExpansion(
	config Config,
	database *sqlite.Database,
	teamService *teamApplication.Service,
	catalogService *catalogApplication.Service,
	runtimeService *runtimeApplication.Service,
	environmentService *environmentsApplication.Service,
	coordinator *operationsApplication.Coordinator,
	version string,
) (*fleetApplication.Service, http.Handler, error) {
	fleetService, err := fleetApplication.NewService(sqlite.NewFleetRepository(database), fleetAdapters.NewHTTPSPeerClient(), teamService)
	if err != nil {
		return nil, nil, err
	}
	if config.RemoteAddr == "" {
		return fleetService, nil, nil
	}
	controllers, err := parseRemoteControllers(config.RemoteControllers)
	if err != nil {
		return nil, nil, err
	}
	agentService, err := fleetApplication.NewAgentService(fleetDomain.Identity{
		ProtocolVersion: fleetDomain.ProtocolVersion, MachineID: config.RemoteMachineID, Name: config.RemoteMachineName,
		Version: version, OS: runtime.GOOS, Architecture: runtime.GOARCH,
		Capabilities: append([]fleetDomain.Capability(nil), fleetDomain.KnownCapabilities...),
	}, fleetAdapters.NewLocalInventory(catalogService, runtimeService, environmentService), fleetAdapters.NewLocalOperator(coordinator), controllers, teamService)
	if err != nil {
		return nil, nil, err
	}
	return fleetService, fleetAdapters.NewAgentHandler(agentService), nil
}
