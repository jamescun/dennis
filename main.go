package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jamescun/dennis/app"
	"github.com/jamescun/dennis/app/config"
	"github.com/jamescun/dennis/app/db"
	"github.com/jamescun/dennis/app/db/file"
	"github.com/jamescun/dennis/app/db/postgres"
	"github.com/jamescun/dennis/app/db/redis"
	"github.com/jamescun/dennis/app/pkg/build"
	"github.com/jamescun/dennis/app/pkg/http/web"
)

var (
	configFile  = flag.String("config", "/etc/dennis/config.yml", "path to configuration JSON or YAML")
	showVersion = flag.Bool("version", false, "show version information")
)

func run(ctx context.Context, configFile string) int {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg, err := config.Read(configFile)
	if err != nil {
		return exitError(2, "config: %s", err)
	}

	log := cfg.Logging.GetLogger()

	queryMaxAge := time.Duration(0)
	if cfg.QueryMaxAge > 0 {
		queryMaxAge = time.Duration(cfg.QueryMaxAge) * time.Second
	}

	conn, err := getDB(ctx, cfg.DB, queryMaxAge)
	if err != nil {
		return exitError(1, "db: %s", err)
	}

	if queryMaxAge > 0 {
		log.Debug("beginning to expire old queries", slog.Duration("max_age", queryMaxAge))

		go expireOldQueries(ctx, log, conn, queryMaxAge)
	}

	api := app.NewServer(conn, cfg.Resolvers, log)
	ui := app.NewUI(api, log)

	r := web.New(log)
	r.Route("/", ui.Routes)

	s := &http.Server{
		Addr:    cfg.Listen.Addr,
		Handler: r,
	}

	// launch goroutine to initiate a graceful shutdown when an interrupt is
	// received.
	go func() {
		<-ctx.Done()

		// the parent context has already been canceled, create a new base
		// context for our graceful shutdown timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Info("shutdown DENNIS gracefully...")

		err := s.Shutdown(ctx)
		if err != nil {
			log.Error("could not shutdown gracefully", slog.String("error", err.Error()))
		}
	}()

	log.Info(
		"starting DENNIS...",
		slog.String("addr", cfg.Listen.Addr),
		slog.String("version", build.GetVersion()), slog.String("commit", build.GetCommit(7)),
	)

	err = s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("DENNIS server error", slog.String("error", err.Error()))
	}

	return 0
}

// getDB configures a database backend from the configuration file.
func getDB(ctx context.Context, cfg config.DB, maxAge time.Duration) (db.DB, error) {
	switch {
	case cfg.File != nil:
		conn, err := file.FromConfig(ctx, cfg.File)
		if err != nil {
			return nil, fmt.Errorf("file: %w", err)
		}

		return conn, nil

	case cfg.Postgres != nil:
		conn, err := postgres.FromConfig(ctx, cfg.Postgres)
		if err != nil {
			return nil, fmt.Errorf("postgres: %w", err)
		}

		return conn, nil

	case cfg.Redis != nil:
		conn, err := redis.FromConfig(ctx, cfg.Redis, maxAge)
		if err != nil {
			return nil, fmt.Errorf("redis: %w", err)
		}

		return conn, nil

	default:
		return nil, fmt.Errorf("no database configured")
	}
}

// expireOldQueries is a scheduled task that runs every maxAge/2 to remove any
// Queries that are no longer requires as per the configuration of QueryMaxAge.
func expireOldQueries(ctx context.Context, log *slog.Logger, conn db.DB, maxAge time.Duration) {
	ticker := time.NewTicker(maxAge / 2)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		err := conn.DeleteQueriesOlderThan(ctx, maxAge)
		if err != nil {
			log.Error("could not expire old queries", slog.String("error", err.Error()))
		}

	case <-ctx.Done():
		// process is shutting down, stop expiring old Queries.
		return
	}
}

func main() {
	flag.Parse()

	// if requested, print version information then exit.
	if *showVersion {
		fmt.Printf("Version: %s\nCommit:  %s\n", build.GetVersion(), build.GetCommit(7))
		return
	}

	os.Exit(run(context.Background(), *configFile))
}

// exitError prints an formattable error message to STDERR and returns the
// expected exit status of os.Exit().
func exitError(code int, format string, args ...any) int {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	return code
}
