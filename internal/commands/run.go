package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	runURL     string
	runSpec    string
	runRunDir  string
	runLibrary string
	runJSON    bool
	runDryRun  bool
	runAuto    bool
	runUntil   string
	runFrom    string
)

type RunResult struct {
	Slug    string            `json:"slug"`
	Phases  map[string]string `json:"phases"`
	Status  string            `json:"status"`
	WorkDir string            `json:"work_dir,omitempty"`
	Dest    string            `json:"dest,omitempty"`
}

var phaseOrder = []string{"research", "generate", "build", "verify", "publish"}

func phaseIndex(name string) int {
	for i, p := range phaseOrder {
		if p == name {
			return i
		}
	}
	return -1
}

var runCmd = &cobra.Command{
	Use:   "run <api-name>",
	Short: "Run all phases end-to-end without a Hermes session",
	Long: `Chains all five phases into a single command for standalone use:

  Phase 1 — research   writes research.json
  Phase 2 — generate   scaffolds the Go module
  Phase 3 — build      you implement the CLI (interactive pause, or --auto to skip)
  Phase 4 — verify     go vet, --help, binary checks
  Phase 5 — publish    promotes to ~/hermes-press/library/<slug>

Split the workflow at any phase boundary with --until and --from:

  # Stop after scaffold (then write your code):
  hermes-press run <api> --until generate

  # Resume from verification after writing code:
  hermes-press run <api> --from verify

  # Full non-interactive run (code already in working dir):
  hermes-press run <api> --auto`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		api := args[0]
		slug := strings.ToLower(strings.ReplaceAll(api, " ", "-"))

		phases := map[string]string{
			"research": "skipped",
			"generate": "skipped",
			"build":    "skipped",
			"verify":   "skipped",
			"publish":  "skipped",
		}

		fromIdx := 0
		untilIdx := len(phaseOrder) - 1
		if runFrom != "" {
			if idx := phaseIndex(runFrom); idx >= 0 {
				fromIdx = idx
			}
		}
		if runUntil != "" {
			if idx := phaseIndex(runUntil); idx >= 0 {
				untilIdx = idx
			}
		}
		shouldRun := func(name string) bool {
			idx := phaseIndex(name)
			return idx >= 0 && idx >= fromIdx && idx <= untilIdx
		}

		home, _ := os.UserHomeDir()
		runDir := runRunDir
		if runDir == "" {
			runDir = filepath.Join(home, "hermes-press", "runs", slug)
		}
		library := runLibrary
		if library == "" {
			library = filepath.Join(home, "hermes-press", "library")
		}
		workDir := filepath.Join(runDir, "working", slug)
		dest := filepath.Join(library, slug)

		// Phase 1: Research
		if shouldRun("research") {
			if !runJSON {
				fmt.Printf("[1/5] Research: %s\n", slug)
			}
			if _, err := execResearch(api, runSpec, runURL, runDir); err != nil {
				return fmt.Errorf("phase 1 (research) failed: %w", err)
			}
			phases["research"] = "done"
		}

		// Phase 2: Generate
		if shouldRun("generate") {
			if !runJSON {
				fmt.Printf("[2/5] Generate scaffold: %s\n", workDir)
			}
			if _, err := execGenerate(slug, runDir, runDryRun); err != nil {
				return fmt.Errorf("phase 2 (generate) failed: %w", err)
			}
			phases["generate"] = "done"
		}

		// Phase 3: Build — manual implementation step
		if shouldRun("build") {
			if runDryRun || runAuto {
				phases["build"] = "skipped"
			} else if runJSON {
				// In JSON mode emit a checkpoint and exit so the caller can handle phase 3
				phases["build"] = "awaiting-implementation"
				return printRunResult(RunResult{
					Slug:    slug,
					Phases:  phases,
					Status:  "paused-at-build",
					WorkDir: workDir,
				}, true)
			} else {
				fmt.Printf("\n[3/5] Build — implement your CLI in:\n  %s\n\n", workDir)
				fmt.Printf("Suggested files:\n")
				fmt.Printf("  cmd/%s-cli/main.go   — Cobra root, --version, --json, --no-input\n", slug)
				fmt.Printf("  internal/client/      — typed HTTP client with retry\n")
				fmt.Printf("  internal/store/       — SQLite store, FTS5 search\n\n")
				fmt.Print("Press Enter when the code is ready to verify (Ctrl-C to stop here): ")
				bufio.NewReader(os.Stdin).ReadString('\n')
				phases["build"] = "done"
			}
		}

		// Phase 4: Verify
		if shouldRun("verify") {
			if !runJSON {
				fmt.Printf("[4/5] Verify: %s\n", workDir)
			}
			result, err := execVerify(slug, workDir, "")
			if err != nil {
				return fmt.Errorf("phase 4 (verify) failed: %w", err)
			}
			if !result.Passed {
				phases["verify"] = "failed"
				if runJSON {
					data, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(data))
				} else {
					for _, ch := range result.Checks {
						mark := "✓"
						if !ch.Passed {
							mark = "✗"
						}
						fmt.Printf("  %s %s: %s\n", mark, ch.Name, ch.Detail)
					}
				}
				return fmt.Errorf("verify failed — fix the errors above, then re-run with --from verify")
			}
			phases["verify"] = "done"
		}

		// Phase 5: Publish
		if shouldRun("publish") {
			if !runJSON {
				fmt.Printf("[5/5] Publish: %s\n", dest)
			}
			if _, err := execPublish(slug, runDir, library, runDryRun); err != nil {
				return fmt.Errorf("phase 5 (publish) failed: %w", err)
			}
			phases["publish"] = "done"
		}

		status := "done"
		if runUntil != "" {
			status = "stopped-after-" + runUntil
		}
		return printRunResult(RunResult{
			Slug:    slug,
			Phases:  phases,
			Status:  status,
			WorkDir: workDir,
			Dest:    dest,
		}, runJSON)
	},
}

func printRunResult(r RunResult, asJSON bool) error {
	if asJSON {
		data, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("\nrun: %s (%s)\n", r.Slug, r.Status)
	if r.WorkDir != "" {
		fmt.Printf("work_dir:     %s\n", r.WorkDir)
	}
	if r.Phases["publish"] == "done" && r.Dest != "" {
		fmt.Printf("published to: %s\n", r.Dest)
	}
	return nil
}

func init() {
	runCmd.Flags().StringVar(&runURL, "url", "", "API website URL (passed to research)")
	runCmd.Flags().StringVar(&runSpec, "spec", "", "Path to OpenAPI spec (passed to research)")
	runCmd.Flags().StringVar(&runRunDir, "run-dir", "", "Override run directory (default: ~/hermes-press/runs/<slug>)")
	runCmd.Flags().StringVar(&runLibrary, "library", "", "Override library directory (default: ~/hermes-press/library)")
	runCmd.Flags().BoolVar(&runJSON, "json", false, "Machine-readable JSON output")
	runCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Validate without writing files")
	runCmd.Flags().BoolVar(&runAuto, "auto", false, "Skip Phase 3 interactive prompt (code already in working dir)")
	runCmd.Flags().StringVar(&runUntil, "until", "", "Stop after this phase: research|generate|build|verify|publish")
	runCmd.Flags().StringVar(&runFrom, "from", "", "Start from this phase: research|generate|build|verify|publish")
}
