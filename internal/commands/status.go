package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	statusJSON bool
)

type StatusResult struct {
	BinaryVersion string   `json:"binary_version"`
	LibraryDir    string   `json:"library_dir"`
	RunsDir       string   `json:"runs_dir"`
	Published     []string `json:"published"`
	ActiveRuns    []string `json:"active_runs"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show library and active runs",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		libraryDir := filepath.Join(home, "hermes-press", "library")
		runsDir := filepath.Join(home, "hermes-press", "runs")

		published := listDirEntries(libraryDir)
		activeRuns := listDirEntries(runsDir)

		result := StatusResult{
			BinaryVersion: Version,
			LibraryDir:    libraryDir,
			RunsDir:       runsDir,
			Published:     published,
			ActiveRuns:    activeRuns,
		}

		if statusJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("hermes-press v%s\n\n", Version)
			fmt.Printf("library:     %s (%d)\n", libraryDir, len(published))
			for _, p := range published {
				fmt.Printf("  - %s\n", p)
			}
			fmt.Printf("active runs: %s (%d)\n", runsDir, len(activeRuns))
			for _, r := range activeRuns {
				fmt.Printf("  - %s\n", r)
			}
		}
		return nil
	},
}

func listDirEntries(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output machine-readable JSON")
}
