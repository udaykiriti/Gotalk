package handlers

import (
	"log"
	"net/http"
	"path/filepath"

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
	
	// Serve the static file. ADJUST PATH AS NEEDED.
	// In a real app, you might embed this or use a config.
	// For now, we assume it's in the project root.
	absPath, err := filepath.Abs("web/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, absPath)
}

func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(h.hub, w, r)
}
