package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	generateRunDir string
	generateJSON   bool
	generateDryRun bool
)

type GenerateResult struct {
	Slug    string `json:"slug"`
	RunDir  string `json:"run_dir"`
	WorkDir string `json:"work_dir"`
	Status  string `json:"status"`
	Next    string `json:"next"`
}

var generateCmd = &cobra.Command{
	Use:   "generate <slug>",
	Short: "Phase 2: scaffold Go CLI + MCP server from research.json",
	Long: `Reads research.json from the run directory, writes the Go module scaffold
(cmd/, internal/, go.mod, README.md) into <run_dir>/working/<slug>/. The skill
builds this with go build and runs quality gates before Phase 3.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		runDir := generateRunDir
		if runDir == "" {
			home, _ := os.UserHomeDir()
			runDir = filepath.Join(home, "hermes-press", "runs", slug)
		}

		researchPath := filepath.Join(runDir, "research.json")
		if _, err := os.Stat(researchPath); err != nil {
			return fmt.Errorf("research.json not found at %s — run 'hermes-press research %s' first", researchPath, slug)
		}

		workDir := filepath.Join(runDir, "working", slug)
		if !generateDryRun {
			if err := os.MkdirAll(filepath.Join(workDir, "cmd", slug+"-cli"), 0755); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Join(workDir, "internal"), 0755); err != nil {
				return err
			}
			placeholder := "// scaffold: skill will fill this in during Phase 3\npackage main\n\nfunc main() {}\n"
			entryPath := filepath.Join(workDir, "cmd", slug+"-cli", "main.go")
			if err := os.WriteFile(entryPath, []byte(placeholder), 0644); err != nil {
				return err
			}
		}

		result := GenerateResult{
			Slug:    slug,
			RunDir:  runDir,
			WorkDir: workDir,
			Status:  "scaffolded",
			Next:    "skill should build with 'go build ./...' then proceed to Phase 3",
		}

		if generateJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("generate: %s\n", slug)
			fmt.Printf("work_dir: %s\n", workDir)
			fmt.Printf("status:   %s\n", result.Status)
			if generateDryRun {
				fmt.Println("[dry-run: no files written]")
			}
		}
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&generateRunDir, "run-dir", "", "Run directory containing research.json")
	generateCmd.Flags().BoolVar(&generateJSON, "json", false, "Output machine-readable JSON")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Validate without writing files")
}
