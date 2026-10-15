package application

import (
	"testing"

	"envarium.dev/envarium/internal/manifest/domain"
)

func BenchmarkRepeatedManifestValidation(b *testing.B) {
	manifest := domain.Manifest{
		SchemaVersion: domain.SchemaVersion, Kind: domain.KindProject,
		Metadata: domain.Metadata{ID: "example", Name: "Example"}, Repository: domain.Repository{Root: "."},
	}
	if err := validateSchema(manifest); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := validateSchema(manifest); err != nil {
			b.Fatal(err)
		}
	}
}
