package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"envarium.dev/envarium/internal/manifest/domain"
)

func TestResolverExactSourcesRemainFreshAndIndependent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, ".envarium")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	resolver := NewResolver()
	accepted := domain.Manifest{SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject, Metadata: domain.Metadata{ID: "fixture", Name: "Accepted"}, Repository: domain.Repository{Root: "."}}
	compare := func(override []byte) EffectiveManifest {
		t.Helper()
		want, err := Resolve(root, domain.Manifest{}, accepted, override)
		if err != nil {
			t.Fatal(err)
		}
		got, err := resolver.Resolve(root, domain.Manifest{}, accepted, override)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatal("memoized resolution differs from fresh resolution")
		}
		return got
	}
	first := compare(nil)
	encoded, err := json.Marshal(first.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(encoded)
	if first.ContentHash != hex.EncodeToString(digest[:]) {
		t.Fatal("declaration identity differs from canonical JSON")
	}
	first.Manifest.Metadata.Name = "mutated"
	first.Provenance["/metadata/name"] = "mutated"
	first.Sources[0].Name = "mutated"
	if compare(nil).Manifest.Metadata.Name != "Accepted" {
		t.Fatal("caller mutation leaked")
	}
	path := filepath.Join(directory, "project.local.yml")
	write := func(v string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(v), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("metadata: {name: First}")
	if compare(nil).Manifest.Metadata.Name != "First" {
		t.Fatal("new file ignored")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	write("metadata: {name: Other}")
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if compare(nil).Manifest.Metadata.Name != "Other" {
		t.Fatal("same-time edit ignored")
	}
	if compare([]byte("metadata: {name: Override}")).Manifest.Metadata.Name != "Override" {
		t.Fatal("runtime override ignored")
	}
	write("metadata: {unknownField: invalid}")
	if _, err := resolver.Resolve(root, domain.Manifest{}, accepted, nil); err == nil {
		t.Fatal("invalid edit reused prior success")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	accepted.Metadata.Name = "Reaccepted"
	if compare(nil).Manifest.Metadata.Name != "Reaccepted" {
		t.Fatal("accepted revision ignored")
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(root, domain.Manifest{}, accepted, nil); err == nil {
		t.Fatal("unreadable source reused prior success")
	}
}

func TestResolverConcurrentReadersAndMemoryBounds(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	resolver := NewResolver()
	accepted := domain.Manifest{SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject, Metadata: domain.Metadata{ID: "fixture", Name: "Fixture"}, Repository: domain.Repository{Root: "."}}
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			for range 10 {
				got, err := resolver.Resolve(root, domain.Manifest{}, accepted, nil)
				if err != nil || got.Manifest.Metadata.Name != "Fixture" {
					t.Errorf("resolution: %v", err)
				}
			}
		})
	}
	workers.Wait()
	for i := range maximumResolutions * 2 {
		accepted.Metadata.Name = fmt.Sprintf("Fixture-%d", i)
		if _, err := resolver.Resolve(root, domain.Manifest{}, accepted, nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(resolver.entries) > maximumResolutions || resolver.bytes > maximumResolutionBytes {
		t.Fatal("unbounded memoization")
	}
}

func BenchmarkEffectiveManifestResolution(b *testing.B) {
	root := b.TempDir()
	accepted := domain.Manifest{SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject, Metadata: domain.Metadata{ID: "fixture", Name: "Fixture"}, Repository: domain.Repository{Root: "."}}
	for _, cached := range []bool{false, true} {
		name := "fresh"
		if cached {
			name = "memoized"
		}
		b.Run(name, func(b *testing.B) {
			resolver := NewResolver()
			b.ReportAllocs()
			for b.Loop() {
				var err error
				if cached {
					_, err = resolver.Resolve(root, domain.Manifest{}, accepted, nil)
				} else {
					_, err = Resolve(root, domain.Manifest{}, accepted, nil)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestAcceptedRevisionAvoidsSnapshotReadsButNotLocalSourceReads(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	resolver := NewResolver()
	calls := 0
	accepted := domain.Manifest{SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject, Metadata: domain.Metadata{ID: "fixture", Name: "First"}, Repository: domain.Repository{Root: "."}}
	load := func() (domain.Manifest, error) { calls++; return accepted, nil }
	for range 10 {
		if _, err := resolver.ResolveRevision(root, "project:1", load, nil); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Fatalf("snapshot reads: %d", calls)
	}
	accepted.Metadata.Name = "Second"
	got, err := resolver.ResolveRevision(root, "project:2", load, nil)
	if err != nil || got.Manifest.Metadata.Name != "Second" || calls != 2 {
		t.Fatalf("revision: %v %v %d", got, err, calls)
	}
	directory := filepath.Join(root, ".envarium")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "project.local.yml")
	if err := os.WriteFile(path, []byte("metadata: {name: Local}"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = resolver.ResolveRevision(root, "project:2", load, nil)
	if err != nil || got.Manifest.Metadata.Name != "Local" {
		t.Fatalf("local edit: %v %v", got, err)
	}
	if err := os.WriteFile(path, []byte("invalid: ["), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveRevision(root, "project:2", load, nil); err == nil {
		t.Fatal("invalid local source reused revision memo")
	}
}

func TestRevisionIdentityChecksCurrentSourcesWithoutExposingCachedManifest(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".envarium"), 0700); err != nil {
		t.Fatal(err)
	}
	accepted := domain.Manifest{SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject, Metadata: domain.Metadata{ID: "one", Name: "One"}, Repository: domain.Repository{Root: "."}}
	resolver := NewResolver()
	loads := 0
	load := func() (domain.Manifest, error) { loads++; return accepted, nil }
	hash, err := resolver.RevisionContentHash(root, "one:1", load, nil)
	if err != nil || hash == "" {
		t.Fatalf("hash=%q err=%v", hash, err)
	}
	again, err := resolver.RevisionContentHash(root, "one:1", load, nil)
	if err != nil || again != hash || loads != 1 {
		t.Fatal("revision identity was not reused")
	}
	file := filepath.Join(root, ".envarium", "project.local.yml")
	if err := os.WriteFile(file, []byte("metadata: {name: Changed}"), 0600); err != nil {
		t.Fatal(err)
	}
	changed, err := resolver.RevisionContentHash(root, "one:1", load, nil)
	if err != nil || changed == hash {
		t.Fatal("source edit hidden")
	}
	if err := os.WriteFile(file, []byte("invalid: ["), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.RevisionContentHash(root, "one:1", load, nil); err == nil {
		t.Fatal("invalid source reused old identity")
	}
}
