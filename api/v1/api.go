package apiv1

import (
	"context"
)

// API is the interface implemented by both client and server implementations
// of DENNIS.
type API interface {
	// CreateQuery instructs DENNIS to begin querying the upstream DNS
	// resolvers for the requested DNS record type and name.
	CreateQuery(ctx context.Context, req *CreateQueryRequest) (*CreateQueryResponse, error)

	// GetQuery retrieves a previously requested Query by it's unique ID. If it
	// does not exist, either because it never did or because it's been
	// removed, the `NotFound` error code will be returned.
	GetQuery(ctx context.Context, req *GetQueryRequest) (*GetQueryResponse, error)
}
