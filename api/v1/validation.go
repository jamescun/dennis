package apiv1

import (
	"regexp"
)

// Validate asserts that all required fields are set, and all set fields are
// valid.
func (c *CreateQueryRequest) Validate() error {
	if c == nil {
		return &Error{Code: ErrorCodeBadRequest, Field: ".", Message: "Request body is required"}
	}

	if c.Type == "" {
		return &Error{Code: ErrorCodeBadRequest, Field: ".type", Message: "Type of record is required"}
	} else if c.Name == "" {
		return &Error{Code: ErrorCodeBadRequest, Field: ".name", Message: "Name of domain is required"}
	}

	if !validRecordType(c.Type) {
		return &Error{Code: ErrorCodeBadRequest, Field: ".type", Message: "Record type is not supported"}
	}

	if len(c.Name) < 4 {
		return &Error{Code: ErrorCodeBadRequest, Field: ".name", Message: "Name of domain must be at least 4 characters"}
	} else if len(c.Name) > 253 {
		// 253 is the upper limit of a DNS packet.
		return &Error{Code: ErrorCodeBadRequest, Field: ".name", Message: "Name of domain cannot be longest than 253 characters"}
	} else if !validRecordName(c.Name) {
		return &Error{Code: ErrorCodeBadRequest, Field: ".name", Message: "Name of domain is invalid"}
	}

	return nil
}

// Validate asserts that all required fields are set, and all set fields are
// valid.
func (g *GetQueryRequest) Validate() error {
	if g == nil {
		return &Error{Code: ErrorCodeBadRequest, Field: ".", Message: "Request body is required"}
	}

	if g.ID == "" {
		return &Error{Code: ErrorCodeBadRequest, Field: ".id", Message: "ID of Query is required"}
	}

	return nil
}

// validRecordType returns true if DNS record type t is a type supported by
// DENNIS.
func validRecordType(t string) bool {
	switch t {
	case "A", "AAAA", "CAA", "CNAME", "DNSKEY", "MX", "NS", "PTR", "SOA", "SRV", "SVCB", "TXT":
		return true

	default:
		return false
	}
}

// hostname is a regex that matches a hostname. The TLD must be between 2 and
// 18 characters in length (not including `xn--` for i18n).
//
// fun fact: longest is 18 characters, `.northwesternmutual`.
var hostname = regexp.MustCompile(`^([a-z0-9\-\.]+)\.((xn\-\-)?[a-z0-9]{1,18})$`)

// validRecordName returns true if DNS record name t is a (roughly) valid
// hostname. It doesn't actually resolve the name itself, just checks if it
// is likely to be accepted by a DNS resolver.
func validRecordName(n string) bool {
	return hostname.MatchString(n)
}
