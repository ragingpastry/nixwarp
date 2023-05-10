package cmd


import (
	"fmt"
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
	Use: "update",
	Aliases: []string{"u"},
	Short: "Run updates",
}

var updateNodeCmd = &cobra.Command{
	Use: "node [node1|node1,node2]",
	Aliases: []string{"n"},
	Args: cobra.MaximumNArgs(1),
	Short: "Run updates on NixOS nodes",
	Long: "Run updates on NixOS nodes",
	Run: func(cmd *cobra.Command, args []string) {
		deps := []string{"nix", "ssh", "nixos-rebuild"}
		utils.CheckDependencies(deps)
		var nodes []string
		if updateConfig.UpdateFlake {
			update.UpdateFlake()
		}
		if updateConfig.AllNodes {
		  	log.Info("Running updates on all nodes")
		  	nodes = utils.ParseNodes()
		} else if len(args) > 0 {
			nodes = strings.Split(args[0], ",")
		} else {
			log.Error("🚫 Must specify either `--all-nodes` or a comma-separated list of nodes to update!")
		}

		if len(nodes) > 0 {
			for _, node := range nodes {
				update.UpdateNode(node, updateConfig.Reboot)
			}
		} else {
			log.Error("🚫 Something went wrong! Couldn't find any nodes in our node list. Cowardly aborting...")
		}
	},
}

var updateFlakeCmd = &cobra.Command{
	Use: "flake",
	Aliases: []string{"f"},
	Short: "Run `nix flake update`",
	Long: "Run `nix flake update`",
	Run: func(cmd *cobra.Command, args []string) {
		update.UpdateFlake()
	},

}

var updatePkgCmd = &cobra.Command{
	Use: "package [NAME]",
	Aliases: []string{"p"},
	Short: "Run nix-update on a package",
	Long: "Run nix-update on a package. If no packages are specified try to be smart about it and about all packages.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Updating locally defined packages")
		deps := []string{"nix-update"}
		utils.CheckDependencies(deps)
		if len(args) > 0 {
			nixPackages := strings.Split(args[0], ",")
			for _, nixPackage := range nixPackages {
				update.UpdatePackage(nixPackage, updateConfig.PkgDir)
			}
		} else {
			update.UpdatePackages(updateConfig.PkgDir)
		}
	},
}

var updateAllCmd = &cobra.Command{
	Use: "all",
	Aliases: []string{"a"},
	Short: "Run all updates. Flake, nixpkgs, and nodes.",
	Run: func(cmd *cobra.Command, args []string) {
		deps := []string{"nix-update", "nix", "ssh", "nixos-rebuild"}
		utils.CheckDependencies(deps)
		update.UpdateFlake()
		update.UpdatePackages(updateConfig.PkgDir)
		for _, node := range utils.ParseNodes() {
			message := fmt.Sprintf("🚀 Starting updates for node %s", node)
			log.Info(message)
			update.UpdateNode(node, updateConfig.Reboot)
		}
	},
}

func bindUpdateFlags() {
	updateNodeFlags := updateNodeCmd.Flags()
	updatePkgFlags := updatePkgCmd.Flags()

	updateNodeFlags.BoolVarP(&updateConfig.AllNodes, "all-nodes", "a", false, "Run updates on all nodes")
	updateNodeFlags.BoolVarP(&updateConfig.Reboot, "reboot", "r", false, "Force reboots on nodes")
	updateNodeFlags.BoolVarP(&updateConfig.UpdateFlake, "update-flake", "", false, "Updates the local flake")
	updatePkgFlags.StringVarP(&updateConfig.PkgDir, "pkg-dir", "", "pkgs/", "Directory containing local nixpkg definitions")
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateFlakeCmd)
	updateCmd.AddCommand(updateNodeCmd)
	updateCmd.AddCommand(updatePkgCmd)
	updateCmd.AddCommand(updateAllCmd)

	bindUpdateFlags()
}
