package bootstrap

import (
	"context"
	"log/slog"
	"path/filepath"

	catalogApplication "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	observabilityAdapters "envarium.dev/envarium/internal/observability/adapters"
	"envarium.dev/envarium/internal/platform/sqlite"
	pluginsAdapters "envarium.dev/envarium/internal/plugins/adapters"
	pluginsApplication "envarium.dev/envarium/internal/plugins/application"
)

func newPluginService(
	ctx context.Context,
	database *sqlite.Database,
	catalog *catalogApplication.Service,
	redactor *observabilityAdapters.Redactor,
	dataDir, hostVersion string,
	logger *slog.Logger,
) (*pluginsApplication.Service, error) {
	service, err := pluginsApplication.NewService(
		sqlite.NewPluginRepository(database),
		pluginsAdapters.NewDirectoryDiscovery(filepath.Join(dataDir, "plugins")),
		pluginsAdapters.NewProcessRunner(hostVersion, redactor),
		pluginsAdapters.NewProjectLookup(func(projectCtx context.Context, id string) (pluginsApplication.Project, error) {
			project, projectErr := catalog.GetProject(projectCtx, id)
			if projectErr != nil {
				return pluginsApplication.Project{}, projectErr
			}
			return pluginsApplication.Project{
				ID: project.ID, DisplayName: project.DisplayName, Root: project.PrimaryLocation,
				Trusted: project.TrustState == catalogDomain.TrustTrusted,
			}, nil
		}),
	)
	if err != nil {
		return nil, err
	}
	if _, err := service.Refresh(ctx); err != nil {
		logger.Warn("plugin discovery unavailable", "component", "plugins", "error", err)
	}
	return service, nil
}
