package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type API struct {
	store *Store
	hub   *Hub
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// GET /windows — list every window with its current playlist.
func (a *API) handleListWindows(w http.ResponseWriter, r *http.Request) {
	windows, err := a.store.ListWindows(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load windows")
		return
	}
	writeJSON(w, http.StatusOK, windows)
}

// POST /windows/{id}/media — append a new media item to a window's playlist.
// Broadcasts a "playlist_updated" event so every connected client (including
// other browser tabs/windows) picks up the change immediately.
func (a *API) handleAddMedia(w http.ResponseWriter, r *http.Request) {
	windowID := r.PathValue("id")

	var req AddMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	switch req.Type {
	case MediaImage, MediaVideo, MediaBlank:
	default:
		writeError(w, http.StatusBadRequest, "type must be one of: image, video, blank")
		return
	}
	if req.Type != MediaBlank && req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required for image/video items")
		return
	}
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = 8 // sensible default for images/blank; ignored for video
	}

	item := MediaItem{
		ID:              generateID(),
		Type:            req.Type,
		URL:             req.URL,
		DurationSeconds: req.DurationSeconds,
		AddedAt:         time.Now().UTC(),
	}

	updated, err := a.store.AddMedia(r.Context(), windowID, item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add media")
		return
	}

	a.hub.Broadcast(Event{
		Type:    "playlist_updated",
		Payload: PlaylistUpdatedPayload{Window: *updated},
	})

	writeJSON(w, http.StatusCreated, updated)
}

// POST /sync — tell every window to display the same media item at the
// same time, for the given duration, then resume its own sequence.
// All the actual "show this now, then resume" logic lives in the
// frontend; the backend's job is just to fan the event out to every
// connected client at (as close to) the same instant as possible.
func (a *API) handleSync(w http.ResponseWriter, r *http.Request) {
	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.MediaItem.Type == "" || (req.MediaItem.Type != MediaBlank && req.MediaItem.URL == "") {
		writeError(w, http.StatusBadRequest, "mediaItem with a valid type/url is required")
		return
	}
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = 8
	}

	a.hub.Broadcast(Event{
		Type: "sync",
		Payload: SyncPayload{
			MediaItem:       req.MediaItem,
			DurationSeconds: req.DurationSeconds,
		},
	})

	writeJSON(w, http.StatusOK, map[string]string{"status": "sync triggered"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
