package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "hermes-press",
	Short: "Printing Press — agent-driven CLI generator for APIs",
	Long: `hermes-press is a CLI tool that scaffolds, verifies, and publishes
production-ready Go CLIs for any API. It can be used standalone or driven by
the printing-press-hermes Hermes plugin.

Standalone workflow (no Hermes required):
  hermes-press run <api>               — full end-to-end run
  hermes-press run <api> --until generate   — scaffold only (then write code)
  hermes-press run <api> --from verify      — verify + publish after writing code

Individual phases:
  hermes-press research <api>          — phase 1: write research.json
  hermes-press generate <slug>         — phase 2: scaffold Go module
  hermes-press verify   <slug>         — phase 4: quality gates
  hermes-press publish  <slug>         — phase 5: promote to library`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(researchCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(statusCmd)
}
