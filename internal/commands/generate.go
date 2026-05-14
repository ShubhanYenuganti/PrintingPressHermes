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

func execGenerate(slug, runDir string, dryRun bool) (*GenerateResult, error) {
	if runDir == "" {
		home, _ := os.UserHomeDir()
		runDir = filepath.Join(home, "hermes-press", "runs", slug)
	}
	researchPath := filepath.Join(runDir, "research.json")
	if _, err := os.Stat(researchPath); err != nil {
		return nil, fmt.Errorf("research.json not found at %s — run 'hermes-press research %s' first", researchPath, slug)
	}
	workDir := filepath.Join(runDir, "working", slug)
	if !dryRun {
		if err := os.MkdirAll(filepath.Join(workDir, "cmd", slug+"-cli"), 0755); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Join(workDir, "internal"), 0755); err != nil {
			return nil, err
		}
		placeholder := "// scaffold: implement your CLI here\npackage main\n\nfunc main() {}\n"
		entryPath := filepath.Join(workDir, "cmd", slug+"-cli", "main.go")
		if err := os.WriteFile(entryPath, []byte(placeholder), 0644); err != nil {
			return nil, err
		}
	}
	return &GenerateResult{
		Slug:    slug,
		RunDir:  runDir,
		WorkDir: workDir,
		Status:  "scaffolded",
		Next:    "implement the CLI in the working directory, then run verify",
	}, nil
}

var generateCmd = &cobra.Command{
	Use:   "generate <slug>",
	Short: "Phase 2: scaffold Go CLI module from research.json",
	Long: `Reads research.json from the run directory, writes the Go module scaffold
(cmd/, internal/, go.mod) into <run_dir>/working/<slug>/. Implement the CLI there,
then run verify.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := execGenerate(args[0], generateRunDir, generateDryRun)
		if err != nil {
			return err
		}
		if generateJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("generate: %s\n", result.Slug)
			fmt.Printf("work_dir: %s\n", result.WorkDir)
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
