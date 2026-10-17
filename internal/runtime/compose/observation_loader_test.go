package compose

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestReferenceObservationResolvesEnvMergeIncludesAndProfiles(t *testing.T) {
	t.Setenv("DOCKER_CONFIG", t.TempDir())
	t.Setenv("DOCKER_HOST", "")
	t.Setenv("DOCKER_CONTEXT", "")
	t.Setenv("COMPOSE_PROJECT_NAME", "")
	t.Setenv("COMPOSE_PROFILES", "")
	t.Setenv("COMPOSE_ENV_FILES", "")
	root := t.TempDir()
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(".env", "OBSERVATION_NAME=from_dotenv\n")
	write("base.yaml", "name: ${OBSERVATION_NAME}\ninclude: [included.yaml]\nservices:\n  api:\n    image: alpine\n  optional:\n    image: alpine\n    profiles: [extra]\n")
	write("included.yaml", "services:\n  worker:\n    image: alpine\n")
	write("override.yaml", "services:\n  api:\n    environment: {EXAMPLE: value}\n")
	reader := configReader{}
	project := domain.ProjectRuntime{Root: root, Compose: &domain.ComposeRuntime{Files: []string{"base.yaml", "override.yaml"}}}
	got, err := reader.loadObservation(context.Background(), project)
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectName != "from_dotenv" || !reflect.DeepEqual(got.Services, []string{"api", "worker"}) {
		t.Fatalf("metadata: %v", got)
	}
	project.Compose.ProjectName = "explicit"
	got, err = reader.loadObservation(context.Background(), project)
	if err != nil || got.ProjectName != "explicit" {
		t.Fatalf("explicit name: %v %v", got, err)
	}
	write("included.yaml", "services:\n  changed:\n    image: alpine\n")
	got, err = reader.loadObservation(context.Background(), project)
	if err != nil || !reflect.DeepEqual(got.Services, []string{"api", "changed"}) {
		t.Fatalf("include edit: %v %v", got, err)
	}
	write("included.yaml", "invalid: [")
	if _, err := reader.loadObservation(context.Background(), project); err == nil {
		t.Fatal("invalid include accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reader.loadObservation(ctx, project); err == nil {
		t.Fatal("cancelled load succeeded")
	}
	project.Compose.Files = []string{"../outside.yaml"}
	if _, err := reader.loadObservation(context.Background(), project); err == nil {
		t.Fatal("file escaped root")
	}
}

func TestReferenceObservationMatchesInstalledCLI(t *testing.T) {
	if os.Getenv("ENVARIUM_COMPOSE_PARITY") != "1" {
		t.Skip("opt-in comparison with installed Docker Compose; never launches containers")
	}
	root := t.TempDir()
	file := filepath.Join(root, "compose.yaml")
	contents := "name: parity-fixture\nservices:\n  api:\n    image: alpine\n  optional:\n    image: alpine\n    profiles: [extra]\n"
	if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	runner := osCommandRunner{}
	reader := configReader{metadata: newMetadataMemo(), runner: runner, contexts: contextResolver{runner: runner}}
	project := domain.ProjectRuntime{Root: root, Compose: &domain.ComposeRuntime{Files: []string{"compose.yaml"}}}
	compareObservationCLI(t, reader, project)
}

func compareObservationCLI(t *testing.T, reader configReader, project domain.ProjectRuntime) {
	t.Helper()
	want, err := reader.Normalize(context.Background(), project)
	if err != nil {
		t.Fatal("installed CLI normalization failed")
	}
	got, err := reader.loadObservation(context.Background(), project)
	if err != nil {
		t.Fatalf("reference loader failed: %v", err)
	}
	if got.ProjectName != want.ProjectName || !reflect.DeepEqual(got.Services, want.Services) || got.Connection != want.Connection || !reflect.DeepEqual(got.BaseArgs, want.BaseArgs) {
		t.Fatalf("observation metadata differs: got=%+v want=%+v", got, want)
	}
}

func TestActualProjectsObservationParity(t *testing.T) {
	path := os.Getenv("ENVARIUM_COMPOSE_PARITY_INPUTS")
	if path == "" {
		t.Skip("opt-in private runtime input document")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var projects []domain.ProjectRuntime
	if err := json.Unmarshal(contents, &projects); err != nil {
		t.Fatal(err)
	}
	runner := osCommandRunner{}
	reader := configReader{metadata: newMetadataMemo(), runner: runner, contexts: contextResolver{runner: runner}}
	for _, project := range projects {
		t.Run(project.ProjectID, func(t *testing.T) {
			compareObservationCLI(t, reader, project)
			compareObservationCLI(t, reader, project)
		})
	}
	t.Logf("memoized local graphs: %d of %d", len(reader.metadata.entries), len(projects))
}
