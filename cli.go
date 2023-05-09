package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ragingpastry/nixos-configs/nixwarp/logger"
	"github.com/ragingpastry/nixos-configs/nixwarp/utils"
	"github.com/urfave/cli/v2"
)

var log = logger.NewLogger(false)

func updateFlake() {
	cmd := exec.Command("nix", "flake", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := fmt.Sprintf("🚫 Error running `nix flake update`! Error: %s", err.Error())
		log.Error(message)
	}
	log.Debug("Updating flake.lock")
	log.Debug(string(out))

}

func updatePackages() {
	log.Debug("Updating local nixPkgs")
	command := &utils.RunCommand{
		Command:   "grep -lr fetchFromGitHub",
		Directory: "pkgs/",
	}
	output := utils.RunCmd(command)
	pkgs := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, pkg := range pkgs {
		pkgName := filepath.Dir(pkg)
		log.Info(fmt.Sprintf("🚀 Updating package %s", pkgName))
		pkgUpdateCommand := &utils.RunCommand{
			Command:   fmt.Sprintf("nix-update %s", pkgName),
			Directory: "pkgs/",
		}
		utils.RunCmd(pkgUpdateCommand)
	}

}

func rebootRequired(node string) bool {
	log.Debug("Checking if node needs to be rebooted")
	cmd := exec.Command("ssh", node, "bash -c ':; diff <(readlink /run/booted-system/{initrd,kernel,kernel-modules,systemd}) <(readlink /nix/var/nix/profiles/system/{initrd,kernel,kernel-modules,systemd})'")
	log.Debug(fmt.Sprintf("Running command :%s", cmd.Args))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 1 {
			log.Debug("Node requires a reboot to complete updates")
			return true
		} else {
			message := fmt.Sprintf("🚫 Error running `%s`! Error: %s", cmd.Args, string(stderr.Bytes()))
			log.Error(message)
		};
	};
    log.Debug(string(stdout.Bytes()))
	log.Debug("Node does not require a reboot")
	return false
}

func updateNode(node string, reboot bool) {
	log.Debug(fmt.Sprintf("Running nixos-rebuild on node: %s", node))
	command := &utils.RunCommand{
		Command: fmt.Sprintf("nixos-rebuild --flake .#%s switch --target-host %s --use-remote-sudo --use-substitutes", node, node),
	}
	utils.RunCmd(command)
	log.Info(fmt.Sprintf("✅ Updates have completed for node %s", node))
	if rebootRequired(node) {
		log.Warn(fmt.Sprintf("❗ Reboot is required for node %s", node))
		if reboot {
			rebootCmd := &utils.RunCommand{
				Command:   "sudo shutdown -r +1 'System will reboot in 1 minute.'",
				Host: node,
			}
			utils.RunRemoteCmd(rebootCmd, node)
			log.Warn(fmt.Sprintf("❗ Reboot scheduled in 1 minute for node %s", node))
		}
	}
}

func runUpdates(nodes []string, reboot bool) {
	updateFlake()
	updatePackages()
	for _, node := range nodes {
		message := fmt.Sprintf("🚀 Starting updates for node %s", node)
		log.Info(message)
		updateNode(node, reboot)
	}
}

type Hosts struct {
	HostType string `json:"type"`
}

type FlakeOutput struct {
	NixosConfigurations map[string]Hosts `json:"nixosConfigurations"`
}

func parseNodes() []string {
	var nodes []string
	command := &utils.RunCommand{
		Command: "nix flake show --json",
	}
	flakeCmdOutput := utils.RunCmd(command)
	var flakeOutput FlakeOutput
	json.Unmarshal([]byte(flakeCmdOutput), &flakeOutput)
	for hostName, _ := range flakeOutput.NixosConfigurations {
		log.Debug(fmt.Sprintf("Found node %s", hostName))
		nodes = append(nodes, hostName)
	}

	return nodes
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Display debug information",
				Action: func(ctx *cli.Context, debug bool) error {
					if debug {
						log.EnableDebug()
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:     "update",
				Aliases:  []string{"u"},
				Usage:    "Runs updates on NixOS nodes",
				Category: "update",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("all-nodes") {
						log.Info("Running updates on all nodes")
						nodes := parseNodes()
						runUpdates(nodes, cCtx.Bool("reboot"))
					} else if cCtx.NArg() > 0 {
						nodes := strings.Split(cCtx.Args().First(), ",")
						runUpdates(nodes, cCtx.Bool("reboot"))
					} else {
						log.Error("🚫 Must specify either `--all-nodes` or a comma-separated list of nodes to update!")
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "all-nodes",
						Usage: "Run updates on all nodes",
					},
					&cli.BoolFlag{
						Name: "reboot",
						Usage: "Reboot the node (if required) after updates are finished.",
					},
				},
			},
		},
	}
	utils.Log = log
	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
	}
}
