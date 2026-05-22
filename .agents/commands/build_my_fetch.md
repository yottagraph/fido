# /build_my_fetch

Start here. This command turns the generic Fido template in this repo
into a customised fetch project for one specific data source.

## Preconditions

Before running this command, the following should have been populated:

- `DESIGN.md` — what data is being fetched, from where, in what format,
  on what cadence, into which GCS bucket.
- `schema.yaml` — the structured data model for the output.
- `DATA_DICTIONARY.md` — prose definitions for each field in `schema.yaml`.

If any of these still look like the un-customised template (placeholder
text, no source named, no bucket named), **stop and ask** before touching
the rest of the repo.

## Instructions for the agent

1. **Load context.**
   - Read [`.agents/skills/fido/SKILL.md`](../skills/fido/SKILL.md) end to
     end. It is the canonical guide for this repo and includes the GCP
     terminology, conventions, and self-review checklist you will use.
   - Read `DESIGN.md`, `schema.yaml`, and `DATA_DICTIONARY.md`.
   - Skim `AGENTS.md` and `README.md` for context.

2. **Take stock of the template.**
   - List the files under `cmd/`, `internal/`, `tf/`, `.github/`, plus
     the top-level `Dockerfile` and `cloudbuild.yaml`.
   - Identify every placeholder identifier (typically `fido-fetch`, the
     module name in `go.mod`, the image name in `cloudbuild.yaml`, the
     bucket name in `tf/`, the service-account names) and decide what
     they should become for this project based on `DESIGN.md`.

3. **Customise the template.** Walk through these areas. Edit each file
   so it describes and implements the specific source named in
   `DESIGN.md`:
   - **Cloud Run job entrypoint** under `cmd/` — flags, defaults,
     wiring of source client and output writer.
   - **Internal packages** under `internal/` — source-specific client,
     parsing/normalising code, output writer that emits the format named
     in `DESIGN.md`.
   - **`schema.yaml` + `DATA_DICTIONARY.md`** — confirm they match each
     other and the fields the code emits. Tighten any vague descriptions.
   - **`Dockerfile`** — make sure it builds the right binary path.
   - **`cloudbuild.yaml`** — image name and Artifact Registry repo
     should reflect this project.
   - **`tf/main.tf`** — bucket name, Cloud Run job name, scheduler
     cadence, IAM bindings, and any source-specific env vars or
     secrets. Remove resources the project does not need.
   - **`.github/workflows/test.yml`** — should still pass for whatever
     this project ships.
   - **`README.md`** — one paragraph describing *this* project, plus a
     "Local quickstart" block that actually works.

4. **Keep the template generic where it makes sense.** Do not invent a
   second Cloud Run workload, an Eventarc trigger, or a downstream
   publish step unless `DESIGN.md` asks for one. The default shape is
   one Cloud Run job + one GCS bucket.

5. **Self-review.** Walk the self-review checklist in the Fido skill,
   line by line. For each item:
   - If it passes, note it (mentally is fine).
   - If it fails, **fix it now** before reporting back.

   In particular, look for:
   - Stray references to a previous project's domain (the template
     evolved from an ERC-20 example; phrases like `erc20`, `USDC`,
     `etherscan`, `eth_address` should not survive).
   - Placeholder identifiers like `fido-fetch` that you forgot to rename.
   - Inconsistencies between `DESIGN.md`, `schema.yaml`,
     `DATA_DICTIONARY.md`, and the code.
   - Imprecise GCP terminology (see the cheat-sheet in the Fido skill).

6. **Verify the build.** Run `go build ./...` and `go test ./...`. If a
   build or test fails, fix it. If a failure depends on external state
   (network, credentials, a fixture that is not present), report the
   blocker clearly rather than papering over it.

7. **Report back.** Summarise:
   - What you changed, grouped by area (entrypoint, internal, schema,
     infra, CI).
   - Anything you intentionally left as a TODO and why.
   - Anything you noticed during self-review that you fixed.
   - Any blocker that prevented `go build`/`go test` from passing.

## What not to do

- Do not regenerate from a `fido-dev` upstream — this repo is edited
  directly now. There is no build step copying files in.
- Do not invent fields not described in `DESIGN.md` or
  `DATA_DICTIONARY.md`.
- Do not commit secrets. Use Cloud Run env vars wired up through
  Secret Manager in `tf/`.
- Do not silently delete files. If a template file is genuinely not
  needed for this project, mention it in the report.
