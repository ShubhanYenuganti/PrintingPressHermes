---
name: printing-press
description: Generate a ship-ready CLI for an API — research, scaffold, verify, publish.
version: 1.0.0
platforms: [macos, linux]
metadata:
  hermes:
    tags: [go, cli, generation, api]
    category: devops
    requires_toolsets: [terminal]
    config:
      - key: printing_press.binary
        description: "Name or path of the hermes-press binary"
        default: "hermes-press"
        prompt: "Where is hermes-press installed? (leave blank to use PATH)"
---

# printing-press

Generate the best useful CLI for an API. The skill drives the `hermes-press` Go
binary through a fixed phase loop; the binary handles state and deterministic
checks while the skill handles research, writing, and decisions.

## Sibling skills

- `skill_view("printing-press-hermes:printing-press-reprint")` — reprint an existing CLI under a newer version of the press.

## When to use

User says any of:
- "print a CLI for <API>"
- `/printing-press <API or URL>`
- "generate a CLI for <API>"

## Pre-flight: verify binary

Before Phase 0, verify the binary is installed and meets the minimum version.

```bash
# Resolve binary — prefer explicit env var, fall back to PATH
HERMES_PRESS="${HERMES_PRESS_BINARY:-hermes-press}"

if ! command -v "$HERMES_PRESS" >/dev/null 2>&1; then
  echo "[setup-error] hermes-press binary not found."
  echo ""
  if command -v go >/dev/null 2>&1; then
    echo "Install it:"
    echo "  go install github.com/shubhany/printing-press-hermes/cmd/hermes-press@latest"
  else
    echo "Go is also missing. Install Go from https://go.dev/dl/, then:"
    echo "  go install github.com/shubhany/printing-press-hermes/cmd/hermes-press@latest"
  fi
  echo ""
  echo "Then re-run /printing-press."
  exit 1
fi

BINARY_VERSION="$("$HERMES_PRESS" version 2>/dev/null)"
echo "hermes-press $BINARY_VERSION ready"
```

If the binary check fails, stop and show the user the install instructions. Do not
proceed to Phase 0.

## Phase 0 — Resolve

Determine the canonical API slug and run directory.

```bash
API_SLUG="$(echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')"
RUN_DIR="$HOME/hermes-press/runs/$API_SLUG"
echo "slug: $API_SLUG"
echo "run_dir: $RUN_DIR"
```

Check `hermes-press status --json` for an existing run:

```bash
"$HERMES_PRESS" status --json
```

If a prior run exists for this slug, ask the user: resume it or start fresh?

## Phase 1 — Research

Call `hermes-press research` to create the run directory and write `research.json`.
Pass `--spec` or `--url` if the user provided one; otherwise let the binary infer.

```bash
"$HERMES_PRESS" research "$API_SLUG" --json
# With spec:  --spec path/to/openapi.yaml
# With URL:   --url https://example.com/api-docs
```

Read `$RUN_DIR/research.json`. This is the source of truth for all subsequent phases.
If `status` is not `ready`, stop and report the error.

**Research questions to answer before Phase 2 (do your own web research here):**

1. What is the API's canonical name and authentication method?
2. Who are the 2–3 top CLI competitors? What features do they have?
3. What is a compelling compound command that only makes sense with local data
   (e.g. a cross-resource insight that no raw API call can return in one shot)?
4. What is the natural SQLite schema for this API's primary resources?

Write answers as notes into `research.json` before proceeding. The binary uses them
in Phase 2.

## Phase 2 — Generate scaffold

```bash
"$HERMES_PRESS" generate "$API_SLUG" --run-dir "$RUN_DIR" --json
```

Verify output: `work_dir` in the JSON must exist and contain `cmd/<slug>-cli/main.go`.

```bash
ls "$RUN_DIR/working/$API_SLUG/cmd/$API_SLUG-cli/"
```

If the scaffold is missing, re-run generate. Do not proceed to Phase 3 with a broken scaffold.

## Phase 3 — Build the CLI

This is the main writing phase. You will implement the full CLI in
`$RUN_DIR/working/$API_SLUG/`. Follow this checklist:

- [ ] `cmd/<slug>-cli/main.go` — Cobra root, version flag, `--json`, `--no-input`, `--compact`
- [ ] `internal/client/` — typed HTTP client with retry and rate-limit handling
- [ ] `internal/store/` — SQLite store using `modernc.org/sqlite`, FTS5 full-text search
- [ ] Resource commands: `sync`, `list`, `get`, `search` for each primary resource
- [ ] Transcendence command: the compound insight from Phase 1 research
- [ ] `doctor` command — checks auth, connectivity, DB integrity
- [ ] `go.mod` / `go.sum` — tidy after writing

Build and fix until clean:

```bash
cd "$RUN_DIR/working/$API_SLUG" && go build ./... 2>&1
go vet ./... 2>&1
```

## Phase 4 — Verify

```bash
"$HERMES_PRESS" verify "$API_SLUG" \
  --work-dir "$RUN_DIR/working/$API_SLUG" \
  --strict --json
```

Parse the JSON. If `passed` is `false`, fix each failed check inline and re-run.
Do not proceed to Phase 5 until `passed` is `true`.

Quick smoke checks to run manually alongside verify:

```bash
BINARY="$RUN_DIR/working/$API_SLUG/$API_SLUG-cli"
go build -o "$BINARY" "$RUN_DIR/working/$API_SLUG/cmd/$API_SLUG-cli"
"$BINARY" --help
"$BINARY" --version
"$BINARY" list --json 2>&1 | head -5
"$BINARY" doctor
```

## Phase 5 — Publish

Only call publish when Phase 4 verify has `passed: true`.

```bash
"$HERMES_PRESS" publish "$API_SLUG" --run-dir "$RUN_DIR" --json
```

Confirm `status: published` and `dest` path in the JSON output.

## Completion

Tell the user:
- Where the CLI binary is (`~/hermes-press/library/<slug>/`)
- The three most compelling commands
- The transcendence command and what insight it unlocks
- How to add the binary to PATH

## Pitfalls

- **Binary not found**: the Go binary must be installed separately from the plugin. Show install instructions.
- **Phase 3 build fails**: run `go mod tidy` before `go build`. Missing `go.sum` is the most common cause.
- **verify --strict fails on go-vet**: do not skip — fix the vet errors. They are almost always real bugs.
- **Wrong slug**: slugs are lower-kebab. "Notion API" → `notion-api`. Match exactly across all phases.
- **Reusing a stale run**: if `research.json` exists from a prior run, check the timestamp. If it's old, re-run Phase 1.

## Verification

The run is complete when all are true:
- `hermes-press verify --strict` exits 0
- `hermes-press publish` reports `status: published`
- `$HOME/hermes-press/library/<slug>/` exists and contains a buildable Go module
