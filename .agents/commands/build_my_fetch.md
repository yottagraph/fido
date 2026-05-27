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
   - List the files under `cmd/`, `internal/`, `.github/`, plus the
     top-level `Dockerfile` and `go.mod`.
   - Identify every placeholder identifier (typically the module name
     in `go.mod`, the import path inside `cmd/fetch/main.go`, and any
     `fido-fetch` strings the code uses for object naming) and decide
     what they should become for this project based on `DESIGN.md`.
   - **You do NOT customise `cloudbuild.yaml` or `tf/`** — neither
     file exists in this template, by design. Image build (Cloud
     Build) and runtime deploy (Cloud Run job + Cloud Scheduler +
     GCS bucket + IAM) are owned by the Broadchurch Portal, which
     provisions everything imperatively from this repo's `main`
     branch when the user clicks the **Deploy Cloud Run job**
     button in the cockpit. This is *not* outstanding work that
     blocks you — you don't have to wait for it, prepare for it,
     or flag it as pending in your report. One Deploy click does
     the full sweep: GCS bucket → service account → IAM → Cloud
     Build (image push) → Cloud Run job → Cloud Scheduler. Your
     job is finished when `main` builds + tests cleanly.

3. **Customise the template.** Walk through these areas. Edit each file
   so it describes and implements the specific source named in
   `DESIGN.md`:
   - **`go.mod`** — replace the `github.com/example/fido-fetch` module
     path with the real path, and update the matching import in
     `cmd/fetch/main.go`. Run `go mod tidy` after.
   - **Cloud Run job entrypoint** under `cmd/` — flags, defaults,
     wiring of source client and output writer.
   - **Internal packages** under `internal/` — source-specific client,
     parsing/normalising code, output writer that emits the format named
     in `DESIGN.md`. **Leave `internal/fetch/storage.go` alone** unless
     `DESIGN.md` calls for a non-GCS sink; the template ships with a
     working `gs://` + `file://` backend.
   - **`schema.yaml` + `DATA_DICTIONARY.md`** — confirm they match each
     other and the fields the code emits. Tighten any vague descriptions.
   - **`Dockerfile`** — make sure it builds the right binary path.
   - **`.github/workflows/test.yml`** — should still pass for whatever
     this project ships.
   - **`README.md`** — one paragraph describing *this* project, plus a
     "Local quickstart" block that actually works.

4. **Keep the template generic where it makes sense.** Do not invent a
   second Cloud Run workload, an Eventarc trigger, or a downstream
   publish step unless `DESIGN.md` asks for one. The default shape is
   one Cloud Run job + one GCS bucket, both provisioned by the
   Broadchurch Portal.

5. **Self-review.** Walk the self-review checklist in the Fido skill,
   line by line. For each item:
   - If it passes, note it (mentally is fine).
   - If it fails, **fix it now** before reporting back.

   In particular, look for:
   - Stray references to a previous project's domain (the template
     evolved from an ERC-20 example; phrases like `erc20`, `USDC`,
     `etherscan`, `eth_address` should not survive).
   - Placeholder identifiers like `github.com/example/fido-fetch` in
     `go.mod` or imports that you forgot to rename.
   - Inconsistencies between `DESIGN.md`, `schema.yaml`,
     `DATA_DICTIONARY.md`, and the code.
   - Imprecise GCP terminology (see the cheat-sheet in the Fido skill).

6. **Verify the build.** Run `go build ./...` and `go test ./...`. If a
   build or test fails, fix it. If a failure depends on external state
   (network, credentials, a fixture that is not present), report the
   blocker clearly rather than papering over it.

7. **Push the result directly to `main`.** Once the build is green and
   you've self-reviewed against `DESIGN.md`:
   - Commit your changes on `main` and run `git push origin main`.
   - **Do NOT** create a feature branch.
   - **Do NOT** open a pull request or run `gh pr create`.
   - Pushing to `main` is the entire handoff. The Broadchurch Portal
     does the rest when the user clicks **Deploy Cloud Run job** in
     the cockpit: provision the GCS bucket, create the per-job
     service account, bind IAM, run Cloud Build to push the image,
     upsert the Cloud Run job, and bind the Cloud Scheduler trigger.
     Nothing ships from a push alone, and you don't need to do
     anything to "trigger" the Portal — the Deploy button is what
     does it, not your push.

8. **Report back.** Summarize:
   - What you changed, grouped by area (entrypoint, internal, schema,
     CI).
   - Anything you intentionally left as a TODO and why.
   - Anything you noticed during self-review that you fixed.
   - Any blocker that prevented `go build`/`go test` from passing.
   - Confirm you pushed to `main` (and did not open a PR).
   - **Do not list Portal-owned pieces (GCS bucket, Cloud Scheduler,
     Cloud Build, Cloud Run job, IAM, service account) as
     outstanding work or external dependencies.** They are not
     blockers — they materialise on Deploy. If you find yourself
     writing "the Portal still needs to provision X before Deploy
     works," that is wrong; reread step 2 above.

## What not to do

- Do not regenerate from a `fido-dev` upstream — this repo is edited
  directly now. There is no build step copying files in.
- Do not invent fields not described in `DESIGN.md` or
  `DATA_DICTIONARY.md`.
- Do not commit secrets. Per-data-source secrets land in the
  per-tenant Secret Manager via the cockpit's Secrets panel, and the
  Portal binds them as env vars on the Cloud Run job at Deploy time —
  the template never sees the raw values.
- Do not silently delete files. If a template file is genuinely not
  needed for this project, mention it in the report.
- Do not create a feature branch or pull request for the result. Push
  the finished work straight to `main`; the Broadchurch Portal handles
  deploys on demand from there.
