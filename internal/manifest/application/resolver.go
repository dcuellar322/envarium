package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"envarium.dev/envarium/internal/foundation/repositoryfile"
	"envarium.dev/envarium/internal/manifest/domain"
)

const (
	portableManifest = ".envarium/project.yml"
	localManifest    = ".envarium/project.local.yml"
)

// EffectiveManifest includes every effective field's winning source.
type EffectiveManifest struct {
	// ContentHash identifies the resolved declaration; source files are still read
	// on every resolution. It is internal metadata and never part of the API.
	ContentHash string            `json:"-"`
	Manifest    domain.Manifest   `json:"manifest"`
	Provenance  map[string]string `json:"provenance"`
	Sources     []SourceDocument  `json:"sources"`
}

// SourceDocument is one existing precedence layer.
type SourceDocument struct {
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
}

// Resolve merges discovery, accepted inference, portable YAML, local YAML, and runtime overrides.
func Resolve(root string, discovered, accepted domain.Manifest, runtimeOverride []byte) (EffectiveManifest, error) {
	documents, err := resolutionDocuments(root, discovered, accepted, runtimeOverride)
	if err != nil {
		return EffectiveManifest{}, err
	}
	return resolveDocuments(root, documents)
}

type resolutionDocument struct {
	name, path string
	contents   []byte
}

func resolutionDocuments(root string, discovered, accepted domain.Manifest, override []byte) ([]resolutionDocument, error) {
	documents := []resolutionDocument{{name: "discovery"}, {name: "accepted-inference"},
		{name: "portable-manifest", path: filepath.Join(root, portableManifest)},
		{name: "local-overlay", path: filepath.Join(root, localManifest)},
		{name: "runtime-override", contents: override}}
	for index, value := range []domain.Manifest{discovered, accepted} {
		if value.SchemaVersion == "" {
			continue
		}
		contents, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		documents[index].contents = contents
	}
	// An absent parent proves both optional source files are absent at this
	// observation point. Avoid walking the same missing path twice per project.
	if _, err := os.Stat(filepath.Join(root, ".envarium")); os.IsNotExist(err) {
		return documents, nil
	}
	for index := range documents {
		if documents[index].path == "" {
			continue
		}
		relative, err := filepath.Rel(root, documents[index].path)
		if err != nil {
			return nil, err
		}
		contents, err := repositoryfile.Read(root, relative, 1<<20)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read manifest source %s: %w", documents[index].path, err)
		}
		documents[index].contents = contents
	}
	return documents, nil
}

func resolveDocuments(root string, documents []resolutionDocument) (EffectiveManifest, error) {
	merged := make(map[string]any)
	provenance := make(map[string]string)
	var sources []SourceDocument
	for _, source := range documents {
		if len(source.contents) == 0 {
			continue
		}
		value, err := decodeMap(source.contents)
		if err != nil {
			return EffectiveManifest{}, err
		}
		mergeMap(merged, value, "", source.name, provenance)
		sources = append(sources, SourceDocument{Name: source.name, Path: relativePath(root, source.path)})
	}
	manifest, err := mapToManifest(merged)
	if err != nil {
		return EffectiveManifest{}, err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return EffectiveManifest{}, err
	}
	digest := sha256.Sum256(encoded)
	return EffectiveManifest{Manifest: manifest, Provenance: provenance, Sources: sources, ContentHash: hex.EncodeToString(digest[:])}, nil
}

func mergeMap(target, source map[string]any, prefix, sourceName string, provenance map[string]string) {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := prefix + "/" + escapePointer(key)
		sourceChild, sourceIsMap := source[key].(map[string]any)
		targetChild, targetIsMap := target[key].(map[string]any)
		if sourceIsMap {
			if !targetIsMap {
				targetChild = make(map[string]any)
				target[key] = targetChild
			}
			mergeMap(targetChild, sourceChild, path, sourceName, provenance)
			continue
		}
		target[key] = source[key]
		provenance[path] = sourceName
	}
}

func escapePointer(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
}

func relativePath(root, path string) string {
	if path == "" {
		return ""
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return relative
}
