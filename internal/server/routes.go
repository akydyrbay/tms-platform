package server

import (
	"encoding/json"
	"net/http"
	"tms-platform/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", s.healthHandler)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", s.auth.Register)
		r.Post("/login", s.auth.Login)
	})

	writer := auth.RequireRoles(auth.WriteRoles...)
	manager := auth.RequireRoles(auth.ManageRoles...)

	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/", s.project.Create)
		r.Get("/", s.project.List)
		r.Get("/{id}", s.project.Get)
	})

	r.Route("/api/v1/folders", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/", s.folder.Create)
		r.Get("/", s.folder.Tree)
	})

	r.Route("/api/v1/test-cases", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/", s.testcase.Create)
		r.Get("/", s.testcase.List)
		r.Get("/{id}", s.testcase.GetCurrent)
		r.With(writer).Put("/{id}", s.testcase.Edit)
		r.Get("/{id}/versions", s.testcase.History)
		r.Get("/{id}/versions/{number}", s.testcase.GetVersion)
	})

	r.Route("/api/v1/suites", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/", s.suite.Create)
		r.Get("/", s.suite.List)
		r.Get("/{id}", s.suite.Get)
		r.With(writer).Post("/{id}/cases", s.suite.AddCase)
		r.With(manager).Delete("/{id}/cases/{caseID}", s.suite.RemoveCase)
	})

	r.Route("/api/v1/test-runs", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/", s.testrun.Create)
		r.Get("/", s.testrun.List)
		r.Get("/{id}", s.testrun.Get)
		r.With(writer).Patch("/{id}", s.testrun.SetStatus)
		r.Get("/{id}/results", s.testrun.Results)
		r.With(writer).Post("/{id}/results/{caseID}", s.testrun.Mark)
		r.Get("/{id}/stats", s.testrun.Stats)
		r.With(writer).Post("/{id}/ci-import", s.testrun.CIImport)
	})

	r.Route("/api/v1/run-results", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.With(writer).Post("/{resultID}/bugs", s.integration.Attach)
		r.Get("/{resultID}/bugs", s.integration.List)
	})
	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
