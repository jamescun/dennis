package config

import (
	"path/filepath"
	"strconv"
)

// ValidationError is an error returned by validation functions attached to
// configuration objects when they are not validate.
type ValidationError struct {
	// Field is a JSONPath-compatible description of the field at fault.
	Field string

	// Message is the human-readable description of that is invalid about the
	// fields.
	Message string
}

func (ve *ValidationError) prefix(p string) *ValidationError {
	if ve == nil {
		return nil
	}

	if ve.Field != "" {
		p = p + "." + ve.Field
	}

	return &ValidationError{
		Field:   p,
		Message: ve.Message,
	}
}

func (ve *ValidationError) prefixIdx(p string, i int) *ValidationError {
	if ve == nil {
		return nil
	}

	return &ValidationError{
		Field:   p + "[" + strconv.Itoa(i) + "]." + ve.Field,
		Message: ve.Message,
	}
}

func (ve *ValidationError) Error() string {
	return ve.Field + ": " + ve.Message
}

// Validate asserts the validity of Config, returning a ValidationError for the
// first invalid field encountered.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return &ValidationError{Field: "version", Message: "unsupported config version"}
	}

	if err := c.Listen.validate(); err != nil {
		return err.prefix("listen")
	}

	for i, r := range c.Resolvers {
		if err := r.validate(); err != nil {
			return err.prefixIdx("resolvers", i)
		}
	}

	if err := c.DB.validate(); err != nil {
		return err.prefix("db")
	}

	return nil
}

func (l *Listener) validate() *ValidationError {
	if l == nil {
		return &ValidationError{Message: "listener is required"}
	}

	if l.Addr == "" {
		return &ValidationError{Field: "addr", Message: "addr to listen on is required"}
	}

	return nil
}

func (r *Resolver) validate() *ValidationError {
	if r == nil {
		return &ValidationError{Message: "resolver is required"}
	}

	if r.Name == "" {
		return &ValidationError{Field: "name", Message: "name of resolver is required"}
	}

	if r.Addr == "" {
		return &ValidationError{Field: "addr", Message: "addr of resolver is required"}
	}

	return nil
}

func (d *DB) validate() *ValidationError {
	switch {
	case d.File != nil:
		if d.Redis != nil {
			return &ValidationError{Field: "file", Message: "only one database can be configured at once"}
		}

		return d.File.validate().prefix("file")

	case d.Redis != nil:
		if d.File != nil {
			return &ValidationError{Field: "redis", Message: "only one database can be configured at once"}
		}

		return d.Redis.validate().prefix("redis")

	default:
		return &ValidationError{Message: "at least file, postgres or redis configuration is required"}
	}
}

func (f *FileDB) validate() *ValidationError {
	if f.Path == "" {
		return &ValidationError{Field: "path", Message: "path to local file is required"}
	}

	if filepath.IsAbs(f.Path) {
		return &ValidationError{Field: "path", Message: "path is not an absolute path"}
	}

	return nil
}

func (r *RedisDB) validate() *ValidationError {
	if r.Addr == "" {
		return &ValidationError{Field: "addr", Message: "redis server address is required"}
	}

	if r.DB < 0 {
		return &ValidationError{Field: "db", Message: "redis database must be zero or greater"}
	}

	return nil
}
