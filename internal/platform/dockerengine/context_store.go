package dockerengine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadConnection reads Docker's local context metadata without launching the CLI.
// Selection follows explicit context, DOCKER_CONTEXT, DOCKER_HOST, config, default.
// It never loads registry credentials into a model or invokes credential helpers.
func ReadConnection(ctx context.Context, requested, workingDirectory string) (Connection, error) {
	if err := ctx.Err(); err != nil {
		return Connection{}, err
	}
	name := requested
	if name == "" {
		name = os.Getenv("DOCKER_CONTEXT")
	}
	if name == "" && os.Getenv("DOCKER_HOST") != "" {
		return Connection{FromEnv: true}, nil
	}
	directory := os.Getenv("DOCKER_CONFIG")
	if directory == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Connection{}, err
		}
		directory = filepath.Join(home, ".docker")
	} else if !filepath.IsAbs(directory) && workingDirectory != "" {
		directory = filepath.Join(workingDirectory, directory)
	}
	if name == "" {
		contents, err := readContextFile(ctx, filepath.Join(directory, "config.json"))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Connection{}, err
		}
		if err == nil {
			var selection struct {
				CurrentContext string `json:"currentContext"`
			}
			if err := json.Unmarshal(contents, &selection); err != nil {
				return Connection{}, errors.New("invalid Docker context selection")
			}
			name = selection.CurrentContext
		}
	}
	if name == "" || name == "default" {
		return Connection{ContextName: "default", FromEnv: true}, nil
	}
	digest := sha256.Sum256([]byte(name))
	id := hex.EncodeToString(digest[:])
	contents, err := readContextFile(ctx, filepath.Join(directory, "contexts", "meta", id, "meta.json"))
	if err != nil {
		return Connection{}, fmt.Errorf("read Docker context %q: %w", name, err)
	}
	connection, err := DecodeContext(append(append([]byte{'['}, contents...), ']'))
	if err != nil {
		return Connection{}, err
	}
	if connection.ContextName != name {
		return Connection{}, errors.New("docker context metadata name mismatch")
	}
	connection.TLSPath = filepath.Join(directory, "contexts", "tls", id)
	return connection, nil
}

func readContextFile(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	contents, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(contents) > 1<<20 {
		return nil, errors.New("docker context metadata exceeds 1 MiB")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return contents, nil
}
