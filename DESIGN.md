# DESIGN

_This file is the per-project design for one Fido fetch project. Until
it has been customised, this file describes the template defaults and
calls out the sections that need to be filled in._

## Project

- **Name**: _e.g. `acme-fetch-orders`_
- **One-line summary**: _e.g. "Hourly fetch of order events from the
  Acme partner API into `gs://acme-fetch-orders-raw`."_

## Source

- **What it is**: _public API, partner SFTP drop, upstream GCS bucket,
  JSON-RPC node, …_
- **Endpoint or upstream URI**: _`https://...`, `sftp://...`,
  `gs://...`_
- **Auth**: _none, API key (Secret Manager secret name), OAuth client
  credentials, etc._
- **Rate limits / quotas**: _e.g. 5 req/s, 10k requests/day_

## Sink

- **GCS bucket**: _`gs://<project>-<name>-raw`_
- **Object layout**: _`output/<YYYY-MM-DD>/<window-key>.<ext>`_
- **Output format**: _`json`, `ndjson`, `csv`, `parquet`, …_
- **Compression**: _none, `gzip`, `zstd`_
- **Retention**: _e.g. 90 days, lifecycle rule managed in `tf/`_

## Cadence and windowing

- **Trigger**: Cloud Scheduler → Cloud Run job. _Cron expression goes
  here, e.g. `0 * * * *` for hourly._
- **Window definition**: _e.g. "one block range per invocation",
  "last 1h of events", "everything new since the checkpoint"_
- **Checkpointing**: stored at `checkpoints/checkpoint.json` in the
  output bucket. Describe what the checkpoint contains (last block,
  last cursor, last timestamp, …).

## Data model summary

Pointer to the structured model — keep it short here; the source of
truth is `schema.yaml` + `DATA_DICTIONARY.md`.

- **Primary entity**: _e.g. order, address, sensor reading_
- **Key fields**: _e.g. order id, timestamp, amount_
- **Relationships**: _e.g. order belongs_to customer_

## Out of scope (P0)

List anything explicitly *not* in the first cut. Examples:

- Backfill of historical data.
- Multi-region / multi-tenant fanout.
- Per-record metrics or tracing.
- Real-time streaming (this template is batch-on-a-cron).
