package compose

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"envarium.dev/envarium/internal/runtime/domain"
)

func TestMetadataMemoDetectsExactDependencyAndEnvironmentChanges(t *testing.T) {
	t.Setenv("DOCKER_CONFIG", t.TempDir())
	for _, key := range []string{"DOCKER_CONTEXT", "DOCKER_HOST", "COMPOSE_ENV_FILES", "COMPOSE_PROJECT_NAME", "COMPOSE_PROFILES"} {
		t.Setenv(key, "")
	}
	t.Setenv("COMPOSE_DISABLE_ENV_FILE", "false")
	if err := os.Unsetenv("COMPOSE_DISABLE_ENV_FILE"); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	write := func(name, value string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("compose.yaml", "name: ${MEMO_NAME:-original}\nservices:\n  api:\n    image: alpine\n    env_file:\n      - service.env\n      - {path: optional.env, required: false}\n    label_file: labels.env\n")
	write("service.env", "VALUE=one\n")
	write("labels.env", "example=one\n")
	reader := configReader{metadata: newMetadataMemo()}
	project := domain.ProjectRuntime{Root: root, Compose: &domain.ComposeRuntime{Files: []string{"compose.yaml"}}}
	load := func(want string) {
		t.Helper()
		got, err := reader.loadObservation(context.Background(), project)
		if err != nil || got.ProjectName != want {
			t.Fatalf("name=%q err=%v want=%q", got.ProjectName, err, want)
		}
	}
	load("original")
	files := []string{filepath.Join(root, "compose.yaml")}
	key := metadataKey(files, root, "")
	if _, ok := reader.metadata.lookup(context.Background(), key); !ok {
		t.Fatal("validated local graph was not memoized")
	}
	stat, err := os.Stat(filepath.Join(root, "service.env"))
	if err != nil {
		t.Fatal(err)
	}
	write("service.env", "VALUE=two\n")
	if err := os.Chtimes(filepath.Join(root, "service.env"), stat.ModTime(), stat.ModTime()); err != nil {
		t.Fatal(err)
	}
	if _, ok := reader.metadata.lookup(context.Background(), key); ok {
		t.Fatal("same-mtime edit was hidden")
	}
	load("original")
	write(".env", "MEMO_NAME=dotenv\n")
	load("dotenv")
	t.Setenv("MEMO_NAME", "environment")
	load("environment")
	if err := os.Remove(filepath.Join(root, "service.env")); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.loadObservation(context.Background(), project); err == nil {
		t.Fatal("removed required env file hidden")
	}
	write("service.env", "VALUE=restored\n")
	load("environment")
	write("optional.env", "OPTIONAL=created\n")
	if _, ok := reader.metadata.lookup(context.Background(), metadataKey(files, root, "")); ok {
		t.Fatal("optional file creation hidden")
	}
	load("environment")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reader.loadObservation(ctx, project); err == nil {
		t.Fatal("cancelled memo load succeeded")
	}
}

func TestObservationInputsFallBackForUnknownGraphs(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "compose.yaml")
	for _, contents := range []string{
		"include: [other.yaml]\nservices: {}\n",
		"services:\n  api:\n    extends: {file: other.yaml, service: api}\n",
		"services:\n  api:\n    env_file: ${DYNAMIC_PATH}\n",
		"services:\n  api:\n    label_file: {unexpected: syntax}\n",
		"services:\n  api: &api\n    image: alpine\n  copy: *api\n",
	} {
		if err := os.WriteFile(file, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
		if got := observationInputs(context.Background(), []string{file}, root); got != nil {
			t.Fatalf("unknown dependency graph memoized: %q", contents)
		}
	}
	if err := os.WriteFile(file, []byte("services: {api: {image: alpine}}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COMPOSE_ENV_FILES", "custom.env")
	if got := observationInputs(context.Background(), []string{file}, root); got != nil {
		t.Fatal("explicit dotenv graph memoized")
	}
}
