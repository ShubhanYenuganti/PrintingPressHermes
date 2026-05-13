package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print binary version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
