package web

import (
	"context"

	"github.com/gofrs/uuid"
)

// contextKey is a string type that is used to prevent collisions in the
// context.Context keyspace.
type contextKey string

// GetRequestID retrieves the request's unique identifier (in the form of a
// UUID) from context. If it is not set, uuid.Nil is returned.
func GetRequestID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(contextKey("requestID")).(uuid.UUID); ok {
		return id
	}

	return uuid.Nil
}

// setRequestID sets the request's unique identifier in the context,
// overwriting any value that may have previously been set.
func setRequestID(parent context.Context, id uuid.UUID) context.Context {
	return context.WithValue(parent, contextKey("requestID"), id)
}
