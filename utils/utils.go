package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ragingpastry/nixwarp/logger"
)

var Log *logger.Logger

type RunCommand struct {
	Command     string
	Environment string
	Directory   string
	Host        string
}

type Hosts struct {
	HostType string `json:"type"`
}

type FlakeOutput struct {
	NixosConfigurations map[string]Hosts `json:"nixosConfigurations"`
}

func RunCmd(command *RunCommand) ([]byte, error) {
	cmd := exec.Command(strings.Split(command.Command, " ")[0], strings.Split(command.Command, " ")[1:]...)
	Log.Debug(fmt.Sprintf("Running command :%s", cmd.Args))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if command.Environment != "" {
		newEnv := append(os.Environ(), command.Environment)
		cmd.Env = newEnv
	}
	if command.Directory != "" {
		cmd.Dir = command.Directory
	}
	err := cmd.Run()
	if err != nil {
		message_stdout := fmt.Sprintf("🚫 Error running `%s`! Error: %s\n", cmd.Args, string(stdout.Bytes()))
		message_stderr := fmt.Sprintf("🚫 Error running `%s`! Error: %s\n", cmd.Args, string(stderr.Bytes()))
		Log.Debug(message_stdout + message_stderr)
	}
	Log.Debug(string(stdout.Bytes()))
	return stdout.Bytes(), err
}

func RunRemoteCmd(command *RunCommand, node string) ([]byte, error) {
	cmd := exec.Command("ssh", node, command.Command)
	Log.Debug(fmt.Sprintf("Running command :%s", cmd.Args))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if command.Environment != "" {
		newEnv := append(os.Environ(), command.Environment)
		cmd.Env = newEnv
	}
	if command.Directory != "" {
		cmd.Dir = command.Directory
	}
	err := cmd.Run()
	if err != nil {
		message := fmt.Sprintf("🚫 Error running `%s`! Error: %s", cmd.Args, string(stderr.Bytes()))
		Log.Error(message)
		os.Exit(1)
	}
	Log.Debug(string(stdout.Bytes()))
	return stdout.Bytes(), err

}

func ParseNodes() []string {
	var nodes []string
	command := &RunCommand{
		Command: "nix flake show --json",
	}
	flakeCmdOutput, err := RunCmd(command)
	if err != nil {
		Log.Error(fmt.Sprintf("🚫 Error running `nix flake show --json`! Error: %s", err.Error()))
		os.Exit(1)
	}
	var flakeOutput FlakeOutput
	json.Unmarshal([]byte(flakeCmdOutput), &flakeOutput)
	for hostName, _ := range flakeOutput.NixosConfigurations {
		Log.Debug(fmt.Sprintf("Found node %s", hostName))
		nodes = append(nodes, hostName)
	}

	return nodes
}

func CheckDependencies(dependencies []string) {
	var depsMissing []string
	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			depsMissing = append(depsMissing)
		}
	}
	if len(depsMissing) > 0 {
		Log.Error(fmt.Sprintf("🚫 Error missing dependencies: %s", depsMissing))
		os.Exit(1)
	}
}

func CheckNodeOnline(node string) bool {
	command := &RunCommand{
		Command: fmt.Sprintf("ssh -o ConnectTimeout=10 %s echo ping", node),
	}
	output, err := RunCmd(command)
	if err != nil {
		Log.Debug(fmt.Sprintf("🚫 Error connecting to node %s! Error: %s", node, err.Error()))
		return false
	}
	if strings.Contains(string(output), "ping") {
		return true
	} else {
		return false
	}
}
