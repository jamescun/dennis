package models

import (
	"time"
)

// Lookup represents a single Lookup executed against a single DNS resolver as
// part of a Query of multiple DNS resolvers.
type Lookup struct {
	// Resolver is the name DNS resolver used for this Lookup, as configured by
	// `name` in Config.Resolvers.
	Resolver string `json:"resolver"`

	// RTT is the round-trip time taken by DENNIS's resolver to execute the
	// request against the upstream DNS resolver, in milliseconds.
	RTT int `json:"rtt"`

	// Error is the error rcode returned by a DNS resolver if the name could
	// not be resolved.
	Error string `json:"error,omitempty"`

	// Records are the results, if any, returned by a DNS resolver.
	Records []*Record `json:"records"`

	// ResolvedAt is the UTC timestamp indicating when this lookup completed.
	ResolvedAt time.Time `json:"resolvedAt"`
}
