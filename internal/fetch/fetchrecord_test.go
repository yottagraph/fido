package fetch

import (
	"bytes"
	"context"
	"testing"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"

	"github.com/example/fido-fetch/internal/fetchrecord"
)

func TestFetchRecordObjectPath(t *testing.T) {
	got := FetchRecordObjectPath("2024-04-01T12-00-00Z")
	want := "output/2024-04-01/2024-04-01T12-00-00Z.binpb.zst"
	if got != want {
		t.Errorf("FetchRecordObjectPath = %q, want %q", got, want)
	}
}

// TestWriteFetchMessageRoundTrip writes a FetchMessage through the
// fetch-record write path, then reads the object back, zstd-decompresses
// it, and proto-unmarshals it — mirroring what the elemental ingest path
// does — and asserts the message survives the round trip intact.
func TestWriteFetchMessageRoundTrip(t *testing.T) {
	ctx := context.Background()
	store, err := NewStoreFromURI(ctx, "file://"+t.TempDir())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	defer store.Close()

	const window = "2024-04-01T12-00-00Z"
	msg := &fetchrecord.FetchMessage{
		Citation:                "https://example.invalid/api?window=" + window,
		SourceDownloadTimestamp: 1_700_000_000_000_000,
		Records: []*fetchrecord.Record{{
			Source:    "example-source",
			Timestamp: 1_700_000_000_000_000,
			Subject: &fetchrecord.ProtoEntity{
				Name:   "Acme Corp",
				Flavor: "company",
			},
			Atoms: []*fetchrecord.Atom{{
				Property:  "annual_revenue",
				Value:     &fetchrecord.Atom_FloatVal{FloatVal: 1234.5},
				Timestamp: 1_700_000_000_000_000,
			}},
		}},
	}

	path, err := WriteFetchMessage(ctx, store, window, msg)
	if err != nil {
		t.Fatalf("WriteFetchMessage: %v", err)
	}
	if want := "output/2024-04-01/" + window + ".binpb.zst"; path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	raw, err := store.Read(ctx, path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	dec, err := zstd.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("zstd reader: %v", err)
	}
	defer dec.Close()
	var plain bytes.Buffer
	if _, err := plain.ReadFrom(dec); err != nil {
		t.Fatalf("zstd decompress: %v", err)
	}

	var got fetchrecord.FetchMessage
	if err := proto.Unmarshal(plain.Bytes(), &got); err != nil {
		t.Fatalf("proto unmarshal: %v", err)
	}
	if !proto.Equal(msg, &got) {
		t.Errorf("round-trip mismatch:\n got = %v\nwant = %v", &got, msg)
	}
	if len(got.Records) != 1 || got.Records[0].Subject.GetFlavor() != "company" {
		t.Errorf("unexpected decoded records: %v", got.Records)
	}
}
