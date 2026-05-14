package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	researchSpec   string
	researchURL    string
	researchOutDir string
	researchJSON   bool
)

type ResearchResult struct {
	API        string   `json:"api"`
	Slug       string   `json:"slug"`
	OutDir     string   `json:"out_dir"`
	SpecSource string   `json:"spec_source,omitempty"`
	URL        string   `json:"url,omitempty"`
	Timestamp  string   `json:"timestamp"`
	Status     string   `json:"status"`
	Notes      []string `json:"notes,omitempty"`
}

func execResearch(api, spec, url, outDir string) (*ResearchResult, error) {
	slug := strings.ToLower(strings.ReplaceAll(api, " ", "-"))
	if outDir == "" {
		home, _ := os.UserHomeDir()
		outDir = filepath.Join(home, "hermes-press", "runs", slug)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create output dir: %w", err)
	}
	result := &ResearchResult{
		API:       api,
		Slug:      slug,
		OutDir:    outDir,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    "ready",
		Notes:     []string{"research.json written — proceed to generate"},
	}
	if spec != "" {
		result.SpecSource = spec
	}
	if url != "" {
		result.URL = url
	}
	researchPath := filepath.Join(outDir, "research.json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(researchPath, data, 0644); err != nil {
		return nil, fmt.Errorf("cannot write research.json: %w", err)
	}
	return result, nil
}

var researchCmd = &cobra.Command{
	Use:   "research <api-name>",
	Short: "Phase 1: resolve API identity and write research.json",
	Long: `Resolves the API name to a canonical slug, detects or fetches the OpenAPI spec,
and writes research.json into the run directory. Run before generate.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := execResearch(args[0], researchSpec, researchURL, researchOutDir)
		if err != nil {
			return err
		}
		if researchJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("research: %s -> %s\n", result.Slug, filepath.Join(result.OutDir, "research.json"))
			fmt.Printf("status:   %s\n", result.Status)
		}
		return nil
	},
}

func init() {
	researchCmd.Flags().StringVar(&researchSpec, "spec", "", "Path or URL to OpenAPI spec")
	researchCmd.Flags().StringVar(&researchURL, "url", "", "API website URL (browser-sniff fallback)")
	researchCmd.Flags().StringVar(&researchOutDir, "out", "", "Output directory (default: ~/hermes-press/runs/<slug>)")
	researchCmd.Flags().BoolVar(&researchJSON, "json", false, "Output machine-readable JSON")
}
