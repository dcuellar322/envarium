package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListProjectsBatchesTagsWithoutLosingIsolationOrOrder(t *testing.T) {
	database, err := Open(context.Background(), filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()
	ctx := context.Background()
	for _, row := range []struct{ id, name string }{{"a", "Zulu"}, {"b", "Alpha"}, {"c", "Empty"}} {
		_, err := database.connection.ExecContext(ctx, `INSERT INTO projects (id,slug,display_name,description,trust_state,primary_location,manifest_revision,created_at,updated_at) VALUES (?,?,?,'','trusted',?,1,?,?)`, row.id, row.id, row.name, "/"+row.id, "2026-10-16T00:00:00Z", "2026-10-16T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, row := range []struct{ id, tag string }{{"a", "two"}, {"b", "other"}, {"a", "one"}} {
		if _, err := database.connection.ExecContext(ctx, `INSERT INTO project_tags (project_id,tag) VALUES (?,?)`, row.id, row.tag); err != nil {
			t.Fatal(err)
		}
	}
	repository := NewCatalogRepository(database)
	projects, err := repository.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 3 || projects[0].ID != "b" || projects[1].ID != "c" || projects[2].ID != "a" {
		t.Fatal("project order changed")
	}
	if !reflect.DeepEqual(projects[0].Tags, []string{"other"}) || !reflect.DeepEqual(projects[1].Tags, []string{}) || !reflect.DeepEqual(projects[2].Tags, []string{"one", "two"}) {
		t.Fatalf("tag assignment: %v", projects)
	}
	for _, want := range projects {
		got, err := repository.GetProject(ctx, want.ID)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("single-read project/tags mismatch: %v %v", got, err)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repository.ListProjects(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
