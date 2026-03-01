package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	apiv1 "github.com/jamescun/dennis/api/v1"
	"github.com/jamescun/dennis/app/config"
	"github.com/jamescun/dennis/app/db"
	"github.com/jamescun/dennis/app/models"

	"codeberg.org/miekg/dns"
	"github.com/gofrs/uuid"
)

// Server is an implementation of api/v1/apiv1.API backed by the database. It is
// consumed by both the API and Web interfaces.
type Server struct {
	db  db.DB
	rsv []*resolver
	wg  *sync.WaitGroup
	log *slog.Logger
}

type resolver struct {
	name   string
	addr   string
	client interface {
		Exchange(ctx context.Context, msg *dns.Msg, network, address string) (*dns.Msg, time.Duration, error)
	}
}

// NewServer initializes a new Server implementation of api/v1/apiv1.API backed
// by the given database. log is the destination for error messages generated
// by the asynchronous resolution process.
func NewServer(db db.DB, rsv []*config.Resolver, log *slog.Logger) *Server {
	s := &Server{
		db:  db,
		wg:  new(sync.WaitGroup),
		log: log,
	}

	client := new(dns.Client)

	for _, r := range rsv {
		port := "53"
		if r.Port > 0 {
			port = strconv.Itoa(r.Port)
		}

		s.rsv = append(s.rsv, &resolver{
			name:   r.Name,
			addr:   net.JoinHostPort(r.Addr, port),
			client: client,
		})
	}

	return s
}

// Close waits until all resolutions have completed before returning, as part
// of a graceful shutdown.
func (s *Server) Close() error {
	s.wg.Wait()
	return nil
}

func (s *Server) resolveAll(query *models.Query) {
	defer s.wg.Done()

	wg := new(sync.WaitGroup)
	log := s.log.With(
		slog.String("query_id", query.ID.String()),
		slog.String("query_type", query.Type),
		slog.String("query_name", query.Name),
	)

	// this context must be detached from the request context, as it needs to
	// continue after the end of the requests lifecycle.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, rsv := range s.rsv {
		wg.Add(1)
		go s.resolve(ctx, wg, log, rsv, query)
	}

	wg.Wait()

	now := time.Now().UTC()
	query.FinishedAt = &now

	err := s.db.UpdateQuery(ctx, query)
	if err != nil {
		log.Error("could not update query", slog.String("error", err.Error()))
	}
}

func (s *Server) resolve(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger, rsv *resolver, query *models.Query) {
	defer wg.Done()

	log.Debug("starting resolution...", slog.String("resolver", rsv.name))
	defer log.Debug("resolution complete", slog.String("resolver", rsv.name))

	req := dns.NewMsg(query.Name, dns.StringToType[query.Type])
	res, rtt, err := rsv.client.Exchange(ctx, req, "udp", rsv.addr)
	if err != nil {
		log.Error("could not resolve query", slog.String("resolver", rsv.name), slog.String("error", err.Error()))
		return
	}

	l := &models.Lookup{
		Resolver:   rsv.name,
		RTT:        int(rtt / time.Millisecond),
		ResolvedAt: time.Now().UTC(),
	}

	if res.Rcode != dns.RcodeSuccess {
		l.Error = dns.RcodeToString[res.Rcode]
	} else if errors.Is(err, context.Canceled) {
		l.Error = "CANCELED"
	} else {
		for _, answer := range res.Answer {
			rr := models.RecordFromRR(answer)
			if rr != nil {
				l.Records = append(l.Records, rr)
			}
		}
	}

	err = s.db.CreateLookup(ctx, query.ID, l)
	if err != nil {
		log.Error("could not create lookup", slog.String("resolver", rsv.name), slog.String("error", err.Error()))
		return
	}
}

func (s *Server) CreateQuery(ctx context.Context, req *apiv1.CreateQueryRequest) (*apiv1.CreateQueryResponse, error) {
	s.wg.Add(1)
	defer s.wg.Done()

	if err := req.Validate(); err != nil {
		return nil, err
	}

	query := &models.Query{
		Type: req.Type,
		Name: req.Name,

		// NOTE(jc): cannot be null, Redis will not append to a null value.
		Lookups: []*models.Lookup{},
	}

	err := s.db.CreateQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	s.wg.Add(1)
	go s.resolveAll(query)

	return &apiv1.CreateQueryResponse{
		Query: query,
	}, nil
}

func (s *Server) GetQuery(ctx context.Context, req *apiv1.GetQueryRequest) (*apiv1.GetQueryResponse, error) {
	s.wg.Add(1)
	defer s.wg.Done()

	if err := req.Validate(); err != nil {
		return nil, err
	}

	id, err := uuid.FromString(req.ID)
	if err != nil {
		return nil, &apiv1.Error{Code: apiv1.ErrorCodeBadRequest, Field: ".id", Message: "Invalid UUID for Query ID"}
	}

	query, err := s.db.GetQueryByID(ctx, id)
	if errors.Is(err, db.ErrQueryNotFound) {
		return nil, &apiv1.Error{Code: apiv1.ErrorCodeNotFound, Message: "Query not found by ID"}
	} else if err != nil {
		return nil, err
	}

	return &apiv1.GetQueryResponse{
		Query: query,
	}, nil
}
