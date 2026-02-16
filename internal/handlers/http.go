package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"gotalk/internal/ws"
)

type Handler struct {
	hub *ws.Hub
}

func NewHandler(hub *ws.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) ServeHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indexPath, err := resolveIndexPath()
	if err != nil {
		http.Error(w, "File not found", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, indexPath)
}

func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(h.hub, w, r)
}

func resolveIndexPath() (string, error) {
	candidates := []string{}

	// Runtime cwd path (works with `go run` from project root).
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "web", "index.html"))
	}

	// Binary-relative path (works when deploying with web/ next to the binary).
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "web", "index.html"))
	}

	// Source-relative path (works in most local dev builds).
	if _, thisFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(thisFile), "..", "..", "web", "index.html"))
	}

	for _, p := range candidates {
		clean := filepath.Clean(p)
		if info, err := os.Stat(clean); err == nil && !info.IsDir() {
			return clean, nil
		}
	}

	return "", os.ErrNotExist
}
