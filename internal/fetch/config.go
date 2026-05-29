// Package fetch contains the shared types and runtime for the Cloud Run
// job. The cmd/fetch entrypoint is intentionally thin — it parses flags
// and delegates the actual work to fetch.Run.
//
// Customise the source-specific bits (the upstream client, the parsing /
// normalising step, the FetchMessage builder) for the data source
// described in DESIGN.md. Keep the Store interface and checkpoint shape
// intact unless DESIGN.md calls for something different.
//
// Output is always fetch records: one fetchrecord.FetchMessage per
// window, written as a zstd-compressed binary-protobuf .binpb.zst object
// via WriteFetchMessage (see fetchrecord.go). There is no alternate
// output format.
package fetch

// Config parameterises a single Run invocation. One Cloud Run job
// execution does one Run start to end — there is no daemon loop.
type Config struct {
	SourceURL string
	Window    string // window key; empty → resume from checkpoint
}

// Result summarises one Run.
type Result struct {
	Window        string
	Wrote         []string // object paths written, relative to the store root
	Records       int
	NewCheckpoint *Checkpoint
}
