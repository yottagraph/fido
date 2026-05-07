# DESIGN

Fido is a template repo, following the same structure as aether-dev. In the broadchurch portal, when a user clicks "New Fetch Project", we create a new fido project for them. Like aether-dev, commands and skills are created in their respective directories, packages into an npm package, and propagated to fido tenants.

## Skills

### fido

Similar to the aether skill, the fido skill provides general guidance for interacting with the repository

### fetch-onboarding

This skill provides the actual step-by-step instructions for an agent onboarding a new dataset.

### data-model-skill

This is the same skill imported and incorporated into aether-dev, and it should be imported and incorporated into fido-dev in the same way.

## Commands

None yet, but we keep space for them.