package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"gotalk/internal/ws"
	webassets "gotalk/web"
)

type Handler struct {
	hub *ws.Hub
}

func NewHandler(hub *ws.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) ServeHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	// Serve index.html for the root path
	if r.URL.Path == "/" {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		content, err := webassets.Assets.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
		return
	}

	// Serve static assets from the embedded filesystem
	// The embed FS contains "static/*"
	http.FileServer(http.FS(webassets.Assets)).ServeHTTP(w, r)
}

func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	ws.ServeWs(h.hub, w, r)
}

// HandleRooms serves the list of active rooms as JSON
func (h *Handler) HandleRooms(w http.ResponseWriter, r *http.Request) {
	rooms := h.hub.GetActiveRooms()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}
