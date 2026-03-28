package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/agent19710101/skillforge/internal/catalog"
)

type Server struct {
	index *catalog.Index
}

func NewServer(index *catalog.Index) *Server {
	return &Server{index: index}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/skills", s.handleSkills)
	mux.HandleFunc("/api/v1/skills/", s.handleSkillByName)
	mux.HandleFunc("/api/v1/index/status", s.handleIndexStatus)
	return mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	opts, err := catalog.ParseListOptions(
		r.URL.Query().Get("validation"),
		r.URL.Query().Get("offset"),
		r.URL.Query().Get("limit"),
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}

	if opts.Limit == 0 {
		opts.Limit = 50
	}
	if opts.Limit > 200 {
		opts.Limit = 200
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"skills": s.index.List(opts),
		"total":  s.index.Total(catalog.ListOptions{Validation: opts.Validation}),
		"offset": opts.Offset,
		"limit":  opts.Limit,
	})
}

func (s *Server) handleSkillByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	if name == "" || strings.Contains(name, "/") {
		writeError(w, http.StatusNotFound, "not_found", "skill not found")
		return
	}

	skill, ok := s.index.Get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (s *Server) handleIndexStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.index.Status())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}
