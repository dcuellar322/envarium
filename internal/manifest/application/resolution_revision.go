package application

import (
	"encoding/json"

	"envarium.dev/envarium/internal/manifest/domain"
)

// ResolveRevision accepts an authoritative immutable accepted-snapshot revision.
// Current local source files are still read before every lookup. Callers must
// read the current catalog revision first and use a different key after edits.
func (r *Resolver) ResolveRevision(root, revision string, load func() (domain.Manifest, error), override []byte) (EffectiveManifest, error) {
	return r.resolveRevision(root, revision, load, override, true)
}

// RevisionContentHash reads current source bytes and accepted revision before
// returning identity alone. It never exposes the retained mutable declaration.
func (r *Resolver) RevisionContentHash(root, revision string, load func() (domain.Manifest, error), override []byte) (string, error) {
	result, err := r.resolveRevision(root, revision, load, override, false)
	return result.ContentHash, err
}

func (r *Resolver) resolveRevision(root, revision string, load func() (domain.Manifest, error), override []byte, copyResult bool) (EffectiveManifest, error) {
	documents, err := resolutionDocuments(root, domain.Manifest{}, domain.Manifest{}, override)
	if err != nil {
		return EffectiveManifest{}, err
	}
	key := resolutionKey(root, "revision:"+revision, documents)
	r.mu.Lock()
	result, ok := r.entries[key]
	r.mu.Unlock()
	if ok {
		if copyResult {
			result = cloneResolution(result)
		}
		return result, nil
	}
	accepted, err := load()
	if err != nil {
		return EffectiveManifest{}, err
	}
	if accepted.SchemaVersion != "" {
		documents[1].contents, err = json.Marshal(accepted)
		if err != nil {
			return EffectiveManifest{}, err
		}
	}
	return r.resolveAndStore(root, key, documents)
}
