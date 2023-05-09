package update

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ragingpastry/nixwarp/logger"
	"github.com/ragingpastry/nixwarp/utils"
)

var Log *logger.Logger


func updateFlake() {
	cmd := exec.Command("nix", "flake", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := fmt.Sprintf("🚫 Error running `nix flake update`! Error: %s", err.Error())
		Log.Error(message)
	}
	Log.Debug("Updating flake.lock")
	Log.Debug(string(out))

}

func updatePackages() {
	Log.Debug("Updating local nixPkgs")
	command := &utils.RunCommand{
		Command:   "grep -lr fetchFromGitHub",
		Directory: "pkgs/",
	}
	output := utils.RunCmd(command)
	pkgs := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, pkg := range pkgs {
		pkgName := filepath.Dir(pkg)
		Log.Info(fmt.Sprintf("🚀 Updating package %s", pkgName))
		pkgUpdateCommand := &utils.RunCommand{
			Command:   fmt.Sprintf("nix-update %s", pkgName),
			Directory: "pkgs/",
		}
		utils.RunCmd(pkgUpdateCommand)
	}

}

func rebootRequired(node string) bool {
	Log.Debug("Checking if node needs to be rebooted")
	cmd := exec.Command("ssh", node, "bash -c ':; diff <(readlink /run/booted-system/{initrd,kernel,kernel-modules,systemd}) <(readlink /nix/var/nix/profiles/system/{initrd,kernel,kernel-modules,systemd})'")
	Log.Debug(fmt.Sprintf("Running command :%s", cmd.Args))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if exitError.ExitCode() == 1 {
			Log.Debug("Node requires a reboot to complete updates")
			return true
		} else {
			message := fmt.Sprintf("🚫 Error running `%s`! Error: %s", cmd.Args, string(stderr.Bytes()))
			Log.Error(message)
		};
	};
    Log.Debug(string(stdout.Bytes()))
	Log.Debug("Node does not require a reboot")
	return false
}

func updateNode(node string, reboot bool) {
	Log.Debug(fmt.Sprintf("Running nixos-rebuild on node: %s", node))
	command := &utils.RunCommand{
		Command: fmt.Sprintf("nixos-rebuild --flake .#%s switch --target-host %s --use-remote-sudo --use-substitutes", node, node),
	}
	utils.RunCmd(command)
	Log.Info(fmt.Sprintf("✅ Updates have completed for node %s", node))
	if rebootRequired(node) {
		Log.Warn(fmt.Sprintf("❗ Reboot is required for node %s", node))
		if reboot {
			rebootCmd := &utils.RunCommand{
				Command:   "sudo shutdown -r +1 'System will reboot in 1 minute.'",
				Host: node,
			}
			utils.RunRemoteCmd(rebootCmd, node)
			Log.Warn(fmt.Sprintf("❗ Reboot scheduled in 1 minute for node %s", node))
		}
	}
}

func RunUpdates(nodes []string, reboot bool) {
	updateFlake()
	updatePackages()
	for _, node := range nodes {
		message := fmt.Sprintf("🚀 Starting updates for node %s", node)
		Log.Info(message)
		updateNode(node, reboot)
	}
}