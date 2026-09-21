// Package api contains the HTTP handlers for the demo service.
package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Server struct {
	mu     sync.Mutex
	items  []Item
	nextID int
	start  time.Time
}

func New() *Server {
	return &Server{nextID: 1, start: time.Now()}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/items", s.listItems)
	mux.HandleFunc("POST /api/items", s.createItem)
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"uptime": time.Since(s.start).Round(time.Second).String(),
	})
}

func (s *Server) listItems(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.items
	if items == nil {
		items = []Item{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createItem(w http.ResponseWriter, r *http.Request) {
	var in Item
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	s.mu.Lock()
	in.ID = s.nextID
	s.nextID++
	s.items = append(s.items, in)
	s.mu.Unlock()

	writeJSON(w, http.StatusCreated, in)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}
