# Fix and Re-audit Plan: fix/audit-hardening

## Goal
Repair the confirmed missing-icon build blocker, harden path-containment checks against ancestor junctions, eliminate stale asynchronous UI state updates, implement the inactive About control, and improve keyboard/reduced-motion accessibility.

## Task List
- [ ] Restore a canonical, square SVG icon at `ui/icon.svg`; ensure React and the Windows icon-generation workflow use the same asset.
- [ ] Add a regression test for a destination reached through an ancestor symlink/junction and compare resolved filesystem paths during create-request validation.
- [ ] Make target picker, destination picker, destination validation, navigation, and submit state resistant to stale async completion; add visible operation errors.
- [ ] Implement an accessible About dialog and keyboard Escape behavior.
- [ ] Add visible focus indicators and reduced-motion CSS.
- [ ] Re-audit changed files and inspect current CI status; run tests/build where the environment permits.

## Acceptance Criteria
- Frontend's imported SVG exists and is a square icon suitable for rasterization.
- A destination resolving inside a directory target is rejected even when an ancestor path is a symlink/junction.
- Late async results cannot move the workflow or overwrite newer state.
- About control opens a dismissible, keyboard-accessible dialog.
- Keyboard focus remains visible and motion is reduced when requested by OS settings.
- No claims of passing checks without actual CI or local test evidence.

## Risks
- Windows junction resolution behavior must be confirmed by Windows CI; symlink regression test skips when the host cannot create symlinks.
- GitHub contents API edits are committed directly to the feature branch, so each logical change should be reviewed in its resulting commit.
