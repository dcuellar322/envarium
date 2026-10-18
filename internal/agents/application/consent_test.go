package application

import (
	"context"
	"errors"
	"testing"

	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
)

func TestEnhancementRejectsChangedEvidenceAndLimitsBeforeProvider(t *testing.T) {
	for _, change := range []string{"missing consent", "candidate", "excerpt", "limits", "provider"} {
		t.Run(change, func(t *testing.T) {
			baseline := baselineProposal()
			catalog := &proposalCatalogFake{proposal: baseline, project: catalogDomain.Project{ID: baseline.ProjectID, PrimaryLocation: t.TempDir()}}
			provider := &providerFake{}
			other := &providerFake{id: "other"}
			registry, _ := NewRegistry(provider, other)
			service, _ := NewEnhancementService(catalog, &runRepositoryFake{}, readerFake{excerpt: "reviewed"}, redactorFake{}, domainValidator{}, registry)
			preview, err := service.Preview(context.Background(), baseline.ID, "fixture", Limits{})
			if err != nil {
				t.Fatal(err)
			}
			limits := Limits{}
			providerID := "fixture"
			switch change {
			case "missing consent":
				preview.ConsentSHA256 = ""
			case "candidate":
				catalog.proposal.Candidate.Metadata.Name = "changed"
			case "excerpt":
				service.reader = readerFake{excerpt: "unreviewed"}
			case "limits":
				limits.OutputBytes = 8192
			case "provider":
				providerID = "other"
			}
			err = service.Execute(context.Background(), "op", baseline.ID, providerID, limits, preview.ConsentSHA256)
			if !errors.Is(err, ErrEvidenceChanged) {
				t.Fatalf("error=%v", err)
			}
			if provider.calls != 0 || other.calls != 0 {
				t.Fatal("provider called without matching consent")
			}
		})
	}
}
