package compose

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"go.yaml.in/yaml/v4"
)

const maximumObservationInputs = 256
const maximumObservationInputBytes = 4 << 20

type observationInput struct {
	path    string
	digest  [32]byte
	present bool
}

type metadataMemoEntry struct {
	inputs []observationInput
	config normalizedConfig
}

// metadataMemo retains validated names and membership, never source contents.
// Inputs are read by content, not mtime. Unknown dependency graphs bypass it.
type metadataMemo struct {
	mu      sync.Mutex
	entries map[[32]byte]metadataMemoEntry
}

func newMetadataMemo() *metadataMemo {
	return &metadataMemo{entries: make(map[[32]byte]metadataMemoEntry)}
}

func metadataKey(files []string, root, name string) [32]byte {
	environment := os.Environ()
	slices.Sort(environment)
	encoded, _ := json.Marshal([]any{files, root, name, environment})
	return sha256.Sum256(encoded)
}

func (m *metadataMemo) lookup(ctx context.Context, key [32]byte) (normalizedConfig, bool) {
	if m == nil {
		return normalizedConfig{}, false
	}
	m.mu.Lock()
	entry, ok := m.entries[key]
	m.mu.Unlock()
	if !ok || !unchangedInputs(ctx, entry.inputs) {
		return normalizedConfig{}, false
	}
	return cloneConfig(entry.config), true
}

func (m *metadataMemo) store(key [32]byte, inputs []observationInput, config normalizedConfig) {
	if m == nil || len(inputs) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) >= maximumObservationConfigs {
		clear(m.entries)
	}
	m.entries[key] = metadataMemoEntry{inputs: inputs, config: cloneConfig(config)}
}

func unchangedInputs(ctx context.Context, inputs []observationInput) bool {
	for _, input := range inputs {
		current, _, ok := readObservationInput(ctx, input.path)
		if !ok || current != input {
			return false
		}
	}
	return ctx.Err() == nil
}

func readObservationInput(ctx context.Context, path string) (observationInput, []byte, bool) {
	input := observationInput{path: path}
	if ctx.Err() != nil {
		return input, nil, false
	}
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return input, nil, true
	}
	if err != nil {
		return input, nil, false
	}
	defer func() { _ = file.Close() }()
	value, err := io.ReadAll(io.LimitReader(file, maximumObservationInputBytes+1))
	if err != nil || len(value) > maximumObservationInputBytes || ctx.Err() != nil {
		return input, nil, false
	}
	input.present, input.digest = true, sha256.Sum256(value)
	return input, value, true
}

// Known dependency-free/local-file graphs can avoid full reference-loader
// parsing after a content check. Any uncertainty preserves the existing loader.
func observationInputs(ctx context.Context, files []string, root string) []observationInput {
	if os.Getenv("COMPOSE_ENV_FILES") != "" || os.Getenv("COMPOSE_DISABLE_ENV_FILE") != "" {
		return nil
	}
	paths := append(slices.Clone(files), filepath.Join(root, ".env"))
	inputs := make([]observationInput, 0, len(paths))
	seen := make(map[string]bool)
	for index := 0; index < len(paths); index++ {
		path := filepath.Clean(paths[index])
		if seen[path] {
			continue
		}
		if len(seen) >= maximumObservationInputs {
			return nil
		}
		seen[path] = true
		input, contents, ok := readObservationInput(ctx, path)
		if !ok {
			return nil
		}
		inputs = append(inputs, input)
		if index >= len(files) {
			continue
		}
		var node yaml.Node
		if err := yaml.Unmarshal(contents, &node); err != nil {
			return nil
		}
		dependencies, ok := localDependencies(&node)
		if !ok {
			return nil
		}
		for _, dependency := range dependencies {
			if !filepath.IsAbs(dependency) {
				dependency = filepath.Join(root, dependency)
			}
			paths = append(paths, dependency)
		}
	}
	return inputs
}

func localDependencies(node *yaml.Node) ([]string, bool) {
	if node.Kind == yaml.AliasNode {
		return nil, false
	}
	var result []string
	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			key, value := node.Content[index], node.Content[index+1]
			switch key.Value {
			case "include", "extends":
				return nil, false
			case "env_file", "label_file":
				paths, ok := literalDependencyPaths(value)
				if !ok {
					return nil, false
				}
				result = append(result, paths...)
			default:
				paths, ok := localDependencies(value)
				if !ok {
					return nil, false
				}
				result = append(result, paths...)
			}
		}
	} else {
		for _, child := range node.Content {
			paths, ok := localDependencies(child)
			if !ok {
				return nil, false
			}
			result = append(result, paths...)
		}
	}
	return result, true
}

func literalDependencyPaths(node *yaml.Node) ([]string, bool) {
	if node.Kind == yaml.SequenceNode {
		var paths []string
		for _, child := range node.Content {
			values, ok := literalDependencyPaths(child)
			if !ok {
				return nil, false
			}
			paths = append(paths, values...)
		}
		return paths, true
	}
	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			if node.Content[index].Value == "path" {
				return literalDependencyPaths(node.Content[index+1])
			}
		}
		return nil, false
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!str" || node.Value == "" || strings.ContainsAny(node.Value, "$~") {
		return nil, false
	}
	return []string{node.Value}, true
}
