package utils

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
)

// Most of this file was inspired on https://github.com/flaviostutz/promster/blob/master/utils.go

// ExecShellTimeout execute a shell command (like bash -c 'your command') with a timeout. After that time, the process will be cancelled
func ExecShellTimeout(workingDir string, command string, timeout time.Duration) (string, error) {
	// logrus.Debugf("shell command: %s", command)
	// fmt.Printf("shell command: %s\n", command)

	acmd := cmd.NewCmd("sh", "-c", command)

	// detect if a simple process call instead of a shell could be used
	if strings.HasPrefix(command, "/") &&
		!strings.Contains(command, "|") &&
		!strings.Contains(command, ">") &&
		!strings.Contains(command, "&") &&
		!strings.Contains(command, "\"") &&
		!strings.Contains(command, ";") {
		cmdArgs := strings.Split(command, " ")
		acmd = cmd.NewCmd(cmdArgs[0], cmdArgs[1:]...)
	}

	if workingDir != "" {
		acmd.Dir = workingDir
	}
	statusChan := acmd.Start() // non-blocking
	running := true

	//kill if taking too long
	if timeout > 0 {
		logrus.Debugf("Enforcing timeout %s", timeout)
		go func() {
			startTime := time.Now()
			for running {
				if time.Since(startTime) >= timeout {
					logrus.Warnf("Stopping command execution because it is taking too long (%d seconds)", time.Since(startTime))
					acmd.Stop()
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}

	// logrus.Debugf("Waiting for command to finish...")
	<-statusChan
	// logrus.Debugf("Command finished")
	running = false

	out := GetCmdOutput(acmd)
	status := acmd.Status()
	// logrus.Debugf("shell output (%d): %s", status.Exit, out)
	if status.Exit != 0 {
		return out, fmt.Errorf("Failed to run command: '%s'; exit=%d; out=%s", command, status.Exit, out)
	} else {
		return out, nil
	}
}

// ExecShell execute a shell command (like bash -c 'your command')
func ExecShell(workingDir string, command string) (string, error) {
	return ExecShellTimeout(workingDir, command, 0)
}

// ExecShellf execute a shell command (like bash -c 'your command') but with format replacements
func ExecShellf(workingDir string, command string, args ...interface{}) (string, error) {
	cmd := fmt.Sprintf(command, args...)
	return ExecShellTimeout(workingDir, cmd, 0)
}

// GetCmdOutput join stdout and stderr in a single string from Cmd
func GetCmdOutput(cmd *cmd.Cmd) string {
	status := cmd.Status()
	out := strings.Join(status.Stdout, "\n")
	out = out + "\n" + strings.Join(status.Stderr, "\n")
	return out
}

func linesToArray(lines string) ([]string, error) {
	var result = []string{}
	scanner := bufio.NewScanner(strings.NewReader(lines))
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if scanner.Err() != nil {
		return []string{}, scanner.Err()
	}
	return result, nil
}
