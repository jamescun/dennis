package apiv1

import (
	"net/http"

	"github.com/jamescun/dennis/app/models"
)

// CreateQueryRequest is the arguments given to API when requesting a new set
// of lookups against the configured DNS resolvers.
type CreateQueryRequest struct {
	// Type is the DNS record type to query for.
	//
	// Required.
	// Supported type: A, AAAA, CAA, CNAME, DNSKEY, MX, NS, PTR, SOA, SRV,
	// SVCB and TXT.
	Type string `json:"type"`

	// Name is the domain name to query for.
	//
	// Required.
	Name string `json:"name"`
}

// CreateQueryResponse contains the Query that was created in response to
// CreateQueryRequest.
type CreateQueryResponse struct {
	Query *models.Query `json:"query"`
}

// GetQueryRequest is the arguments given to API when requesting a Query by
// it's ID.
type GetQueryRequest struct {
	// ID is the unique UUID of a previously requested Query.
	ID string `json:"id"`
}

// GetQueryResponse contains the Query that was requested by ID in response to
// GetQueryRequest.
type GetQueryResponse struct {
	Query *models.Query `json:"query"`
}

// the error codes are the values to be contained within Error.Code to
// generically describe what is at fault, Error.Message will be more
// descriptive.
const (
	// ErrorCodeBadRequest is used when a request object contains invalid
	// values. See Error.Field for what value is incorrect.
	ErrorCodeBadRequest = "BadRequest"

	// ErrorCodeNotFound is used when a request references an object, usually
	// by ID, that does not exist (possibly anymore).
	ErrorCodeNotFound = "NotFound"

	// ErrorCodeInternal is used when an unexpected error occurs on the server
	// and the request could not be completed.
	ErrorCodeInternal = "Internal"
)

// Error is returned when something goes wrong, either internally or with the
// request given to DENNIS.
type Error struct {
	// Code is a generic description of the class of error encountered, see
	// above for known values.
	Code string `json:"code"`

	// Field is optionally set to describe what request argument triggered
	// this error. This is generally formatted as a JSONPath value.
	Field string `json:"field,omitempty"`

	// Message is a human-readable description of the error encountered.
	Message string `json:"message"`
}

// StatusCode controls the HTTP Status Code returned with this Error. If Code
// is unknown, HTTP 500 Internal Server Error will be used.
func (e *Error) StatusCode() int {
	switch e.Code {
	case ErrorCodeBadRequest:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}

func (e *Error) Error() string {
	if e.Field != "" {
		return e.Code + ": " + e.Field + ": " + e.Message + "."
	}
	return e.Code + ": " + e.Message + "."
}

// ErrorWrapper wraps an Error into an `error` key within JSON responses.
type ErrorWrapper struct {
	*Error `json:"error"`
}
