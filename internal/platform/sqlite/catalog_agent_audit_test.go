package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	catalogApplication "envarium.dev/envarium/internal/catalog/application"
	catalogDomain "envarium.dev/envarium/internal/catalog/domain"
	discoveryDomain "envarium.dev/envarium/internal/discovery/domain"
	manifestDomain "envarium.dev/envarium/internal/manifest/domain"
)

func TestCatalogProposalPersistsAgentOrigin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "envarium.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	repository := NewCatalogRepository(database)
	at := time.Now().UTC()
	project := catalogDomain.Project{ID: "project-1", Slug: "project", DisplayName: "Project", TrustState: catalogDomain.TrustPending, PrimaryLocation: "/tmp/project", Tags: []string{}, CreatedAt: at, UpdatedAt: at}
	proposal := discoveryDomain.Proposal{
		ID: "proposal-1", ProjectID: project.ID, ScannerVersion: discoveryDomain.ScannerVersion,
		SchemaVersion: manifestDomain.SchemaVersion, Candidate: manifestDomain.Manifest{}, Evidence: []discoveryDomain.Evidence{},
		ConfidenceByField: map[string]float64{}, Unresolved: []string{}, Validation: discoveryDomain.Validation{Errors: []string{}, Warnings: []string{}},
		Status: discoveryDomain.StatusProposed, CreatedAt: at,
	}
	actor := catalogApplication.MutationActor{Type: "agent", ID: "codex/reviewer"}
	if err := repository.CreateProposal(ctx, project, proposal, actor); err != nil {
		t.Fatal(err)
	}
	var actorType, actorID string
	if err := database.connection.QueryRowContext(ctx, `SELECT actor_type, actor_id FROM audit_events WHERE event_type = 'manifest.proposal.created' AND project_id = ?`, project.ID).Scan(&actorType, &actorID); err != nil {
		t.Fatal(err)
	}
	if actorType != actor.Type || actorID != actor.ID {
		t.Fatalf("audit actor = %s/%s", actorType, actorID)
	}
}
