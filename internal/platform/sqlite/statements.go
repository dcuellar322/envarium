package sqlite

import (
	"context"
	"database/sql"
)

// Reuse parsing/planning for hot adapter-owned statements. Values and results are
// never retained. sql.DB owns statement lifetime and closes them with the DB.
func (d *Database) preparedStatement(ctx context.Context, query string) (*sql.Stmt, error) {
	d.statementMu.Lock()
	defer d.statementMu.Unlock()
	if statement := d.statements[query]; statement != nil {
		return statement, nil
	}
	if len(d.statements) >= 64 || d.statementBytes+len(query) > 64<<10 {
		return nil, nil
	}
	statement, err := d.connection.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	if d.statements == nil {
		d.statements = make(map[string]*sql.Stmt)
	}
	d.statements[query] = statement
	d.statementBytes += len(query)
	return statement, nil
}

func (d *Database) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	statement, err := d.preparedStatement(ctx, query)
	if err != nil {
		return nil, err
	}
	if statement == nil {
		return d.connection.QueryContext(ctx, query, args...)
	}
	return statement.QueryContext(ctx, args...)
}

func (d *Database) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	statement, err := d.preparedStatement(ctx, query)
	if err != nil || statement == nil {
		return d.connection.QueryRowContext(ctx, query, args...)
	}
	return statement.QueryRowContext(ctx, args...)
}

func (d *Database) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	statement, err := d.preparedStatement(ctx, query)
	if err != nil {
		return nil, err
	}
	if statement == nil {
		return d.connection.ExecContext(ctx, query, args...)
	}
	return statement.ExecContext(ctx, args...)
}
