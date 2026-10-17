package sqlite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPreparedWritesKeepBindingsFreshAfterFailureAndCancellation(t *testing.T) {
	t.Parallel()
	d := newResourceMetricDatabase(t)
	ctx := context.Background()
	if _, err := d.connection.ExecContext(ctx, "CREATE TEMP TABLE plan_values (value INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	const query = "INSERT INTO plan_values (value) VALUES (?)"
	if _, err := d.execContext(ctx, query, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := d.execContext(ctx, query, 1); err == nil {
		t.Fatal("constraint failure hidden")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := d.execContext(canceled, query, 2); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled write = %v", err)
	}
	if _, err := d.execContext(ctx, query, 2); err != nil {
		t.Fatal(err)
	}
	var count, sum int
	if err := d.connection.QueryRowContext(ctx, "SELECT COUNT(*), SUM(value) FROM plan_values").Scan(&count, &sum); err != nil || count != 2 || sum != 3 {
		t.Fatalf("fresh write bindings: count=%d sum=%d err=%v", count, sum, err)
	}
}

func TestPreparedReadsKeepValuesFreshAndCancellationLocal(t *testing.T) {
	t.Parallel()
	d := newResourceMetricDatabase(t)
	ctx := context.Background()
	const query = "SELECT display_name FROM projects WHERE id = ?"
	var name string
	if err := d.queryRowContext(ctx, query, "project-1").Scan(&name); err != nil || name != "Project One" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	if _, err := d.connection.ExecContext(ctx, "UPDATE projects SET display_name = 'Changed' WHERE id = 'project-1'"); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := d.queryRowContext(canceled, query, "project-1").Scan(&name); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel: %v", err)
	}
	if err := d.queryRowContext(ctx, query, "project-1").Scan(&name); err != nil || name != "Changed" {
		t.Fatalf("stale or poisoned read: %q %v", name, err)
	}
	if len(d.statements) != 1 {
		t.Fatalf("retained %d statements", len(d.statements))
	}
}

func TestPreparedReadRetentionAndFailuresAreBounded(t *testing.T) {
	t.Parallel()
	d := newResourceMetricDatabase(t)
	ctx := context.Background()
	if _, err := d.queryContext(ctx, "SELECT missing FROM nonexistent"); err == nil || len(d.statements) != 0 {
		t.Fatal("cached failed prepare")
	}
	for i := range 70 {
		var value int
		if err := d.queryRowContext(ctx, fmt.Sprintf("SELECT ? /* %d */", i), i).Scan(&value); err != nil || value != i {
			t.Fatalf("binding %d: %d %v", i, value, err)
		}
	}
	if len(d.statements) > 64 || d.statementBytes > 64<<10 {
		t.Fatal("unbounded statements")
	}
	statement, err := d.preparedStatement(ctx, "SELECT 1 "+strings.Repeat(" ", 64<<10))
	if err != nil || statement != nil {
		t.Fatal("retained oversized SQL")
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	var value int
	if err := d.queryRowContext(ctx, "SELECT ? /* 0 */", 0).Scan(&value); err == nil {
		t.Fatal("statement survived closed database")
	}
}
