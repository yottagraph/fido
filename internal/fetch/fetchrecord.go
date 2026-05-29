package fetch

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"

	"github.com/example/fido-fetch/internal/fetchrecord"
)

// The fetch-record write path. A Run builds one fetchrecord.FetchMessage
// per window, and WriteFetchMessage serialises it to binary protobuf,
// zstd-compresses it, and writes it to the output store under
//   output/<YYYY-MM-DD>/<window-key>.binpb.zst
//
// This is the format the elemental ingest path consumes — it keys off the
// .binpb.zst suffix to zstd-decompress and proto.Unmarshal. The schema is
// vendored in proto/fetch_record.proto; see internal/fetchrecord for the
// generated types and the Fido skill for how to map a source's records
// into a FetchMessage (subject ProtoEntity + atoms).

// FetchRecordObjectPath returns the object path for one window's
// fetch-record output, relative to the store root.
//
//	output/<YYYY-MM-DD>/<window-key>.binpb.zst
//
// The date prefix is taken from the leading YYYY-MM-DD of a window key of
// the form YYYY-MM-DDTHH-MM-SSZ; window keys that don't start with a date
// are used as-is.
func FetchRecordObjectPath(window string) string {
	return "output/" + windowDate(window) + "/" + window + ".binpb.zst"
}

// windowDate extracts the YYYY-MM-DD prefix from a window key. If the key
// does not start with a date, the full key is returned.
func windowDate(window string) string {
	if i := strings.Index(window, "T"); i >= 0 {
		return window[:i]
	}
	return window
}

// WriteFetchMessage marshals msg to binary protobuf, zstd-compresses it,
// and writes it to store at the window's fetch-record object path. It
// returns the object path written.
//
// The .zst suffix carries the compression contract (matching the
// canonical fetcher), so the object is written with an opaque binary
// content type rather than a JSON/text one.
func WriteFetchMessage(ctx context.Context, store Store, window string, msg *fetchrecord.FetchMessage) (string, error) {
	// Deterministic marshal: a FetchMessage has several map fields (the
	// *_metadata maps) and Go randomises map iteration order, so the
	// default Marshal yields different bytes for the same logical
	// message across runs. Deterministic output keeps an identical
	// window byte-for-byte reproducible (content-addressing, digest
	// dedup, idempotent overwrites). Ingest is unaffected — it decodes
	// by field number.
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("fetchrecord: marshal: %w", err)
	}

	var buf bytes.Buffer
	enc, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return "", fmt.Errorf("fetchrecord: new zstd encoder: %w", err)
	}
	if _, err := enc.Write(raw); err != nil {
		_ = enc.Close()
		return "", fmt.Errorf("fetchrecord: zstd compress: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("fetchrecord: zstd close: %w", err)
	}

	path := FetchRecordObjectPath(window)
	if err := store.Write(ctx, path, buf.Bytes(), "application/octet-stream"); err != nil {
		return "", err
	}
	return path, nil
}
