# Agent Notes

This repository is primarily a GitHub Actions harness for producing Emacs.app
nightly and release builds. Treat `.github/workflows/` as the active production
surface; the rest of the repository is mostly public-facing metadata and legacy
content.

## Validation

- Run `mise run check` before handing off workflow changes.
- Run `mise run pin:actions` after changing `uses:` references.
- Run `mise run pin:actions:audit` with `GITHUB_TOKEN` set to verify the
  configured minimum action release age.
- `pinact` is configured with a 3-day minimum release age in `.pinact.yaml`.

## Workflow Constraints

- Keep workflow inputs out of inline shell template expansion. Pass inputs
  through `env:` and consume them as shell variables.
- Keep GitHub Actions permissions explicit and scoped to the job's needs.
- Pass reusable workflow secrets by name; avoid `secrets: inherit`.
- Pin external actions to full commit SHAs with version comments via `pinact`.
