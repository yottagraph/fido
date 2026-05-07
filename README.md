# Fido Fetch Project

This repo defines a single Lovelace data source onboarding — a "fetch project"
created from the Broadchurch portal.

## Where to start

- **`DESIGN.md`** — what data source this project is onboarding (filled in
  from the form in the Broadchurch portal).
- **`AGENTS.md`** — orientation for Cursor and Claude Code agents working in
  this repo. Points at the Fido skill.
- **`.agents/skills/fido/SKILL.md`** — the Fido skill index.
- **`.agents/skills/fetch-onboarding/SKILL.md`** — step-by-step fetch
  onboarding workflow.

## Lifecycle

1. Project created in the Broadchurch portal → this repo is provisioned with
   a `DESIGN.md` and the latest `@yottagraph-app/fido-instructions` package
   contents.
2. An agent walks through the fetch-onboarding workflow defined here, in
   collaboration with you.
3. The agent produces a fetch definition + Cloud Run job source.
4. Cloud Run runs the fetch and POSTs back to the Broadchurch portal when
   the data is ready.
5. The new data source appears under "Available Data Sources" in the portal.
