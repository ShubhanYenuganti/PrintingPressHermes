---
name: printing-press-reprint
description: Reprint an existing CLI under the latest press version — upgrade without losing customizations.
version: 1.0.0
platforms: [macos, linux]
metadata:
  hermes:
    tags: [go, cli, generation, upgrade]
    category: devops
    requires_toolsets: [terminal]
---

# printing-press-reprint

Hermes adapter for reprinting an existing CLI. Delegates all binary operations
to the `hermes-press` standalone binary; this skill handles snapshotting
customizations and the reapply merge step.

Sibling: `skill_view("printing-press-hermes:printing-press")` — generate a brand-new CLI.

## When to use

User says any of:
- "reprint \<slug\>"
- `/printing-press-reprint <slug>`
- "upgrade the \<slug\> CLI to the latest press"

## Pre-flight

```bash
HERMES_PRESS="${HERMES_PRESS_BINARY:-hermes-press}"
if ! command -v "$HERMES_PRESS" >/dev/null 2>&1; then
  echo "[setup-error] hermes-press not found — install it first."
  exit 1
fi
"$HERMES_PRESS" version
```

## Step 1 — Locate existing CLI

```bash
API_SLUG="$1"
LIBRARY_DIR="$HOME/hermes-press/library/$API_SLUG"
ls "$LIBRARY_DIR" 2>/dev/null || echo "[not-found] $LIBRARY_DIR does not exist"
```

If missing, ask: print fresh with `skill_view("printing-press-hermes:printing-press")` instead?

## Step 2 — Snapshot customizations

Before regenerating, identify hand-written business logic to preserve:

```bash
find "$LIBRARY_DIR" -name "*.go" | while read f; do
  wc -l "$f"
done
```

List files with custom logic and confirm with the user before continuing.

## Steps 3–4: Re-research and regenerate scaffold (binary)

```bash
"$HERMES_PRESS" run "$API_SLUG" --until generate --json
```

## Step 5 — Reapply customizations

Re-apply the customizations from Step 2 to the new scaffold, file by file.

## Step 6 — Verify and publish (binary)

```bash
"$HERMES_PRESS" run "$API_SLUG" --from verify --json
```

Confirm `phases.publish` is `done`.

## Pitfalls

- **Overwriting custom logic**: always snapshot before regenerating (Step 2).
- **Stale research.json**: the reprint always re-runs Phase 1. Pass an existing
  `research.json` path and use `--from generate` to skip Phase 1 if desired.
