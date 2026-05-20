# AGENTS.md

This is your **Fido fetch project** — a repo for fetching data from a specified source, and writing it to a GCS bucket.

**Read first:** [`.agents/skills/fido/SKILL.md`](.agents/skills/fido/SKILL.md)
— the Fido skill indexes every topic and is the canonical entry point for any
work in this repo. (TODO)

**Source of truth:** `DESIGN.md` — describes the data source you're onboarding
(name, URL or GCS bucket, description).

** Data Model**
- schema.yaml
- DATA_DICTIONARY.md

## Workflow (sketch)

1. Read data model files and `DESIGN.md` to understand the data source.
2. Modify the template files to apply to this specific data source.
3. Run fetcheval/recordeval to confirm that everything is working (TODO, doesn't exist yet)