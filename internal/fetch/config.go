// Package fetch contains the shared types and runtime for the Cloud Run
// job. The cmd/fetch entrypoint is intentionally thin — it parses flags
// and delegates the actual work to fetch.Run.
//
// Customise the source-specific bits (the upstream client, the parsing /
// normalising step, the output writer) for the data source described in
// DESIGN.md. Keep the Store interface and checkpoint shape intact unless
// DESIGN.md calls for something different.
package fetch

// Config parameterises a single Run invocation. One Cloud Run job
// execution does one Run start to end — there is no daemon loop.
type Config struct {
	SourceURL string
	Format    string // "ndjson", "json", "csv", "parquet"
	Window    string // window key; empty → resume from checkpoint
}

// Result summarises one Run.
type Result struct {
	Window        string
	Wrote         []string // object paths written, relative to the store root
	Records       int
	NewCheckpoint *Checkpoint
}
