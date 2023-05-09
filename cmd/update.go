package cmd


import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ragingpastry/nixwarp/logger"
	"github.com/ragingpastry/nixwarp/utils"
	"github.com/ragingpastry/nixwarp/update"
	"github.com/ragingpastry/nixwarp/types"
)

var (
	updateConfig = types.UpdateConfig{}
	log = logger.NewLogger(false)
)

var updateCmd = &cobra.Command{
	Use: "update [node] or update [node1,node2]",
	Aliases: []string{"u"},
	Args: cobra.MaximumNArgs(1),
	Short: "Run updates on NixOS nodes",
	Long: "Run updates on NixOS nodes",
	Run: func(cmd *cobra.Command, args []string) {
		if updateConfig.AllNodes {
			log.Info("Running updates on all nodes")
			nodes := utils.ParseNodes()
			update.RunUpdates(nodes, updateConfig.Reboot)
		} else if len(args) > 0 {
			nodes := strings.Split(args[0], ",")
			update.RunUpdates(nodes, updateConfig.Reboot)
		} else {
			log.Error("🚫 Must specify either `--all-nodes` or a comma-separated list of nodes to update!")
		}
	},
}

func bindUpdateFlags() {
	updateFlags := updateCmd.Flags()

	updateFlags.BoolVarP(&updateConfig.AllNodes, "all-nodes", "a", false, "Run updates on all nodes")
	updateFlags.BoolVarP(&updateConfig.Reboot, "reboot", "r", false, "Force reboots on nodes")
}

func init() {
	rootCmd.AddCommand(updateCmd)

	bindUpdateFlags()
}
