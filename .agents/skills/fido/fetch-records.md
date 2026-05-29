# Fetch records — the output contract

A Fido job's only output is **fetch records**: one protobuf
`FetchMessage` per window, serialised with `proto.Marshal`,
zstd-compressed, and written to
`output/<YYYY-MM-DD>/<window-key>.binpb.zst`. This is the exact shape the
elemental ingest path consumes — it keys off the `.binpb.zst` suffix to
zstd-decompress and `proto.Unmarshal` into a `FetchMessage`.

There is **no JSON / NDJSON / CSV output**. Don't add a format flag or an
alternate writer. The plumbing already exists:

- `proto/fetch_record.proto` — the vendored schema (do not change field
  numbers; the wire format depends on them).
- `internal/fetchrecord/` — generated Go types (`fetchrecord.FetchMessage`,
  `Record`, `Atom`, `ProtoEntity`, …). Regenerate with
  `scripts/gen-proto.sh` only if you edit the `.proto`.
- `internal/fetch/fetchrecord.go` — `WriteFetchMessage(ctx, store,
  window, msg)` does marshal → zstd → `store.Write(...binpb.zst)`.

Your job, in `internal/fetch/run.go`, is to fetch the window's data and
**build the `FetchMessage`**, then call `WriteFetchMessage`.

## The data model

A `FetchMessage` carries a batch of records plus the schema metadata for
the elements those records touch:

```
FetchMessage
├─ records []Record
│   ├─ source        string   (a stable id/uri for the source report)
│   ├─ timestamp     int64    (publication date, unix MICROseconds)
│   ├─ subject       ProtoEntity   (the entity this record is about)
│   │   ├─ name      string   (human-readable entity name)
│   │   └─ flavor    string   (entity type, from schema.yaml)
│   └─ atoms []Atom
│       ├─ property  string   (property/relationship name, from schema.yaml)
│       ├─ value     oneof { float_val double | str_val string | target ProtoEntity }
│       └─ timestamp int64    (when the fact holds, unix MICROseconds — ALWAYS set)
├─ citation                    string  (uri of the raw source for this window)
├─ source_download_timestamp   int64   (when we fetched it, unix micros)
└─ {flavor,property,relationship,attribute}_metadata  map<string,SchemaElementMeta>
```

Map your normalised domain records onto this as follows.

### 1. One `Record` per entity observation
Each thing your source describes (a company, a filing, an exchange-rate
observation, …) becomes one `Record`. Set:
- `subject.name` — the entity's human name (never a strong-id value).
- `subject.flavor` — the primary flavor from `schema.yaml` (e.g.
  `"company"`, `"currency"`).
- `source` — a stable identifier for the underlying report/row.
- `timestamp` — the source publication date in **unix microseconds**.

### 2. One `Atom` per property or relationship
For every property of the entity, add an `Atom`:
- `property` — the property name exactly as it appears in `schema.yaml`.
- `value`:
  - numeric → `&fetchrecord.Atom_FloatVal{FloatVal: x}`
  - string/date/enum → `&fetchrecord.Atom_StrVal{StrVal: s}`
  - a **relationship** to another entity → `&fetchrecord.Atom_Target{Target:
    &fetchrecord.ProtoEntity{Name: ..., Flavor: ...}}`
- `timestamp` — when the fact holds, unix micros. Use the most specific
  date available; otherwise copy `Record.timestamp`. **Never leave zero.**

### 3. Fill the metadata maps from `schema.yaml`
For each flavor/property/relationship/attribute you actually emit, add a
`SchemaElementMeta{Description, DisplayName}` keyed by the element's name,
sourced from `schema.yaml` + `DATA_DICTIONARY.md`. Include only elements
present in this batch — not the whole schema.

### 4. Message-level fields
- `citation` — the source URI the window was fetched from.
- `source_download_timestamp` — fetch time in unix micros.

## Minimal valid record

Ingest accepts a record with just a subject and atoms; resolver info,
strong ids, aliases, and attributes are optional. The smallest useful
shape:

```go
msg := &fetchrecord.FetchMessage{
    Citation:                sourceURL,
    SourceDownloadTimestamp: time.Now().UnixMicro(),
    Records: []*fetchrecord.Record{{
        Source:    reportID,
        Timestamp: pubDate.UnixMicro(),
        Subject:   &fetchrecord.ProtoEntity{Name: name, Flavor: "company"},
        Atoms: []*fetchrecord.Atom{{
            Property:  "annual_revenue",
            Value:     &fetchrecord.Atom_FloatVal{FloatVal: rev},
            Timestamp: pubDate.UnixMicro(),
        }},
    }},
}
path, err := WriteFetchMessage(ctx, store, window, msg)
```

## Timestamps
All timestamps are **unix microseconds** (`t.UnixMicro()`), not seconds or
millis. `Record.timestamp` is the source publication date; `Atom.timestamp`
is when the individual fact holds (often the same — copy it if you have
nothing more specific).

## Out of scope
- **Partitioning.** The canonical fetcher splits very large windows into
  `-p0`, `-p1`, … files. Fido writes one file per window; don't implement
  partitioning unless a window is genuinely too large to hold in memory,
  and flag it if so.
- **Strong-id / resolver tuning.** Populate `ResolverInformation` only if
  the source gives you strong identifiers (e.g. a CIK) or alias/snippet
  context that materially helps resolution. The PoC does not require it.
