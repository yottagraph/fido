package fetch

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Run is the heart of the Cloud Run job. Replace the body with the
// source-specific fetch logic for this project:
//
//  1. Resolve the window to fetch from cfg.Window or cp.Cursor.
//  2. Call into the upstream source client and collect the records.
//  3. Marshal them in cfg.Format and write them via store.Write under
//     the layout documented in DESIGN.md.
//  4. Return a Result with the new checkpoint and the object paths
//     written so the entrypoint can persist them.
//
// The default implementation is a no-op that records an empty window —
// it exists so the template compiles and tests pass before the source
// client is wired up.
func Run(ctx context.Context, cfg Config, store Store, cp *Checkpoint) (*Result, error) {
	if cfg.SourceURL == "" {
		return nil, fmt.Errorf("fetch: source URL is required")
	}
	if store == nil {
		return nil, fmt.Errorf("fetch: output store is required")
	}

	window := cfg.Window
	if window == "" {
		window = cp.Cursor
	}
	if window == "" {
		window = time.Now().UTC().Format("2006-01-02T15-04-05Z")
	}

	slog.Info("fetch: starting",
		"source", cfg.SourceURL,
		"window", window,
		"format", cfg.Format,
		"output_root", store.Root(),
	)

	// TODO: replace with the source-specific fetch + write loop for
	// this project. See DESIGN.md for the upstream contract and the
	// output layout.

	return &Result{
		Window:        window,
		NewCheckpoint: &Checkpoint{Cursor: window},
	}, nil
}
