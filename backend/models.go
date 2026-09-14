package main

import "time"

// MediaType is restricted to the three kinds of playlist entries the
// assignment describes.
type MediaType string

const (
	MediaImage MediaType = "image"
	MediaVideo MediaType = "video"
	MediaBlank MediaType = "blank"
)

// MediaItem is a single entry in a window's playlist.
//
// DurationSeconds is used by the frontend to know how long to show an
// "image" or "blank" item before advancing to the next one. For "video"
// items the frontend instead advances when the video's own playback ends
// (the browser knows the real duration), so DurationSeconds is optional
// there and only used as a fallback if the video fails to load.
type MediaItem struct {
	ID              string    `bson:"id" json:"id"`
	Type            MediaType `bson:"type" json:"type"`
	URL             string    `bson:"url,omitempty" json:"url,omitempty"`
	DurationSeconds int       `bson:"durationSeconds" json:"durationSeconds"`
	AddedAt         time.Time `bson:"addedAt" json:"addedAt"`
}

// Window represents one display window with its own looping playlist.
type Window struct {
	ID       string      `bson:"_id" json:"id"`
	Name     string      `bson:"name" json:"name"`
	Playlist []MediaItem `bson:"playlist" json:"playlist"`
}

// SyncRequest is the payload for POST /sync.
type SyncRequest struct {
	MediaItem       MediaItem `json:"mediaItem"`
	DurationSeconds int       `json:"durationSeconds"`
}

// AddMediaRequest is the payload for POST /windows/{id}/media.
type AddMediaRequest struct {
	Type            MediaType `json:"type"`
	URL             string    `json:"url"`
	DurationSeconds int       `json:"durationSeconds"`
}
