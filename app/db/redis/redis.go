package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jamescun/dennis/app/config"
	"github.com/jamescun/dennis/app/db"
	"github.com/jamescun/dennis/app/models"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
)

// DB is a database implementation backed by an in-memory Redis database.
type DB struct {
	// conn is an interface containing just the methods we need from the Redis
	// client.
	conn interface {
		JSONArrAppend(ctx context.Context, key, path string, values ...any) *redis.IntSliceCmd
		JSONGet(ctx context.Context, key string, paths ...string) *redis.JSONCmd
		JSONSet(ctx context.Context, key, path string, value any) *redis.StatusCmd
	}
}

// New initializes a new Redis database implementation. The PING command will
// be attempted after creating the connection to validate connectivity and
// authentication (if configured).
func New(ctx context.Context, opts *redis.Options) (*DB, error) {
	conn := redis.NewClient(opts)
	err := conn.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("could not ping redis: %w", err)
	}

	return &DB{conn: conn}, nil
}

// FromConfig configures a Redis database implementation from a configuration
// object supplied by the user.
func FromConfig(ctx context.Context, cfg *config.RedisDB) (*DB, error) {
	return New(ctx, &redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		Username: cfg.Username,
		Password: cfg.Password,
	})
}

func (d *DB) CreateQuery(ctx context.Context, query *models.Query) error {
	query.ID = uuid.Must(uuid.NewV7())
	query.CreatedAt = time.Now().UTC()

	bytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	err = d.conn.JSONSet(ctx, queryKey(query.ID), "$", bytes).Err()
	if err != nil {
		return fmt.Errorf("could not set JSON key: %w", err)
	}

	return nil
}

func (d *DB) GetQueryByID(ctx context.Context, id uuid.UUID) (*models.Query, error) {
	result, err := d.conn.JSONGet(ctx, queryKey(id), ".").Result()
	if errors.Is(err, redis.Nil) {
		return nil, db.ErrQueryNotFound
	} else if err != nil {
		return nil, fmt.Errorf("could not get JSON key: %w", err)
	}

	query := &models.Query{}

	err = json.Unmarshal([]byte(result), query)
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	return query, nil
}

func (d *DB) UpdateQuery(ctx context.Context, query *models.Query) error {
	if query.FinishedAt != nil {
		err := d.conn.JSONSet(ctx, queryKey(query.ID), "$.finishedAt", strconv.Quote(query.FinishedAt.Format(time.RFC3339Nano))).Err()
		if err != nil {
			return fmt.Errorf("could not update JSON key: %w", err)
		}
	}

	return nil
}

func (d *DB) CreateLookup(ctx context.Context, queryID uuid.UUID, lookup *models.Lookup) error {
	bytes, err := json.Marshal(lookup)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}

	_, err = d.conn.JSONArrAppend(ctx, queryKey(queryID), "$.lookups", bytes).Result()
	if err != nil {
		return fmt.Errorf("could not set JSON key: %w", err)
	}

	return nil
}

// queryKey generates a stringified key for Redis.
func queryKey(id uuid.UUID) string {
	return "dennis:query:" + id.String()
}
