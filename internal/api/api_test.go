package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	New().Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("got status %q, want %q", body["status"], "ok")
	}
}

func TestListItemsStartsEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	New().Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("got body %q, want %q", got, "[]")
	}
}

func TestCreateThenListItem(t *testing.T) {
	h := New().Routes()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/items", strings.NewReader(`{"name":"widget"}`))
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got status %d, want %d", rec.Code, http.StatusCreated)
	}

	var created Item
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID != 1 || created.Name != "widget" {
		t.Fatalf("got %+v, want {ID:1 Name:widget}", created)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items", nil))

	var items []Item
	if err := json.NewDecoder(rec.Body).Decode(&items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 1 || items[0] != created {
		t.Errorf("got %+v, want [%+v]", items, created)
	}
}

func TestCreateItemRejectsBadInput(t *testing.T) {
	tests := map[string]string{
		"missing name": `{}`,
		"empty name":   `{"name":""}`,
		"invalid json": `{`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/items", strings.NewReader(body))
			New().Routes().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}
