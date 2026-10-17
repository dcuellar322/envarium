package compose

import (
	"context"
	"envarium.dev/envarium/internal/runtime/domain"
	"io"
	"testing"
)

func TestDefaultContextUsesOneFreshInspectCommand(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
	t.Setenv("DOCKER_CONTEXT", "")
	var calls int
	resolver := contextResolver{runner: scriptedRunner{run: func(command domain.Command, stdout, _ io.Writer) error {
		calls++
		if len(command.Arguments) != 2 || command.Arguments[0] != "context" || command.Arguments[1] != "inspect" {
			t.Fatalf("unexpected command: %v", command.Arguments)
		}
		_, err := io.WriteString(stdout, `[{"Name":"active","Endpoints":{"docker":{"Host":"unix:///tmp/docker.sock"}}}]`)
		return err
	}}}
	for range 2 {
		connection, err := resolver.Resolve(context.Background(), "", "/repo")
		if err != nil || connection.ContextName != "active" {
			t.Fatalf("context: %v %v", connection, err)
		}
	}
	if calls != 2 {
		t.Fatal("fresh context resolver was cached")
	}
}
