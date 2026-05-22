package fetch

import (
	"context"
	"testing"
)

func TestRunWritesCheckpoint(t *testing.T) {
	ctx := context.Background()
	store, err := NewStoreFromURI(ctx, "file://"+t.TempDir())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	cfg := Config{
		SourceURL: "https://example.invalid/api",
		Format:    "ndjson",
		Window:    "test-window",
	}
	res, err := Run(ctx, cfg, store, &Checkpoint{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Window != "test-window" {
		t.Errorf("Window = %q, want %q", res.Window, "test-window")
	}
	if res.NewCheckpoint == nil || res.NewCheckpoint.Cursor != "test-window" {
		t.Errorf("NewCheckpoint = %#v, want cursor=test-window", res.NewCheckpoint)
	}

	if err := SaveCheckpoint(ctx, store, res.NewCheckpoint); err != nil {
		t.Fatalf("SaveCheckpoint: %v", err)
	}
	got, err := LoadCheckpoint(ctx, store)
	if err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}
	if got.Cursor != "test-window" {
		t.Errorf("LoadCheckpoint.Cursor = %q, want %q", got.Cursor, "test-window")
	}
}

func TestRunRequiresSource(t *testing.T) {
	ctx := context.Background()
	store, err := NewStoreFromURI(ctx, "file://"+t.TempDir())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	if _, err := Run(ctx, Config{}, store, &Checkpoint{}); err == nil {
		t.Fatal("Run with empty SourceURL should fail")
	}
}
