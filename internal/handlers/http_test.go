package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotalk/internal/ws"
)

func TestServeHome_Success(t *testing.T) {
	h := NewHandler(ws.NewHub())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHome(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("expected text/html content type, got %q", rr.Header().Get("Content-Type"))
	}
	if rr.Body.Len() == 0 {
		t.Fatal("expected non-empty html response body")
	}
}

func TestServeHome_NotFoundForNonRootPath(t *testing.T) {
	h := NewHandler(ws.NewHub())
	req := httptest.NewRequest(http.MethodGet, "/not-root", nil)
	rr := httptest.NewRecorder()

	h.ServeHome(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestServeHome_MethodNotAllowedForNonGET(t *testing.T) {
	h := NewHandler(ws.NewHub())
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHome(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
