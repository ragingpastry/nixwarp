package cmd


import (
  "github.com/spf13/cobra"
  "github.com/ragingpastry/nixwarp/utils"
  "github.com/ragingpastry/nixwarp/update"
  "github.com/ragingpastry/nixwarp/logger"
)

var (
	debug bool
)


var rootCmd = &cobra.Command{
  Use:   "nixwarp [COMMAND]",
  Short: "Run updates on NixOS nodes",
  Long: `Runs updates on my NixOS flake. Built for my convienence
                and to learn more Golang.
                There is no documentation available.`,
  PersistentPreRun: func(cmd *cobra.Command, args []string) {
	var log = logger.NewLogger(debug)
	utils.Log = log
	update.Log = log
  },
  Run: func(cmd *cobra.Command, args []string) {
	cmd.Help()
  },
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug output")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}