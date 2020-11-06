package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	version      = "1.0.0"
	command      = "plink.exe"
	sshSignature = "OpenSSH_8.3p1, OpenSSL 1.1.1g  21 Apr 2020" // "OpenSSH_for_Windows_7.7p1, LibreSSL 2.6.5"
)

func trimArgs(argToTrim string, args []string) []string {
	var result []string
	for _, arg := range args {
		if arg != argToTrim {
			result = append(result, arg)
		}
	}
	return result
}

func main() {

	args := os.Args

	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "ssh2plink %v\n", version)
	}

	fullCmd, err := resolveCmd(command)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Please make sure %v is in your PATH (%v)\n", command, os.Getenv("PATH"))
		return
	}

	if len(args) == 1 {
		cmd := exec.Command(fullCmd, "-V")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	if args[1] == "-V" {
		fmt.Fprint(os.Stderr, sshSignature)
		return
	}

	args = trimArgs("-T", args[1:])

	sshCmd := exec.Command(fullCmd, args...)
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	sshCmd.Stdin = os.Stdin

	err = sshCmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start sshCmd: %v %v", fullCmd, args)
		fmt.Fprintf(os.Stderr, "sshCmd.Start failed: %s\n", err)
		return
	}

	err = sshCmd.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to wait sshCmd: %v %v", fullCmd, args)
		fmt.Fprintf(os.Stderr, "sshCmd.Wait failed: %s\n", err)
		return
	}
}

func resolveCmd(cmd string) (string, error) {
	osPath := os.Getenv("PATH")

	for _, p := range strings.Split(osPath, ";") {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			continue
		}

		cmdPath := filepath.Join(p, cmd)
		info, err := os.Stat(cmdPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			continue
		}

		return cmdPath, nil
	}

	return "", fmt.Errorf("Command not found in PATH: %v", cmd)
}
