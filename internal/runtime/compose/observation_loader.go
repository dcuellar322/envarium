package compose

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"envarium.dev/envarium/internal/platform/dockerengine"
	"envarium.dev/envarium/internal/runtime/domain"
	composecli "github.com/compose-spec/compose-go/v2/cli"
)

// loadObservation resolves only read-side metadata with the reference loader.
// Installed CLI normalization remains authoritative for every lifecycle plan.
func (r configReader) loadObservation(ctx context.Context, project domain.ProjectRuntime) (normalizedConfig, error) {
	if err := ctx.Err(); err != nil {
		return normalizedConfig{}, err
	}
	if project.Compose == nil || len(project.Compose.Files) == 0 {
		return normalizedConfig{}, fmt.Errorf("compose runtime requires at least one file")
	}
	connection, err := dockerengine.ReadConnection(ctx, project.Compose.Context, project.Root)
	if err != nil {
		return normalizedConfig{}, err
	}
	arguments, err := composeBaseArguments(project, connection, project.Compose.ProjectName)
	if err != nil {
		return normalizedConfig{}, err
	}
	files := make([]string, 0, len(project.Compose.Files)+1)
	for _, file := range project.Compose.Files {
		if !filepath.IsAbs(file) {
			file = filepath.Join(project.Root, file)
		}
		files = append(files, file)
	}
	if len(project.Compose.PortOverrides) > 0 {
		overlay, err := writePortOverride(r.artifactDirectory, project)
		if err != nil {
			return normalizedConfig{}, err
		}
		files = append(files, overlay)
		arguments = appendComposeFile(arguments, overlay)
	}
	key := metadataKey(files, project.Root, project.Compose.ProjectName)
	if config, ok := r.metadata.lookup(ctx, key); ok {
		config.Connection, config.BaseArgs = connection, arguments
		return config, nil
	}
	inputs := observationInputs(ctx, files, project.Root)

	envFiles := []string{}
	if value := os.Getenv("COMPOSE_ENV_FILES"); value != "" {
		envFiles = strings.Split(value, ",")
	}
	options, err := composecli.NewProjectOptions(files,
		composecli.WithWorkingDirectory(project.Root), composecli.WithOsEnv,
		composecli.WithEnv([]string{"PWD=" + project.Root}),
		composecli.WithEnvFiles(envFiles...), composecli.WithDotEnv,
		composecli.WithDefaultProfiles(),
		composecli.WithName(project.Compose.ProjectName),
	)
	if err != nil {
		return normalizedConfig{}, fmt.Errorf("resolve Compose observation inputs: %w", err)
	}
	// No remote resource loader is registered: includes/extends remain local.
	model, err := options.LoadProject(ctx)
	if err != nil {
		return normalizedConfig{}, fmt.Errorf("load Compose observation: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return normalizedConfig{}, err
	}
	services := model.ServiceNames()
	sort.Strings(services)
	if model.Name == "" || len(services) == 0 {
		return normalizedConfig{}, fmt.Errorf("compose observation has no name or services")
	}
	config := normalizedConfig{ProjectName: model.Name, Services: services, Connection: connection, BaseArgs: arguments}
	if len(inputs) > 0 && unchangedInputs(ctx, inputs) {
		r.metadata.store(key, inputs, config)
	}
	return config, nil
}
