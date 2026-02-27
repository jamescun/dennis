package db

import (
	"context"
	"errors"

	"github.com/jamescun/dennis/app/models"

	"github.com/gofrs/uuid"
)

// ErrQueryNotFound is returned by a database implementation when attempting to
// retrieve a Query by ID, but it does not exist. It may have existed at one
// point, but expired out of the database.
var ErrQueryNotFound = errors.New("query not found")

// DB is composed of the database object interfaces in this package.
type DB interface {
	Queries
	Lookups
}

// Queries is used to operate on Query objects in the database.
type Queries interface {
	// CreateQuery inserts a new Query into the database. The ID and CreatedAt
	// fields will be set by the database.
	CreateQuery(ctx context.Context, query *models.Query) error

	// GetQueryByID retrieves a Query by it's ID from the database. If it does
	// not exist, ErrQueryNotFound is returned.
	GetQueryByID(ctx context.Context, id uuid.UUID) (*models.Query, error)

	// UpdateQuery updates a Query in the database. Currently only FinishedAt
	// is updatable. If it does not exist, ErrQueryNotFound is returned.
	UpdateQuery(ctx context.Context, query *models.Query) error
}

// Lookups is used to operate on Lookup objects that live under Query objects
// in the database.
type Lookups interface {
	// CreateLookup inserts a Lookup into the database to be associated with a
	// Query. If a Query of queryID does not exist, ErrQueryNotFound is
	// returned.
	CreateLookup(ctx context.Context, queryID uuid.UUID, l *models.Lookup) error
}
