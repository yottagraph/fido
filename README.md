# Fido fetch project

Fido plays fetch.

This repo is a template for a single-purpose **Cloud Run job** that pulls
data from one external source and writes the result to a **Google Cloud
Storage (GCS) bucket** in a configurable output format. Cloud Scheduler
triggers the job on a cadence; each invocation does one fetch window and
exits.

## Layout

```
DESIGN.md            ─ what this specific project fetches and why
schema.yaml          ─ structured data model
DATA_DICTIONARY.md   ─ prose definitions for fields in schema.yaml
cmd/fetch/           ─ Cloud Run job entrypoint
internal/fetch/      ─ source client, output writer, storage abstraction
Dockerfile           ─ container image build
cloudbuild.yaml      ─ Cloud Build pipeline → Artifact Registry
tf/                  ─ Terraform: GCS bucket, Cloud Run job, scheduler, IAM
.github/workflows/   ─ build + test CI
.agents/             ─ agent skills and commands (start at build_my_fetch)
```

## Local quickstart

```sh
go build ./...
go test ./...
```

To exercise the job against the local-file output backend:

```sh
go run ./cmd/fetch \
    --source-url=https://example.invalid/api \
    --output=file:///tmp/fido-output \
    --window=$(date +%FT%H-%M-%SZ)
```

The default `internal/fetch/Run` is a no-op placeholder — replace it
with the source-specific fetch logic for this project before deploying.

## Customising the template

Start an agent with the `/build_my_fetch` command (defined in
[`.agents/commands/build_my_fetch.md`](.agents/commands/build_my_fetch.md)).
That command walks the agent through reading `DESIGN.md`, `schema.yaml`,
and `DATA_DICTIONARY.md`, customising the template files, and
self-reviewing the result.

For background, see [`AGENTS.md`](AGENTS.md) and the Fido skill at
[`.agents/skills/fido/SKILL.md`](.agents/skills/fido/SKILL.md).
