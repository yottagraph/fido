# AGENTS.md

This is your **Fido fetch project** — a workspace for defining and deploying
a new data fetch source onto the Lovelace data platform.

**Read first:** [`.agents/skills/fido/SKILL.md`](.agents/skills/fido/SKILL.md)
— the Fido skill indexes every topic and is the canonical entry point for any
work in this repo.

**Source of truth:** `DESIGN.md` — describes the data source you're onboarding
(name, URL or GCS bucket, description). It was generated from the form you
filled out in the Broadchurch portal.

## Workflow (sketch)

1. Read `DESIGN.md` to understand the data source.
2. Read [`.agents/skills/fido/SKILL.md`](.agents/skills/fido/SKILL.md) to
   orient yourself in the repo.
3. Read
   [`.agents/skills/fetch-onboarding/SKILL.md`](.agents/skills/fetch-onboarding/SKILL.md)
   for the step-by-step onboarding process.

## Updating Instructions

Skills and commands install from the `@yottagraph-app/fido-instructions` npm
package. To pull in the latest version, run:

```
/update_instructions
```

(Command not yet implemented; placeholder.)

## Committing

Push directly to `main`. The Broadchurch portal will pick up status updates
when the fetch is wired up to a Cloud Run job.
