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
//  3. Build a single fetchrecord.FetchMessage for the window: one
//     Record per entity (subject = ProtoEntity{name, flavor}), one Atom
//     per property/relationship, and the schema metadata maps populated
//     from schema.yaml. See the Fido skill for the mapping contract.
//  4. Persist it with WriteFetchMessage (zstd-compressed binary protobuf
//     at output/<YYYY-MM-DD>/<window>.binpb.zst) and return a Result
//     with the new checkpoint and the object paths written.
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
		"output_root", store.Root(),
	)

	// TODO: replace with the source-specific fetch loop for this
	// project. Collect the window's records, build a
	// fetchrecord.FetchMessage, and persist it with
	//   path, err := WriteFetchMessage(ctx, store, window, msg)
	// recording path in Result.Wrote. See DESIGN.md for the upstream
	// contract and the Fido skill for the FetchMessage mapping.

	return &Result{
		Window:        window,
		NewCheckpoint: &Checkpoint{Cursor: window},
	}, nil
}
