package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server is the appctl HTTP daemon.
type Server struct {
	db *sql.DB
}

func New(db *sql.DB) *Server {
	return &Server{db: db}
}

func (s *Server) Run(addr string) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", s.handleHealth)

	r.Route("/apps", func(r chi.Router) {
		r.Get("/", s.handleListApps)
		r.Post("/", s.handleCreateApp)
		r.Route("/{name}", func(r chi.Router) {
			r.Get("/", s.handleGetApp)
			r.Patch("/", s.handleUpdateApp)
			r.Delete("/", s.handleDeleteApp)
			r.Post("/deploy", s.handleDeploy)
			r.Post("/start", s.handleStart)
			r.Post("/stop", s.handleStop)
			r.Post("/restart", s.handleRestart)
			r.Get("/deployments", s.handleListDeployments)
		})
	})

	r.Route("/config", func(r chi.Router) {
		r.Get("/", s.handleGetConfig)
		r.Patch("/", s.handleSetConfig)
	})

	r.Post("/webhooks/github/{name}", s.handleGitHubWebhook)

	return http.ListenAndServe(addr, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
