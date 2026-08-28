package core

import (
	"github.com/pafthang/arcanum/pkg/svcutil"
	loggmodels "github.com/pafthang/arcanum/services/logg/models"

	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/gate/internal/auth"
	"github.com/pafthang/arcanum/services/gate/internal/config"
	"github.com/pafthang/arcanum/services/gate/internal/edge"
	"github.com/pafthang/arcanum/services/gate/internal/middleware"
	"github.com/pafthang/arcanum/services/gate/internal/ops"
	"github.com/pafthang/arcanum/services/gate/internal/policy"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
)

// Gate — HTTP edge: discovers public mini routes and proxies via NATS.
type Gate struct {
	cfg       *config.Config
	mux       *http.ServeMux
	server    *http.Server
	nc        *nats.Conn
	table     *edge.Table
	discover  *edge.Discoverer
	edgeH     *edge.Handler
	authers   []auth.Authenticator
	policyEng policy.Engine
	metrics   *ops.Metrics
	log       *slog.Logger
	loggC     *loggclient.Client
}

// New creates a Gate. nc may be nil; Start will connect using cfg.NATSURL.
func New(cfg *config.Config, nc *nats.Conn) (*Gate, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	g := &Gate{
		cfg:   cfg,
		mux:   http.NewServeMux(),
		nc:    nc,
		table: edge.NewTable(),
		log:   slog.Default(),
	}

	if cfg.Metrics.Enabled {
		g.metrics = ops.NewMetrics()
	}

	eng, err := policy.NewEngine(cfg.Policy)
	if err != nil {
		return nil, fmt.Errorf("policy engine: %w", err)
	}
	g.policyEng = eng

	// Build authenticators — default is HS256 shared secret (matches auth service).
	methods := cfg.Auth.Methods
	if len(methods) == 0 {
		methods = []string{"hmac"}
	}
	for _, kind := range methods {
		var aCfg any
		switch kind {
		case "bearer":
			aCfg = cfg.Auth.Bearer
		case "apikey":
			aCfg = cfg.Auth.APIKey
		case "jwt", "hmac", "hmac-jwt", "hs256":
			aCfg = auth.HMACJWTConfig{
				Secret:   cfg.Auth.HMACSecret,
				Issuer:   svcutil.First(cfg.Auth.JWT.Issuer, cfg.Auth.HMACIssuer),
				Audience: cfg.Auth.JWT.Audience,
			}
			if kind == "jwt" && cfg.Auth.HMACSecret == "" && cfg.Auth.JWT.JWKSURL != "" {
				aCfg = map[string]string{
					"issuer":   cfg.Auth.JWT.Issuer,
					"audience": cfg.Auth.JWT.Audience,
					"jwks_url": cfg.Auth.JWT.JWKSURL,
				}
			}
		case "mtls":
			allowed := make(map[string]bool)
			for _, cn := range cfg.Auth.MTLS.AllowedCNs {
				allowed[cn] = true
			}
			aCfg = allowed
		}
		a, err := auth.NewAuthenticator(kind, aCfg)
		if err != nil {
			// hmac without secret in misconfig — fail early
			return nil, fmt.Errorf("auth %s: %w", kind, err)
		}
		g.authers = append(g.authers, a)
	}

	g.registerLocalRoutes()
	return g, nil
}

func (g *Gate) registerLocalRoutes() {
	g.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	g.mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// ready when we have at least tried discovery (table may be empty early)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	if g.metrics != nil {
		g.mux.Handle(g.cfg.Metrics.Path, g.metrics.Handler())
	}
	g.mux.HandleFunc("/openapi.json", ops.OpenAPIHandler)
}

