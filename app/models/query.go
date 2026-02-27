package models

import (
	"time"

	"github.com/gofrs/uuid"
)

// Query is a user request to lookup a name of a specified DNS record type
// against each configured DNS resolver.
type Query struct {
	// ID is the unique identifier for this Query. Depending on the storage
	// backend configured, this may be a long-lived identifier.
	ID uuid.UUID `json:"id"`

	// Type is the type of DNS record that is to be resolved using each
	// configured DNS resolver.
	Type string `json:"type"`

	// Name is the domain name to resolve against each configured DNS resolver.
	Name string `json:"name"`

	// Lookups are the queries and records returned by the configured DNS
	// resolvers.
	Lookups []*Lookup `json:"lookups"`

	// CreatedAt is the UTC timestamp indicating when this Query was requested
	// by a user.
	CreatedAt time.Time `json:"createdAt"`

	// FinishedAt is the UTC timestamp indicating when this Query completed
	// resolving against each configured DNS resolver, or nil if the Query is
	// still running.
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
}
