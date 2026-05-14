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

Hermes adapter for the Printing Press CLI generator. The heavy lifting lives in
the `hermes-press` standalone binary; this skill drives the AI writing step
(Phase 3) and delegates everything else to the binary.

Sibling: `skill_view("printing-press-hermes:printing-press-reprint")` — upgrade an existing CLI.

## When to use

User says any of:
- "print a CLI for \<API\>"
- `/printing-press <API or URL>`
- "generate a CLI for \<API\>"

## Pre-flight: verify binary

```bash
HERMES_PRESS="${HERMES_PRESS_BINARY:-hermes-press}"
if ! command -v "$HERMES_PRESS" >/dev/null 2>&1; then
  echo "[setup-error] hermes-press not found."
  echo "Install: go install github.com/shubhany/printing-press-hermes/cmd/hermes-press@latest"
  exit 1
fi
"$HERMES_PRESS" version
```

If the binary check fails, stop and show the install instructions. Do not proceed.

## Phases 1–2: Research and scaffold (binary)

Run phases 1 and 2 via the standalone binary. This requires no Hermes state.

```bash
API_SLUG="$(echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')"
"$HERMES_PRESS" run "$API_SLUG" --until generate --json
# With spec:  --spec path/to/openapi.yaml
# With URL:   --url https://example.com/api-docs
```

Parse the JSON. Check `status` is `stopped-after-generate`. Read `work_dir`.
If a prior run exists for this slug, ask the user: resume or start fresh?

**Research questions to answer before Phase 3 (do your own web research here):**

1. What is the API's canonical name and authentication method?
2. Who are the 2–3 top CLI competitors? What do they do well?
3. What compound command would only make sense with local data (cross-resource
   insight no raw API call returns in one shot)?
4. What is the natural SQLite schema for this API's primary resources?

## Phase 3: Build the CLI (agent writes code)

Implement the full CLI in `$WORK_DIR`. Checklist:

- [ ] `cmd/<slug>-cli/main.go` — Cobra root, `--version`, `--json`, `--no-input`, `--compact`
- [ ] `internal/client/` — typed HTTP client with retry and rate-limit handling
- [ ] `internal/store/` — SQLite store using `modernc.org/sqlite`, FTS5 full-text search
- [ ] Resource commands: `sync`, `list`, `get`, `search` for each primary resource
- [ ] Transcendence command: the compound insight from Phase 1 research
- [ ] `doctor` command — checks auth, connectivity, DB integrity
- [ ] `go.mod` / `go.sum` — tidy after writing

Build until clean:

```bash
cd "$WORK_DIR" && go build ./... 2>&1
go vet ./... 2>&1
```

## Phases 4–5: Verify and publish (binary)

```bash
"$HERMES_PRESS" run "$API_SLUG" --from verify --json
```

Parse the JSON. If `phases.verify` is `failed`, fix each failed check and re-run.
Confirm `phases.publish` is `done` and `dest` path exists.

## Completion

Tell the user:
- Where the CLI binary is (`~/hermes-press/library/<slug>/`)
- The three most compelling commands
- The transcendence command and what insight it unlocks
- How to add the binary to PATH

## Pitfalls

- **Binary not found**: install separately with `go install`. Show instructions.
- **Phase 3 build fails**: run `go mod tidy` before `go build`.
- **verify --strict fails on go-vet**: fix the vet errors — do not skip them.
- **Wrong slug**: lower-kebab only. "Notion API" → `notion-api`.

## Verification

Complete when all are true:
- `hermes-press verify --strict` exits 0
- `hermes-press publish` reports `status: published`
- `$HOME/hermes-press/library/<slug>/` exists and contains a buildable Go module