// Handler returns the root HTTP handler.
func (g *Gate) Handler() http.Handler {
	// Edge handler is set in Start once NATS is up; fallback 503 until then.
	edgeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.edgeH != nil {
			g.edgeH.ServeHTTP(w, r)
			return
		}
		// Local probes only until edge is ready
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			g.mux.ServeHTTP(w, r)
			return
		}
		http.Error(w, "gate starting", http.StatusServiceUnavailable)
	})

	// Prefer mux for known local paths; otherwise edge.
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz", "/readyz", "/openapi.json":
			g.mux.ServeHTTP(w, r)
			return
		}
		if g.cfg.Metrics.Enabled && r.URL.Path == g.cfg.Metrics.Path {
			g.mux.ServeHTTP(w, r)
			return
		}
		edgeHandler.ServeHTTP(w, r)
	})

	var h http.Handler = root
	if g.policyEng != nil {
		h = policy.Middleware(g.policyEng)(h)
	}
	// Note: per-route auth is done in edge.Handler — no global auth wall.
	if g.cfg.RateLimit.Enabled {
		h = middleware.RateLimit(middleware.RateLimitOpts{
			RequestsPerSec:  g.cfg.RateLimit.RequestsPerSec,
			Burst:           g.cfg.RateLimit.Burst,
			CleanupInterval: g.cfg.RateLimit.CleanupInterval,
			OnReject: func(r *http.Request) {
				if g.loggC != nil {
					ip := r.Header.Get("X-Forwarded-For")
					if ip == "" {
						ip = r.RemoteAddr
					}
					g.loggC.AppendActivityAsync(&loggmodels.Activity{
						Type:    "gate.ratelimit.exceeded",
						Summary: "Rate limit exceeded",
						Payload: map[string]any{
							"path": r.URL.Path,
							"ip":   ip,
						},
					})
				}
			},
		})(h)
	}
	h = middleware.CORS(g.cfg.CORS)(h)
	if g.metrics != nil {
		h = g.metrics.Middleware(h)
	}
	return h
}

// Start connects NATS (if needed), starts discovery, and serves HTTP.
func (g *Gate) Start(ctx context.Context) error {
	if g.nc == nil {
		url := g.cfg.NATSURL
		if url == "" {
			url = nats.DefaultURL
		}
		nc, err := nats.Connect(url, nats.Name("gate"), nats.MaxReconnects(-1))
		if err != nil {
			return fmt.Errorf("nats: %w", err)
		}
		g.nc = nc
	}

	client, err := mini.NewClient(g.nc, mini.WithClientTimeout(g.cfg.RequestTimeout))
	if err != nil {
		return fmt.Errorf("mini client: %w", err)
	}

	g.loggC = loggclient.MustNew(g.nc)

	g.edgeH = &edge.Handler{
		NC:        g.nc,
		Table:     g.table,
		Auth:      g.authers,
		Claims:    edge.ParseClaimHeaders(g.cfg.ClaimHeaders),
		DefaultTO: g.cfg.RequestTimeout,
		MaxBody:   g.cfg.MaxBodyBytes,
		Client:    client,
		Log:       g.log,
		Logg:      g.loggC,
	}

	g.discover = &edge.Discoverer{
		NC:       g.nc,
		Table:    g.table,
		Interval: g.cfg.DiscoverInterval,
		Wait:     g.cfg.DiscoverWait,
		Log:      g.log,
	}
	g.discover.Start(ctx)

	tlsCfg, err := g.cfg.TLS.BuildTLSConfig()
	if err != nil {
		return fmt.Errorf("tls config: %w", err)
	}

	g.server = &http.Server{
		Addr:         g.cfg.Server.ListenAddr,
		Handler:      g.Handler(),
		ReadTimeout:  g.cfg.Server.ReadTimeout,
		WriteTimeout: g.cfg.Server.WriteTimeout,
		IdleTimeout:  g.cfg.Server.IdleTimeout,
		TLSConfig:    tlsCfg,
	}

	go func() {
		<-ctx.Done()
		g.discover.Stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), g.cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := g.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("gate: shutdown error: %v", err)
		}
		if g.nc != nil {
			g.nc.Close()
		}
	}()

	g.log.Info("gate listening",
		"addr", g.cfg.Server.ListenAddr,
		"nats", g.cfg.NATSURL,
		"auth_methods", g.cfg.Auth.Methods,
	)
	if tlsCfg != nil {
		return g.server.ListenAndServeTLS("", "")
	}
	return g.server.ListenAndServe()
}

// Stop shuts down the server.
func (g *Gate) Stop(ctx context.Context) error {
	if g.discover != nil {
		g.discover.Stop()
	}
	if g.server == nil {
		return nil
	}
	return g.server.Shutdown(ctx)
}

// Mux returns the local ServeMux (probes).
func (g *Gate) Mux() *http.ServeMux {
	return g.mux
}

// Table returns the live public route table.
func (g *Gate) Table() *edge.Table {
	return g.table
}

// Ensure unused import of time when RequestTimeout defaults use it.
var _ = time.Second
