package server

import (
	"encoding/json"
	"net/http"

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

	r.Route("/api/v1/projects", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.Post("/", s.project.Create)
		r.Get("/", s.project.List)
		r.Get("/{id}", s.project.Get)
	})

	r.Route("/api/v1/folders", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.Post("/", s.folder.Create)
		r.Get("/", s.folder.Tree)
	})

	r.Route("/api/v1/test-cases", func(r chi.Router) {
		r.Use(s.jwt.Middleware)
		r.Post("/", s.testcase.Create)
		r.Get("/", s.testcase.List)
		r.Get("/{id}", s.testcase.GetCurrent)
		r.Put("/{id}", s.testcase.Edit)
		r.Get("/{id}/versions", s.testcase.History)
		r.Get("/{id}/versions/{number}", s.testcase.GetVersion)
	})

	return r
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
