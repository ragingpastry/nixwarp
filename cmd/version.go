package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "version",
	Long:    "Prints a helpful version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.5")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
