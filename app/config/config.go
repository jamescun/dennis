package config

import (
	"log/slog"
	"os"
)

// Config is the structure of the configuration file, JSON or YAML, given to
// DENNIS when it starts; configuring logging, resolvers, storage etc.
type Config struct {
	// Version is the revision of this configuration structure implemented
	// within a file. If not set, version `1` is assumed.
	Version int `json:"version"`

	// Logging configure DENNIS's logs.
	Logging Logging `json:"logging"`

	// Listen configures the HTTP server where DENNIS will listen for
	// requests.
	//
	// Required.
	Listen *Listener `json:"listen"`

	// Resolvers configures the upstream DNS resolvers that DENNIS will
	// queries with.
	//
	// Required. At least on Resolver is required.
	Resolvers []*Resolver `json:"resolvers"`

	// DB configures where Query objects will be stored.
	//
	// Required.
	DB DB `json:"db"`
}

// Logging configures the level and format of the log entries emitted by
// DENNIS.
type Logging struct {
	// Debug enables DEBUG-level log entries, otherwise only INFO-level log
	// entries are emitted.
	Debug bool `json:"debug"`

	// JSON configures DENNIS to write JSON-formatted log entries, otherwise
	// text-formatted is used.
	JSON bool `json:"json"`
}

// GetLogger returns a structured logger configured from Logging writing to
// STDOUT.
func (l *Logging) GetLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if l.Debug {
		opts.Level = slog.LevelDebug
	}

	if l.JSON {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

// Listener configures an HTTP server where DENNIS will listen for web and
// API requests from users.
type Listener struct {
	// Addr is the `[host]:<port>` where DENNIS will list. If no host/IP
	// address is specified, DENNIS will listen on all interfaces. IPv4
	// addresses and DENNIS can be specified as-is, however IPv6 addresses
	// must be nested within square brackets.
	//
	// Required.
	Addr string `json:"addr"`
}

// Resolver is one of the DNS resolvers that will be queried for records when
// requested by a user.
type Resolver struct {
	// Name is the name of the Resolver that will be displayed in the web
	// interface.
	//
	// Required.
	Name string `json:"name"`

	// Addr is the IP address of the DNS resolver. If it is not on port 53, set
	// `port` below.
	//
	// Required.
	Addr string `json:"addr"`

	// Port is the port number on the host addr where the DNS resolver accepts
	// queries. If not set, port 53 will be used.
	Port int `json:"port,omitempty"`
}

// DB configures where Query objects will be stored between requests. Only one
// database backend can be configured at once.
type DB struct {
	// File configures a local file as the database.
	File *FileDB `json:"file,omitempty"`

	// Redis configures an in-memory Redis server as the database.
	Redis *RedisDB `json:"redis,omitempty"`
}

// FileDB configures a local file to store Query objects. This database backend
// is suitable for small deployments, consider a database-backed backend for
// larger deployments, such as PostgreSQL or Redis.
type FileDB struct {
	// Path is the file path where Query objects will be stored in the JSON
	// format. If the file does not exist, it will be created.
	//
	// Required.
	Path string `json:"path"`
}

// RedisDB configures a Redis server to store Query objects.
type RedisDB struct {
	// Addr is the `host:port` where the Redis server is configured to accept
	// network connections.
	//
	// Required.
	Addr string `json:"addr"`

	// DB is the ID of the database in Redis to use. If not set, `0` is used.
	DB int `json:"db,omitempty"`

	// Username is optionally set if the Redis server expects authentication.
	Username string `json:"username,omitempty"`

	// Password is optionally set if the Redis server expects authentication.
	Password string `json:"password,omitempty"`
}
