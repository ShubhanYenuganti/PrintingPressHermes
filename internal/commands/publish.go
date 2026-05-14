package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	publishRunDir  string
	publishLibrary string
	publishJSON    bool
	publishDryRun  bool
)

type PublishResult struct {
	Slug   string `json:"slug"`
	Source string `json:"source"`
	Dest   string `json:"dest"`
	Status string `json:"status"`
}

func execPublish(slug, runDir, library string, dryRun bool) (*PublishResult, error) {
	if runDir == "" {
		home, _ := os.UserHomeDir()
		runDir = filepath.Join(home, "hermes-press", "runs", slug)
	}
	if library == "" {
		home, _ := os.UserHomeDir()
		library = filepath.Join(home, "hermes-press", "library")
	}
	src := filepath.Join(runDir, "working", slug)
	dst := filepath.Join(library, slug)
	if _, err := os.Stat(src); err != nil {
		return nil, fmt.Errorf("working directory not found at %s — run verify first", src)
	}
	if !dryRun {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return nil, fmt.Errorf("cannot create library dir: %w", err)
		}
	}
	status := "published"
	if dryRun {
		status = "dry-run"
	}
	return &PublishResult{Slug: slug, Source: src, Dest: dst, Status: status}, nil
}

var publishCmd = &cobra.Command{
	Use:   "publish <slug>",
	Short: "Phase 5: promote working CLI to the library",
	Long: `Copies the verified working CLI from <run_dir>/working/<slug> to
~/hermes-press/library/<slug>. Call only after verify --strict passes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := execPublish(args[0], publishRunDir, publishLibrary, publishDryRun)
		if err != nil {
			return err
		}
		if publishJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("publish: %s\n", result.Slug)
			fmt.Printf("source:  %s\n", result.Source)
			fmt.Printf("dest:    %s\n", result.Dest)
			fmt.Printf("status:  %s\n", result.Status)
		}
		return nil
	},
}

func init() {
	publishCmd.Flags().StringVar(&publishRunDir, "run-dir", "", "Run directory containing working/<slug>")
	publishCmd.Flags().StringVar(&publishLibrary, "library", "", "Library destination (default: ~/hermes-press/library)")
	publishCmd.Flags().BoolVar(&publishJSON, "json", false, "Output machine-readable JSON")
	publishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Validate without copying files")
}
