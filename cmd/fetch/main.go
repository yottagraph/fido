// Command fetch is the Cloud Run job entrypoint for this Fido fetch
// project. One invocation does one fetch window: pulls data from the
// configured source, writes the output to the configured GCS bucket,
// persists a checkpoint, and exits. Cloud Scheduler triggers the next
// invocation; this binary does not loop.
//
// Customise the flags, defaults, and output layout for the specific
// source described in DESIGN.md before shipping.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/example/fido-fetch/internal/fetch"
)

func main() {
	sourceURL := flag.String("source-url", "", "upstream source URL or URI; required")
	outputURI := flag.String("output", "", "output store URI (gs://bucket or file:///path); required")
	windowKey := flag.String("window", "", "window key for this invocation; empty → resume from checkpoint")
	format := flag.String("format", "ndjson", "output format: one of ndjson, json, csv, parquet")
	flag.Parse()

	if *sourceURL == "" {
		fatalf("--source-url is required")
	}
	if *outputURI == "" {
		fatalf("--output is required")
	}

	ctx := context.Background()
	store, err := fetch.NewStoreFromURI(ctx, *outputURI)
	if err != nil {
		fatalf("create output store: %v", err)
	}
	defer store.Close()

	cp, err := fetch.LoadCheckpoint(ctx, store)
	if err != nil {
		slog.Warn("fetch: checkpoint load failed; starting fresh", "error", err)
		cp = &fetch.Checkpoint{}
	}

	cfg := fetch.Config{
		SourceURL: *sourceURL,
		Format:    *format,
		Window:    *windowKey,
	}
	res, err := fetch.Run(ctx, cfg, store, cp)
	if err != nil {
		slog.Error("fetch: failed", "error", err)
		os.Exit(1)
	}
	if res.NewCheckpoint != nil {
		if err := fetch.SaveCheckpoint(ctx, store, res.NewCheckpoint); err != nil {
			slog.Warn("fetch: checkpoint save failed", "error", err)
		}
	}
	slog.Info("fetch: complete",
		"window", res.Window,
		"objects_written", len(res.Wrote),
		"records", res.Records,
	)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fetch: "+format+"\n", args...)
	os.Exit(1)
}
