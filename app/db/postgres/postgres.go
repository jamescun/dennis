package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jamescun/dennis/app/config"
	"github.com/jamescun/dennis/app/db"
	"github.com/jamescun/dennis/app/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is a database implementation backed by a PostgreSQL database.
type DB struct {
	// conn is an interface containing just the methods we need from the
	// PostgreSQL connection pool.
	conn interface {
		Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
		Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
		QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	}
}

// New initializes a new DB database implementation backed by PostgreSQL.
func New(ctx context.Context, url string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("could not parse postgres config: %w", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	d := &DB{conn: conn}
	if err := d.migrate(ctx); err != nil {
		return nil, err
	}

	return d, nil
}

// FromConfig configures a PostgreSQL database implementation from a configuration
// object supplied by the user.
func FromConfig(ctx context.Context, cfg *config.PostgresDB) (*DB, error) {
	return New(ctx, cfg.URL)
}

// migrate creates the database schemas expected by DENNIS and applies any
// required migrations to bring the PostgreSQL database up to what the current
// configuration expected by DENNIS.
//
// TODO(jc): implement migration logic, stop using 'IF NOT EXISTS' condition in
// schema creation.
func (d *DB) migrate(ctx context.Context) error {
	if _, err := d.conn.Exec(ctx, migrationTable); err != nil {
		return fmt.Errorf("could not create `schema_migrations` table: %w", err)
	}

	if _, err := d.conn.Exec(ctx, queryTable); err != nil {
		return fmt.Errorf("could not create `queries` table: %w", err)
	}

	if _, err := d.conn.Exec(ctx, lookupTable); err != nil {
		return fmt.Errorf("could not create `lookups` table: %w", err)
	}

	if _, err := d.conn.Exec(ctx, recordTable); err != nil {
		return fmt.Errorf("could not create `records` table: %w", err)
	}

	return nil
}

func (d *DB) CreateQuery(ctx context.Context, q *models.Query) error {
	const query = `
		INSERT INTO queries (type, name) VALUES ($1, $2)
		RETURNING id, created_at
	`

	err := d.conn.QueryRow(ctx, query, q.Type, q.Name).Scan(&q.ID, &q.CreatedAt)
	if err != nil {
		return fmt.Errorf("could not create query: %w", err)
	}

	return nil
}

func (d *DB) GetQueryByID(ctx context.Context, id uuid.UUID) (*models.Query, error) {
	query, err := d.getQueryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query.Lookups, err = d.listLookupsForQueryID(ctx, query.ID)
	if err != nil {
		return nil, err
	}

	for _, lookup := range query.Lookups {
		lookup.Records, err = d.listRecordsForLookupID(ctx, *lookup.ID)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

func (d *DB) getQueryByID(ctx context.Context, id uuid.UUID) (*models.Query, error) {
	const query = `
		SELECT id, type, name, created_at, finished_at
		FROM queries
		WHERE id = $1
	`

	q := new(models.Query)

	err := d.conn.QueryRow(ctx, query, id).Scan(
		&q.ID, &q.Type, &q.Name, &q.CreatedAt, &q.FinishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, db.ErrQueryNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get query: %w", err)
	}

	return q, nil
}

func (d *DB) listLookupsForQueryID(ctx context.Context, queryID uuid.UUID) ([]*models.Lookup, error) {
	const query = `
		SELECT id, resolver, rtt, error, resolved_at
		FROM lookups
		WHERE query_id = $1
	`

	lks := []*models.Lookup{}

	rows, err := d.conn.Query(ctx, query, queryID)
	if err != nil {
		return nil, fmt.Errorf("could not query lookups: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		lk := new(models.Lookup)
		err := rows.Scan(&lk.ID, &lk.Resolver, &lk.RTT, &lk.Error, &lk.ResolvedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan lookup: %w", err)
		}

		lks = append(lks, lk)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("could not scan lookups: %w", err)
	}

	return lks, nil
}

func (d *DB) listRecordsForLookupID(ctx context.Context, lookupID uuid.UUID) ([]*models.Record, error) {
	const query = `
		SELECT ttl, priority, weight, port, tag, content
		FROM records
		WHERE lookup_id = $1
	`

	recs := []*models.Record{}

	rows, err := d.conn.Query(ctx, query, lookupID)
	if err != nil {
		return nil, fmt.Errorf("could not query records: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		rec := &models.Record{}
		err := rows.Scan(&rec.TTL, &rec.Priority, &rec.Weight, &rec.Port, &rec.Tag, &rec.Content)
		if err != nil {
			return nil, fmt.Errorf("could not scan record: %w", err)
		}

		recs = append(recs, rec)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("could not scan records: %w", err)
	}

	return recs, nil
}

func (d *DB) UpdateQuery(ctx context.Context, q *models.Query) error {
	const query = `
		UPDATE queries
		SET finished_at = $1
		WHERE id = $2
	`

	result, err := d.conn.Exec(ctx, query, q.FinishedAt, q.ID)
	if err != nil {
		return fmt.Errorf("could not update query: %w", err)
	} else if rowsAffected := result.RowsAffected(); rowsAffected != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rowsAffected)
	}

	return nil
}

func (d *DB) DeleteQueriesOlderThan(ctx context.Context, maxAge time.Duration) error {
	return nil
}

func (d *DB) CreateLookup(ctx context.Context, queryID uuid.UUID, lk *models.Lookup) error {
	err := d.createLookup(ctx, queryID, lk)
	if err != nil {
		return err
	}

	for _, rec := range lk.Records {
		err := d.createRecord(ctx, *lk.ID, rec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) createLookup(ctx context.Context, queryID uuid.UUID, lk *models.Lookup) error {
	const query = `
		INSERT INTO lookups (query_id, resolver, rtt, error, resolved_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`

	err := d.conn.QueryRow(
		ctx, query,
		queryID, lk.Resolver, lk.RTT, lk.Error, lk.ResolvedAt,
	).Scan(&lk.ID)
	if err != nil {
		return fmt.Errorf("could not create lookup: %w", err)
	}

	return nil
}

func (d *DB) createRecord(ctx context.Context, lookupID uuid.UUID, rec *models.Record) error {
	const query = `
		INSERT INTO records (lookup_id, ttl, priority, weight, port, tag, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := d.conn.Exec(
		ctx, query,
		lookupID, rec.TTL, rec.Priority, rec.Weight, rec.Port, rec.Tag, rec.Content,
	)
	if err != nil {
		return fmt.Errorf("could not create record: %w", err)
	}

	return nil
}
