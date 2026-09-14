package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store wraps the MongoDB collection used to persist windows and their
// playlists. Each Window document embeds its full playlist array, since
// playlists are small (well under MongoDB's 16MB document limit) and this
// keeps reads/writes simple (one document per window).
type Store struct {
	windows *mongo.Collection
}

// NewStore connects to MongoDB using the given URI and database name.
func NewStore(uri, dbName string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Explicitly set a minimum TLS version. Some container/hosting
	// environments (e.g. certain Render instances) fail to auto-negotiate
	// TLS with MongoDB Atlas, surfacing as "remote error: tls: internal
	// error". Forcing TLS 1.2 avoids that negotiation failure.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetTLSConfig(tlsConfig))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Store{
		windows: client.Database(dbName).Collection("windows"),
	}, nil
}

// ListWindows returns every window, ordered by ID for stable output.
func (s *Store) ListWindows(ctx context.Context) ([]Window, error) {
	cur, err := s.windows.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	windows := []Window{}
	if err := cur.All(ctx, &windows); err != nil {
		return nil, err
	}
	return windows, nil
}

// GetWindow fetches a single window by ID.
func (s *Store) GetWindow(ctx context.Context, id string) (*Window, error) {
	var w Window
	err := s.windows.FindOne(ctx, bson.M{"_id": id}).Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// AddMedia appends a new item to a window's playlist and returns the
// updated window.
func (s *Store) AddMedia(ctx context.Context, windowID string, item MediaItem) (*Window, error) {
	_, err := s.windows.UpdateOne(ctx,
		bson.M{"_id": windowID},
		bson.M{"$push": bson.M{"playlist": item}},
	)
	if err != nil {
		return nil, err
	}
	return s.GetWindow(ctx, windowID)
}

// SeedIfEmpty inserts example windows/playlists the first time the app
// runs against an empty database, so the deployed app has something to
// show out of the box (per the assignment's "seed data" requirement).
func (s *Store) SeedIfEmpty(ctx context.Context) error {
	count, err := s.windows.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	log.Println("seeding example windows...")

	now := time.Now().UTC()
	seed := []interface{}{
		Window{
			ID:   "window-1",
			Name: "Window 1",
			Playlist: []MediaItem{
				{ID: "m1", Type: MediaImage, URL: "https://picsum.photos/id/1015/1280/720", DurationSeconds: 8, AddedAt: now},
				{ID: "m2", Type: MediaVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", DurationSeconds: 12, AddedAt: now},
				{ID: "m3", Type: MediaImage, URL: "https://picsum.photos/id/1025/1280/720", DurationSeconds: 8, AddedAt: now},
			},
		},
		Window{
			ID:   "window-2",
			Name: "Window 2",
			Playlist: []MediaItem{
				{ID: "m4", Type: MediaImage, URL: "https://picsum.photos/id/1035/1280/720", DurationSeconds: 8, AddedAt: now},
				{ID: "m5", Type: MediaImage, URL: "https://picsum.photos/id/1045/1280/720", DurationSeconds: 8, AddedAt: now},
				{ID: "m6", Type: MediaBlank, DurationSeconds: 4, AddedAt: now},
			},
		},
		Window{
			ID:   "window-3",
			Name: "Window 3",
			Playlist: []MediaItem{
				{ID: "m7", Type: MediaVideo, URL: "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", DurationSeconds: 12, AddedAt: now},
				{ID: "m8", Type: MediaImage, URL: "https://picsum.photos/id/1055/1280/720", DurationSeconds: 8, AddedAt: now},
			},
		},
	}

	_, err = s.windows.InsertMany(ctx, seed)
	return err
}