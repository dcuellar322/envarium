package application

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"sync"

	"envarium.dev/envarium/internal/manifest/domain"
)

const maximumResolutionBytes = 4 << 20
const maximumResolutions = 128

// Resolver memoizes exact input documents, with no freshness window. Source
// reads and errors precede every lookup. Typed copies isolate caller mutation
// and conservative retention weights bound the cache; neither errors nor oversized results are retained.
type Resolver struct {
	mu      sync.Mutex
	entries map[[32]byte]EffectiveManifest
	bytes   int
}

// NewResolver creates a bounded resolver owned by its application service.
func NewResolver() *Resolver { return &Resolver{entries: make(map[[32]byte]EffectiveManifest)} }

// Resolve uses the same precedence and validation as the uncached Resolve.
func (r *Resolver) Resolve(root string, discovered, accepted domain.Manifest, override []byte) (EffectiveManifest, error) {
	documents, err := resolutionDocuments(root, discovered, accepted, override)
	if err != nil {
		return EffectiveManifest{}, err
	}
	key := resolutionKey(root, "contents", documents)
	if result, ok, err := r.cachedResolution(key); ok || err != nil {
		return result, err
	}
	return r.resolveAndStore(root, key, documents)
}

func resolutionKey(root, scope string, documents []resolutionDocument) [32]byte {
	hash := sha256.New()
	for _, part := range append([]resolutionDocument{{contents: []byte(root)}, {contents: []byte(scope)}}, documents...) {
		var size [8]byte
		binary.LittleEndian.PutUint64(size[:], uint64(len(part.contents)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write(part.contents)
	}
	var key [32]byte
	copy(key[:], hash.Sum(nil))
	return key
}

func (r *Resolver) cachedResolution(key [32]byte) (EffectiveManifest, bool, error) {
	r.mu.Lock()
	result, ok := r.entries[key]
	r.mu.Unlock()
	if !ok {
		return EffectiveManifest{}, false, nil
	}
	return cloneResolution(result), true, nil
}

func (r *Resolver) resolveAndStore(root string, key [32]byte, documents []resolutionDocument) (EffectiveManifest, error) {
	result, err := resolveDocuments(root, documents)
	if err != nil {
		return EffectiveManifest{}, err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return EffectiveManifest{}, err
	}
	// Account generously for map/slice/struct overhead as well as string payloads.
	// Encoded data is used only for weighing and is not retained.
	weight := 1024 + 8*len(encoded)
	if weight <= maximumResolutionBytes {
		r.mu.Lock()
		if _, exists := r.entries[key]; !exists {
			if len(r.entries) >= maximumResolutions || r.bytes+weight > maximumResolutionBytes {
				clear(r.entries)
				r.bytes = 0
			}
			r.entries[key] = cloneResolution(result)
			r.bytes += weight
		}
		r.mu.Unlock()
	}
	return result, nil
}
