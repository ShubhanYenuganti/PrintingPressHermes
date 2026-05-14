# Printing Press

Generate ship-ready CLIs for any API. `hermes-press` is the standalone runner,
and the Hermes plugin is an optional thin adapter for agentic use inside Hermes.

Two paths, one workflow:

```
hermes-press (standalone binary)  +  printing-press-hermes (Hermes plugin)
       ↑                                     ↑
 runs research, scaffold,           drives the writing step inside Hermes,
 verify, publish                    delegates to the binary for everything else
```

---

## Install

### 1. Install the binary

Requires Go 1.20 or newer ([go.dev/dl](https://go.dev/dl/)).

```bash
go install github.com/shubhany/printing-press-hermes/cmd/hermes-press@latest
```

Verify:

```bash
hermes-press version
# 1.0.0
```

### Standalone usage, no Hermes required

The binary can run the workflow directly from a normal shell. Typical flow:

```bash
# Start a new run and stop after scaffolding so you can implement the CLI
hermes-press run Notion --until generate --json

# Resume after you have written the CLI in the working directory
hermes-press run Notion --from verify

# Or let it run through all phases when the working tree is already prepared
hermes-press run Notion --auto
```

Useful flags:
- `--spec <path>` or `--url <url>` to feed research input
- `--run-dir <path>` to override the default `~/hermes-press/runs/<slug>`
- `--library <path>` to override the publish destination
- `--json` for machine-readable output

### 2. Install the Hermes plugin

Copy the `plugin/` directory from this repo into your Hermes plugins folder:

```bash
cp -r plugin/ ~/.hermes/plugins/printing-press-hermes/
```

Enable it:

```bash
hermes plugins enable printing-press-hermes
hermes plugins list   # should show printing-press-hermes
```

### 3. Start a session

```bash
hermes chat
```

Inside the chat, load the skill:

```
skill_view("printing-press-hermes:printing-press")
```

Then run it:

```
/printing-press Notion
/printing-press "GitHub API"
/printing-press --url https://stripe.com/docs/api
```

---

## Usage

### Generate a new CLI

```
/printing-press <API name or URL>
```

The skill walks through six phases:

| Phase | What happens |
|-------|-------------|
| Pre-flight | Checks `hermes-press` is on PATH and meets the minimum version |
| 0 — Resolve | Derives the canonical slug, checks for an existing run |
| 1 — Research | Calls `hermes-press research`, writes `research.json`, fills in competitor analysis |
| 2 — Generate | Calls `hermes-press generate`, scaffolds a Go module in `~/hermes-press/runs/<slug>/working/` |
| 3 — Build | Agent writes the full CLI: typed HTTP client, SQLite store with FTS5, all resource commands, transcendence command |
| 4 — Verify | Calls `hermes-press verify --strict`: go vet, `--help` walk, `--version` check |
| 5 — Publish | Calls `hermes-press publish`, promotes to `~/hermes-press/library/<slug>/` |

Output lands in `~/hermes-press/library/<slug>/` — a buildable Go module you own.

### Reprint an existing CLI

Reprints upgrade a previously generated CLI under the current press version while
preserving hand-written customizations.

Load the reprint skill:

```
skill_view("printing-press-hermes:printing-press-reprint")
```

Then:

```
/printing-press-reprint notion
```

### Check status

```bash
hermes-press status
hermes-press status --json
```

---

## Binary reference

```
hermes-press run      <api> [--spec path] [--url url] [--run-dir dir] [--library dir] [--until phase] [--from phase] [--auto] [--json]
hermes-press research  <slug> [--spec path] [--url url] [--out dir] [--json]
hermes-press generate  <slug> [--run-dir dir] [--dry-run] [--json]
hermes-press verify    <slug> [--work-dir dir] [--binary path] [--strict] [--json]
hermes-press publish   <slug> [--run-dir dir] [--library dir] [--dry-run] [--json]
hermes-press status            [--json]
hermes-press version
```

Every command that writes state supports `--json` for machine-readable output and
`--dry-run` (where applicable) to validate without side effects. The skill always
passes `--json` and parses the result before proceeding.

---

## Output layout

```
~/hermes-press/
├── runs/
│   └── <slug>/
│       ├── research.json          # Phase 1 output — source of truth
│       └── working/
│           └── <slug>/            # Generated Go module (Phase 2–3)
└── library/
    └── <slug>/                    # Published CLI (Phase 5)
        ├── cmd/<slug>-cli/
        ├── internal/
        └── go.mod
```

---

## Plugin layout

```
plugin/
├── plugin.yaml                    # Hermes plugin manifest
├── __init__.py                    # Registers skills via ctx.register_skill()
└── skills/
    ├── printing-press/
    │   └── SKILL.md               # Main 6-phase agent loop
    └── printing-press-reprint/
        └── SKILL.md               # Upgrade an existing CLI
```

To add the plugin to any Hermes agent, copy `plugin/` to
`~/.hermes/plugins/printing-press-hermes/` and run
`hermes plugins enable printing-press-hermes`. No other configuration required.

---

## Troubleshooting

**`hermes-press` not found**
The binary and the plugin are installed separately. `go install` puts the binary in
`$(go env GOPATH)/bin` — make sure that's on your `PATH`.

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

**Plugin not loading**
Check that both files are present and the plugin is enabled:

```bash
ls ~/.hermes/plugins/printing-press-hermes/
# plugin.yaml  __init__.py  skills/

hermes plugins list | grep printing-press-hermes
```

If it doesn't appear, run `hermes plugins enable printing-press-hermes` and restart
the session.

**Build fails in Phase 3**
Run `go mod tidy` inside the working directory before `go build`:

```bash
cd ~/hermes-press/runs/<slug>/working/<slug>
go mod tidy && go build ./...
```

**`verify --strict` fails**
Do not skip or work around failed checks — fix them. `go vet` errors are almost
always real bugs. The skill will report exactly which check failed and why.

---

## Development

Build locally:

```bash
go build ./cmd/hermes-press
./hermes-press --help
```

Run tests:

```bash
go test ./...
```

To use a local plugin checkout without reinstalling:

```bash
ln -s "$(pwd)/plugin" ~/.hermes/plugins/printing-press-hermes
hermes plugins enable printing-press-hermes
```
