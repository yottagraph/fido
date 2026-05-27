---
name: fido
description: Canonical guide for any work in a Fido fetch project — a Cloud Run job template that fetches data from an external source and writes it to a GCS bucket. Read this skill before editing code, schema, design, or infra in this repo.
---

# Fido skill

A **Fido fetch project** is a single-purpose repo: one Cloud Run job
periodically pulls data from one external source and writes the result
to one GCS bucket in a configurable output format.

This skill is the canonical guide for working on any such project. Read
it whenever you touch this repo.

## Mental model

```
external source ──► Cloud Run job ──► GCS bucket (gs://…/output/…)
                          ▲
                          │ Cloud Scheduler (cron-driven trigger)
```

Concretely:

- **Source**: an HTTP/JSON-RPC API, a public dataset, a partner SFTP
  drop, another GCS bucket, etc. Defined in `DESIGN.md`.
- **Job**: a single-shot Go binary (`cmd/fetch`) packaged as a container.
  One invocation does one fetch window, writes its output, persists a
  checkpoint, and exits. Cloud Scheduler triggers the next run.
- **Sink**: a project-owned GCS bucket. Objects are written under
  `output/<YYYY-MM-DD>/<window-key>.<ext>` with the format declared in
  `DESIGN.md` (JSON, NDJSON, CSV, Parquet, etc.).

## Repo layout

```
DESIGN.md            ─ what data this project fetches and why
schema.yaml          ─ structured data model (fields, types, ids)
DATA_DICTIONARY.md   ─ prose definitions for each field in schema.yaml
README.md            ─ short human-facing description
cmd/fetch/           ─ Cloud Run job entrypoint (main.go)
internal/fetch/      ─ source client, output writer, storage helpers
Dockerfile           ─ container image build
.github/workflows/   ─ GitHub Actions CI (build + test + tidy check)
.agents/             ─ agent-facing skills, commands, and this skill
```

A project may add additional packages under `internal/` (e.g. parsers,
enrichment), but the top-level shape stays the same.

**Not in this repo by design**: `cloudbuild.yaml` and `tf/`. Image
build (Cloud Build) and runtime deploy (Cloud Run job + Cloud
Scheduler + GCS bucket + IAM) are owned by the Broadchurch Portal,
which provisions everything imperatively from this repo's `main`
branch when the user clicks the **Deploy Cloud Run job** button in
the cockpit. The Portal-owned pieces are not "outstanding work"
that blocks template authoring — one Deploy click does the full
sweep (bucket → SA → IAM → Cloud Build → Cloud Run job →
Scheduler). Don't add a `cloudbuild.yaml` or `tf/` here — the
Portal will ignore it, and it will silently drift from the real
deployed shape. Don't flag the Portal-owned pieces as remaining
dependencies in build reports — they materialise on Deploy, not
before.

## Sources of truth

1. **`DESIGN.md`** is authoritative for *what* this project does. It
   spells out the source, the cadence, the access credentials, the
   output format, and the bucket layout.
2. **`schema.yaml`** is authoritative for the *shape* of the data:
   flavors / entities, properties, relationships, attributes.
3. **`DATA_DICTIONARY.md`** is authoritative for the *meaning* of each
   field: prose definition, units, example values, edge cases.

If these disagree, `DESIGN.md` wins and the others should be updated.

## Code conventions

- **Language**: Go (1.25+). The template ships a working `go.mod`. A
  project may switch languages, but the layout above (`cmd/`,
  `internal/`, Dockerfile) and the conventions below stay.
- **Single binary per Cloud Run job.** `cmd/fetch/main.go` parses flags,
  builds a `Config`, and calls into a `Run` function in `internal/`.
- **Storage interface.** All writes go through a `Store` abstraction
  with `gs://` and `file://` backends so the same code path runs in
  Cloud Run and locally.
- **Checkpoints.** Persist progress to `checkpoints/checkpoint.json` in
  the output bucket. One Cloud Run job invocation resumes from the
  checkpoint and persists a new one before exiting.
- **Output layout.** `output/<YYYY-MM-DD>/<window-key>.<ext>`. Use the
  earliest record timestamp in the window to pick the date; fall back to
  "now" for empty windows.
- **No upstream lovelace imports.** A Fido project is meant to be
  isolated from any monorepo — `go.mod` should not pull in
  `lovelace-ai.com/...` packages. CI enforces this.
- **Comments**: explain non-obvious intent, not what the code does.

## What the standard build flow does

The starting command is
[`.agents/commands/build_my_fetch.md`](../../commands/build_my_fetch.md).
At a high level it:

1. Reads `DESIGN.md`, `schema.yaml`, and `DATA_DICTIONARY.md`.
2. Customises the template files in `cmd/`, `internal/`, and
   `Dockerfile` so they describe and implement that specific source
   and output. (`cloudbuild.yaml` and `tf/` are NOT in the template
   — the Broadchurch Portal owns image build + runtime deploy and
   does both on the cockpit's **Deploy Cloud Run job** button.)
3. Renames the `github.com/example/fido-fetch` module path in `go.mod`
   and the matching import in `cmd/fetch/main.go`.
4. Self-reviews the result against the checklist below.
5. Pushes to `main` and stops. The Deploy button is what triggers
   the rest of the pipeline; there is nothing the template author
   needs to do to "hand off" beyond pushing.

## Self-review checklist

After any substantive change, walk through this list:

- [ ] `DESIGN.md` names the data source, its access mechanism, the
      output format, the GCS bucket name, and the cadence.
- [ ] `schema.yaml` matches the fields described in `DATA_DICTIONARY.md`,
      and vice versa.
- [ ] The Cloud Run job's flags, defaults, and config struct match what
      `DESIGN.md` says it should do.
- [ ] Object-path templates in the code match the layout documented in
      `DESIGN.md`.
- [ ] `Dockerfile` builds the right `cmd/...` binary and only that one.
- [ ] `go.mod` declares the real module path (not the
      `github.com/example/fido-fetch` placeholder) and the import in
      `cmd/fetch/main.go` matches.
- [ ] `.github/workflows/test.yml` runs `go build ./...`, `go test ./...`,
      and the isolation check.
- [ ] The only template placeholder identifier
      (`github.com/example/fido-fetch`) has been replaced in `go.mod`
      and in the matching import in `cmd/fetch/main.go`. A clean
      grep for `example/fido-fetch` returns nothing.
- [ ] `README.md` describes *this* project in one paragraph.
- [ ] `go build ./...` and `go test ./...` pass locally (or the agent
      reports a clear blocker if they don't).

## When this skill does not apply

- Editing the Fido template itself (i.e. landing changes that affect
  *every* future Fido project rather than a single project) — that is
  template maintenance, not a per-project build. Treat the template as
  read-only and surface any improvements as a separate suggestion.
- Adding a second Cloud Run workload (e.g. an extract Cloud Run service)
  — call that out explicitly; it is outside the template's default shape.
