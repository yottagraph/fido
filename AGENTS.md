# AGENTS.md

This repo is a **Fido fetch project** — a template for a Cloud Run job that
fetches data from one external source and writes it to a Google Cloud
Storage (GCS) bucket in a configurable output format.

Each project that starts from this template is meant to be customised in
place: the per-project `DESIGN.md`, data model, and code stubs are filled
out so the resulting repo describes and implements exactly one fetch.

## Read first

- [`.agents/skills/fido/SKILL.md`](.agents/skills/fido/SKILL.md) — the Fido
  skill is the canonical guide for any work in this repo. Read it before
  touching code or docs.
- [`.agents/commands/build_my_fetch.md`](.agents/commands/build_my_fetch.md)
  — the standard end-to-end starting command. Invoke this when the user
  asks you to build out the template for their specific data source.

## Sources of truth

- `DESIGN.md` — describes the data source this project targets (name,
  endpoint or upstream bucket, access details, cadence, output format).
- `schema.yaml` — the structured data model.
- `DATA_DICTIONARY.md` — prose definitions for the fields in `schema.yaml`.

When `DESIGN.md`, `schema.yaml`, and `DATA_DICTIONARY.md` disagree, treat
`DESIGN.md` as authoritative and update the others to match.

## Workflow (sketch)

1. Read `DESIGN.md`, `schema.yaml`, and `DATA_DICTIONARY.md` to understand
   the data source and target shape.
2. Customise the template files (Cloud Run job entrypoint, output writer,
   Dockerfile, Cloud Build config, Terraform, CI) to match.
3. Self-review the result against the design and skill checklists.
4. (Future) Run fetcheval / recordeval to confirm the result behaves as
   documented. _These harnesses do not exist yet — leave a TODO if a
   check would belong here._

## Terminology

This template lives on GCP. Use these exact terms:

- **Cloud Run job** — the kind of workload this template ships (one
  invocation runs to completion and exits; Cloud Scheduler triggers the
  next run). Not "Cloud Run instance" and not "Cloud Run function".
- **GCS bucket** — Google Cloud Storage bucket. Object paths are written
  with `gs://<bucket>/<object>` URIs.
- **Artifact Registry** — where built container images live.
- **Cloud Build** — the CI service that builds and pushes those images.
- **Cloud Scheduler** — what triggers the Cloud Run job on a cadence.

## Provenance

This repo was previously regenerated from a `fido-dev` upstream. That is
no longer the case — edits should land here directly. There is no build
step copying files in from elsewhere.
