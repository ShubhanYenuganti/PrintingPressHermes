package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "hermes-press",
	Short: "Hermes Printing Press — agent-driven CLI generator for APIs",
	Long: `hermes-press is the Go binary half of the Hermes Printing Press plugin.
Hermes skill files drive it through research, generation, verification, and
publication phases. The binary handles state, scoring, and deterministic checks;
the skill files handle the agentic loop and user interaction.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(researchCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(statusCmd)
}
