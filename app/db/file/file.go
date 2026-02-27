package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jamescun/dennis/app/config"
	"github.com/jamescun/dennis/app/db"
	"github.com/jamescun/dennis/app/models"

	"github.com/gofrs/uuid"
)

// format is the layout of the local JSON file.
type format struct {
	// Version is the revision of this format contained within the file.
	Version int `json:"version"`

	// Queries are the queries requested by the user and their results.
	Queries []*models.Query `json:"queries"`
}

// getQuery iterates the Queries in format, returning the first that matches
// the given ID, or nil if it does not exist.
func (f *format) getQuery(id uuid.UUID) *models.Query {
	for _, q := range f.Queries {
		if q.ID == id {
			return q
		}
	}

	return nil
}

// DB is a database implementation backed by a local JSON file. Internal
// locking is implemented between calls, so concurrent use is supported.
type DB struct {
	path string
	mu   sync.Mutex
}

// New initializes a new DB implementation backed by a local JSON file. If the
// file does not exist, it will be created.
func New(path string) (*DB, error) {
	d := &DB{path: path}
	err := d.init()
	if err != nil {
		return nil, err
	}

	return d, nil
}

// FromConfig configures a File database implementation from a configuration
// object supplied by the user.
func FromConfig(_ context.Context, cfg *config.FileDB) (*DB, error) {
	return New(cfg.Path)
}

func (d *DB) init() error {
	// if the file does not exist, write an empty one.
	if _, err := os.Stat(d.path); err != nil {
		err = writeJSON(d.path, &format{Version: 1, Queries: []*models.Query{}})
		if err != nil {
			return fmt.Errorf("init: %w", err)
		}
	}

	return nil
}

func (d *DB) CreateQuery(_ context.Context, query *models.Query) error {
	query.ID = uuid.Must(uuid.NewV7())
	query.CreatedAt = time.Now().UTC()

	err := d.write(func(f *format) error {
		f.Queries = append(f.Queries, query)
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not create query: %w", err)
	}

	return nil
}

func (d *DB) GetQueryByID(_ context.Context, id uuid.UUID) (q *models.Query, err error) {
	err = d.read(func(f *format) error {
		q = f.getQuery(id)
		if q == nil {
			return db.ErrQueryNotFound
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("could not get query: %w", err)
	}

	return
}

func (d *DB) UpdateQuery(_ context.Context, query *models.Query) error {
	err := d.write(func(f *format) error {
		q := f.getQuery(query.ID)
		if q == nil {
			return db.ErrQueryNotFound
		}

		q.FinishedAt = query.FinishedAt
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not update query: %w", err)
	}

	return nil
}

func (d *DB) CreateLookup(_ context.Context, queryID uuid.UUID, l *models.Lookup) error {
	err := d.write(func(f *format) error {
		q := f.getQuery(queryID)
		if q == nil {
			return db.ErrQueryNotFound
		}

		q.Lookups = append(q.Lookups, l)
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not create lookup: %w", err)
	}

	return nil
}

func (d *DB) read(fn func(*format) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := readJSON(d.path)
	if err != nil {
		return err
	}

	err = fn(f)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) write(fn func(*format) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := readJSON(d.path)
	if err != nil {
		return err
	}

	err = fn(f)
	if err != nil {
		return err
	}

	err = writeJSON(d.path, f)
	if err != nil {
		return err
	}

	return nil
}

func readJSON(path string) (*format, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	f := new(format)

	err = json.NewDecoder(file).Decode(f)
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	return f, nil
}

func writeJSON(path string, f *format) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(f)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	return nil
}
