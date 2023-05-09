package utils

import (
	"os/exec"
	"bytes"
	"fmt"
	"strings"
	"os"

	"github.com/ragingpastry/nixos-configs/nixwarp/logger"
)

var Log *logger.Logger

type RunCommand struct {
	Command string
	Environment string
	Directory string
	Host string
}

func RunCmd(command *RunCommand) []byte  {
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
    	message := fmt.Sprintf("🚫 Error running `%s`! Error: %s", cmd.Args, string(stderr.Bytes()))
    	Log.Error(message)
    }
    Log.Debug(string(stdout.Bytes()))
	return stdout.Bytes()
}

func RunRemoteCmd(command *RunCommand, node string) []byte {
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
    }
    Log.Debug(string(stdout.Bytes()))
	return stdout.Bytes()

}