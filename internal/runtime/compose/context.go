package compose

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"envarium.dev/envarium/internal/platform/dockerengine"
	"envarium.dev/envarium/internal/runtime/domain"
	"github.com/moby/moby/client"
)

const commandOutputLimit = 4 << 20

type dockerConnection = dockerengine.Connection

type contextResolver struct {
	runner commandRunner
}

func (r contextResolver) Resolve(ctx context.Context, requested, workingDirectory string) (dockerConnection, error) {
	if requested == "" && os.Getenv(client.EnvOverrideHost) != "" && os.Getenv("DOCKER_CONTEXT") == "" {
		return dockerConnection{FromEnv: true}, nil
	}
	arguments := []string{"context", "inspect"}
	if requested != "" {
		arguments = append(arguments, requested)
	}
	var stdout, stderr limitedBuffer
	err := r.runner.Run(ctx, domain.Command{Executable: "docker", Arguments: arguments, WorkingDirectory: workingDirectory}, &stdout, &stderr)
	if err != nil {
		return dockerConnection{}, commandError("inspect Docker context", err, stderr.String())
	}
	return dockerengine.DecodeContext(stdout.Bytes())
}

type limitedBuffer struct {
	bytes.Buffer
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	remaining := commandOutputLimit - b.Len()
	if remaining <= 0 {
		return len(value), nil
	}
	if len(value) > remaining {
		_, _ = b.Buffer.Write(value[:remaining])
		return len(value), nil
	}
	return b.Buffer.Write(value)
}

func commandError(action string, err error, detail string) error {
	detail = strings.TrimSpace(strings.ReplaceAll(detail, "\x00", ""))
	if len(detail) > 2048 {
		detail = detail[:2048] + "..."
	}
	if detail == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, detail)
}
