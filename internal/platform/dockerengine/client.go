// Package dockerengine resolves the operator's Docker context for SDK observers.
package dockerengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/moby/client"
)

// Connection describes a Docker CLI context without retaining credentials.
type Connection struct {
	ContextName   string
	Host          string
	SkipTLSVerify bool
	TLSPath       string
	FromEnv       bool
}

// NewClient connects SDK observers to the same active context as the Docker CLI.
func NewClient(ctx context.Context) (*client.Client, error) {
	if os.Getenv(client.EnvOverrideHost) != "" && os.Getenv("DOCKER_CONTEXT") == "" {
		return (Connection{FromEnv: true}).NewClient()
	}
	command := exec.CommandContext(ctx, "docker", "context", "inspect")
	command.WaitDelay = time.Second
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("inspect active Docker context: %w", err)
	}
	connection, err := DecodeContext(output)
	if err != nil {
		return nil, err
	}
	return connection.NewClient()
}

// DecodeContext validates the non-secret endpoint metadata returned by Docker.
func DecodeContext(contents []byte) (Connection, error) {
	if len(contents) > 1<<20 {
		return Connection{}, errors.New("docker context document exceeds 1 MiB")
	}
	var documents []struct {
		Name      string `json:"Name"`
		Endpoints map[string]struct {
			Host          string `json:"Host"`
			SkipTLSVerify bool   `json:"SkipTLSVerify"`
		} `json:"Endpoints"`
		Storage struct{ TLSPath string } `json:"Storage"`
	}
	if err := json.Unmarshal(contents, &documents); err != nil || len(documents) != 1 {
		return Connection{}, errors.New("docker context inspect returned an invalid document")
	}
	document := documents[0]
	endpoint, ok := document.Endpoints["docker"]
	if !ok || endpoint.Host == "" {
		return Connection{}, fmt.Errorf("docker context %q has no Docker endpoint", document.Name)
	}
	return Connection{ContextName: document.Name, Host: endpoint.Host,
		SkipTLSVerify: endpoint.SkipTLSVerify, TLSPath: document.Storage.TLSPath}, nil
}

// CLIPrefix pins CLI commands to the same context as the SDK connection.
func (c Connection) CLIPrefix() []string {
	if c.ContextName == "" {
		return nil
	}
	return []string{"--context", c.ContextName}
}

// NewClient preserves the context's verified TLS configuration.
func (c Connection) NewClient() (*client.Client, error) {
	if c.FromEnv {
		return client.New(client.FromEnv, client.WithUserAgent("envarium"))
	}
	if strings.HasPrefix(c.Host, "ssh://") {
		return nil, fmt.Errorf("docker SSH context %q is not supported by the Engine SDK observer", c.ContextName)
	}
	options := []client.Opt{client.WithHost(c.Host), client.WithUserAgent("envarium")}
	if strings.HasPrefix(c.Host, "tcp://") {
		if c.SkipTLSVerify {
			return nil, fmt.Errorf("docker context %q disables TLS verification and is refused by Envarium", c.ContextName)
		}
		ca, cert, key := contextTLSFiles(c.TLSPath)
		options = append(options, client.WithTLSClientConfig(ca, cert, key))
	}
	return client.New(options...)
}

func contextTLSFiles(root string) (string, string, string) {
	if root == "" {
		return "", "", ""
	}
	directory := filepath.Join(root, "docker")
	files := []string{filepath.Join(directory, "ca.pem"), filepath.Join(directory, "cert.pem"), filepath.Join(directory, "key.pem")}
	for index, path := range files {
		if _, err := os.Stat(path); err != nil {
			files[index] = ""
		}
	}
	return files[0], files[1], files[2]
}
