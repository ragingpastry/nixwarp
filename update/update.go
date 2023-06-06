package update

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ragingpastry/nixwarp/logger"
	"github.com/ragingpastry/nixwarp/types"
	"github.com/ragingpastry/nixwarp/utils"
	"golang.org/x/sync/errgroup"
)

var Log *logger.Logger

func UpdateFlake() {
	Log.Info("Updating flake.lock")
	cmd := exec.Command("nix", "flake", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		message := fmt.Sprintf("🚫 Error running `nix flake update`! Error: %s", err.Error())
		Log.Error(message)
	}
	Log.Debug(string(out))

}

func UpdatePackage(pkgName string, pkgDir string) {
	Log.Info(fmt.Sprintf("🚀 Updating package %s", pkgName))
	pkgUpdateCommand := &utils.RunCommand{
		Command:   fmt.Sprintf("nix-update %s", pkgName),
		Directory: pkgDir,
	}
	utils.RunCmd(pkgUpdateCommand)
}

func UpdatePackages(pkgDir string) {
	Log.Debug("Updating local nixPkgs")
	command := &utils.RunCommand{
		Command:   "grep -lr fetchFromGitHub",
		Directory: pkgDir,
	}
	output, err := utils.RunCmd(command)
	if err != nil {
		message := fmt.Sprintf("🚫 Error running `%s`! Error: %s", command.Command, err.Error())
		Log.Error(message)
		os.Exit(1)
	}
	pkgs := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, pkg := range pkgs {
		pkgName := filepath.Dir(pkg)
		UpdatePackage(pkgName, pkgDir)
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
		}
	}
	Log.Debug(string(stdout.Bytes()))
	Log.Debug("Node does not require a reboot")
	return false
}

func UpdateNode(node string, reboot bool) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	defer s.Stop()

	utils.SpinnerMessage(fmt.Sprintf(" Checking if node %s is online...", node), s)
	if !utils.CheckNodeOnline(node) {
		utils.SpinnerMessageWarn(fmt.Sprintf("Node %s is offline. Skipping...", node), s)
		return nil
	}
	utils.SpinnerMessage(fmt.Sprintf(" Node %s is online. Running updates...", node), s)
	command := &utils.RunCommand{
		Command: fmt.Sprintf("nixos-rebuild --flake .#%s switch --target-host %s --use-remote-sudo --use-substitutes", node, node),
	}
	output, err := utils.RunCmd(command)
	Log.Debug(string(output))
	if err != nil {
		utils.SpinnerMessageError(fmt.Sprintf("🚫 Error updating node %s!", node), s)
		nodeErr := types.NodeError{Node: node, Err: err, Command: command.Command}
		return &nodeErr
	}

	utils.SpinnerMessageInfo(fmt.Sprintf("✅ Updates have completed for node %s", node), s)
	if rebootRequired(node) {
		utils.SpinnerMessageWarn(fmt.Sprintf("Reboot is required for node %s", node), s)
		if reboot {
			rebootCmd := &utils.RunCommand{
				Command: "sudo shutdown -r +1 'System will reboot in 1 minute.'",
				Host:    node,
			}
			utils.RunRemoteCmd(rebootCmd, node)
			utils.SpinnerMessageWarn(fmt.Sprintf("❗ Reboot scheduled in 1 minute for node %s", node), s)
		}
	}

	return nil
}

func UpdateNodes(nodes []string, reboot bool) {
	var g errgroup.Group
	errChan := make(chan *types.NodeError, len(nodes))
	for _, node := range nodes {
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
		defer s.Stop()
		Log.Info(fmt.Sprintf("🚀 Running updates on node %s", node))
		g.Go(func(node string) func() error {
			return func() error {
				err := UpdateNode(node, reboot)
				if err != nil {
					nodeErr := &types.NodeError{Node: node, Err: err}
					errChan <- nodeErr
					return nodeErr
				}
				errChan <- nil
				return nil
			}
		}(node))
	}
	var nodeErrors []*types.NodeError
	g.Wait()

	for i := 0; i < len(nodes); i++ {
		nodeErrors = append(nodeErrors, <-errChan)
	}
	close(errChan)
	for _, err := range nodeErrors {
		if err != nil {
			Log.Error(fmt.Sprintf("🚫 Error updating nodes %s", err.Node))
		}
	}
}
