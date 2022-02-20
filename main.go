package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mikkeloscar/sshconfig"
)

const (
	version = "1.1.0"
	command = "plink.exe"
)

// boolean options from ssh that are not supported by plink
var dropBool = [...]string{"-f", "-g", "-G", "-k", "-K", "-q", "-y"}

// string options from ssh that are not supported by plink
var dropStr = [...]string{"-b", "-B", "-c", "-e", "-E", "-F", "-I", "-J", "-m", "-o", "-O", "-Q", "-S", "-w", "-W"}

func replaceOrSetArgValue(argToReplace string, newValue string, args []string) []string {
	isSet := false
	for i := 0; i < len(args); i = i + 1 {
		if (args[i] == argToReplace) && ((i+1) < len(args) && !strings.HasPrefix(args[i+1], "-")) {
			args[i+1] = newValue
			isSet = true
			break
		}
	}
	if !isSet {
		args = append([]string{argToReplace, newValue}, args...)
	}
	return args
}

func replaceArgs(argToReplace string, newArg string, args []string) []string {
	for i := 0; i < len(args); i = i + 1 {
		if args[i] == argToReplace {
			args[i] = newArg
			break
		}
	}
	return args
}

func removeIndex(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func hasEntry(arr []string, value string) bool {
	res := false
	for _, entry := range arr {
		if entry == value {
			res = true
			break
		}
	}
	return res
}

func removeSshOptionsUnsupportedByPlink(args []string) []string {
	idxToRemove := []int{}
	for i := 0; i < len(args); i = i + 1 {
		if hasEntry(dropStr[:], args[i]) {
			idxToRemove = append(idxToRemove, i, i+1)
		} else if hasEntry(dropBool[:], args[i]) {
			idxToRemove = append(idxToRemove, i)
		}
	}

	if len(idxToRemove) > 0 {
		for i := len(idxToRemove) - 1; i >= 0; i = i - 1 {
			args = removeIndex(args, idxToRemove[i])
		}
	}
	return args
}

// if existing the ssh config file is taken into account
// from the real progam arguments the first one is the host and it is checked if this Host could be found in ssh config
// if found and setting for HostName, Port, IdentityFile are found the correponding options for plink will be added or overwritten as options
func handleSshConfig(realArgs []string, args []string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}

	if fileExists(filepath.Join(home, ".ssh", "config")) {
		sshconfigs, err := sshconfig.Parse(filepath.Join(home, ".ssh", "config"))
		if err != nil {
			fmt.Println(err)
		}

		for _, sshCfg := range sshconfigs {
			for _, host := range sshCfg.Host {
				// first argument is the host
				if host == realArgs[0] {
					args = replaceArgs(realArgs[0], sshCfg.HostName, args)
					if sshCfg.User != "" {
						args = replaceOrSetArgValue("-l", sshCfg.User, args)
					}
					if sshCfg.IdentityFile != "" {
						args = replaceOrSetArgValue("-i", sshCfg.IdentityFile, args)
					}
					if sshCfg.Port > 0 {
						args = replaceOrSetArgValue("-p", strconv.Itoa(sshCfg.Port), args)
					}
					break
				}
			}
		}
	}
	return args
}

// if program is started with -V VSCode is checking for the version of the ssh client
// so we are calling the real ssh client and returning the version as wanted
func handleSshVersion() {
	sshPath, err := resolveCmd("ssh.exe")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Please make sure %v is in your PATH (%v)\n", command, os.Getenv("PATH"))
		return
	}
	sshCmdVersion := exec.Command(sshPath, "-V")
	out, err := sshCmdVersion.CombinedOutput()

	if err != nil {
		fmt.Fprintf(os.Stderr, "ssh not could not started (wait): %s\n", err)
	}
	filterRegExp := regexp.MustCompile(`[\n]+`)
	version := filterRegExp.ReplaceAllString(string(out), "")
	fmt.Fprintf(os.Stdout, "%s", version)
}

// register all possible ssh options to be able to get the real arguments like host & shell
func registerSshOptions() {
	flag.Bool("4", false, "Forces ssh to use IPv4 addresses only")
	flag.Bool("6", false, "Forces ssh to use IPv6 addresses only")
	flag.Bool("A", false, "Enables forwarding of connections from an authentication agent such as ssh-agent(1).")
	flag.Bool("a", false, "Disables forwarding of the authentication agent connection.")
	flag.Bool("C", false, "Requests compression of all data (including stdin, stdout, stderr, and data for forwarded X11, TCP and UNIX-domain connections).")
	flag.Bool("f", false, "Requests ssh to go to background just before command execution.")
	flag.Bool("G", false, "Causes ssh to print its configuration after evaluating Host and Match blocks and exit. ")
	flag.Bool("g", false, "Allows remote hosts to connect to local forwarded ports. If used on a multiplexed connection, then this option must be specified on the master process.")
	flag.Bool("K", false, "Enables GSSAPI-based authentication and forwarding (delegation) of GSSAPI credentials to the server. ")
	flag.Bool("k", false, "Disables forwarding (delegation) of GSSAPI credentials to the server.")
	flag.Bool("M", false, "Places the ssh client into “master” mode for connection sharing.")
	flag.Bool("N", false, "Do not execute a remote command. This is useful for just forwarding ports. Refer to the description of SessionType in ssh_config(5) for details.")
	flag.Bool("n", false, "Redirects stdin from /dev/null (actually, prevents reading from stdin). This must be used when ssh is run in the background.")
	flag.Bool("q", false, "Quiet mode. Causes most warning and diagnostic messages to be suppressed.")
	flag.Bool("s", false, "May be used to request invocation of a subsystem on the remote system. ")
	flag.Bool("T", false, "Disable pseudo-terminal allocation.")
	flag.Bool("t", false, "Force pseudo-terminal allocation.")
	flag.Bool("V", false, "Display the version number and exit. ")
	flag.Bool("v", false, "Verbose mode.")
	flag.Bool("X", false, "Enables X11 forwarding. This can also be specified on a per-host basis in a configuration file. ")
	flag.Bool("x", false, "Disables X11 forwarding. ")
	flag.Bool("Y", false, "Enables trusted X11 forwarding. Trusted X11 forwardings are not subjected to the X11 SECURITY extension controls. ")
	flag.Bool("y", false, "Send log information using the syslog(3) system module. By default this information is sent to stderr.")

	flag.String("B", "", "-B <bind_interface> - Bind to the address of bind_interface before attempting to connect to the destination host.")
	flag.String("b", "", "-b <bind_address> - Use bind_address on the local machine as the source address of the connection.")
	flag.String("c", "", "-c <cipher_spec> - Selects the cipher specification for encrypting the session. cipher_spec is a comma-separated list of ciphers listed in order of preference.")
	flag.String("D", "", "-D <[bind_address:]port> - Specifies a local “dynamic” application-level port forwarding.")
	flag.String("E", "", "-E <log_file> - Append debug logs to log_file instead of standard error.")
	flag.String("e", "", "-e <escape_char> - Sets the escape character for sessions with a pty (default: ~).")
	flag.String("F", "", "-F <config_file> - Specifies an alternative per-user configuration file.")
	flag.String("I", "", "-I <pkcs11> - Specify the PKCS#11 shared library ssh should use to communicate with a PKCS#11 token providing keys for user authentication.")
	flag.String("i", "", "-i <identity_file> - Selects a file from which the identity (private key) for public key authentication is read.")
	flag.String("J", "", "-J <destination> - Connect to the target host by first making a ssh connection to the jump host described by destination and then establishing a TCP forwarding to the ultimate destination from there.")
	flag.String("L", "", "-L <[bind_address:]port:host:hostport>\n-L <[bind_address:]port:remote_socket>\n-L <local_socket:host:hostport>\n-L <local_socket:remote_socket> - Specifies that connections to the given TCP port or Unix socket on the local (client) host are to be forwarded to the given host and port, or Unix socket, on the remote side.")
	flag.String("l", "", "-l <login_name> - Specifies the user to log in as on the remote machine.")
	flag.String("m", "", "-m <mac_spec> - A comma-separated list of MAC (message authentication code) algorithms, specified in order of preference.")
	flag.String("O", "", "-O <ctl_cmd> - Control an active connection multiplexing master process.")
	flag.String("o", "", "-o <option> - Can be used to give options in the format used in the configuration file.")
	flag.String("p", "", "-p <port> - Port to connect to on the remote host. ")
	flag.String("Q", "", "-Q <query_option> - Queries for the algorithms supported.")
	flag.String("R", "", "-R <address> - Specifies that connections to the given TCP port or Unix socket on the remote (server) host are to be forwarded to the local side. ")
	flag.String("S", "", "-S <ctl_path> - Specifies the location of a control socket for connection sharing, or the string “none” to disable connection sharing. ")
	flag.String("W", "", "-W <host:port> - Requests that standard input and output on the client be forwarded to host on port over the secure channel.")
	flag.String("w", "", "-w <local_tun[:remote_tun]> - Requests tunnel device forwarding with the specified tun(4) devices between the client (local_tun) and the server (remote_tun). ")
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// tries first to find the cmd file in the current working directory
// if not there loop through paths given in PATH environment variable and return filepath path if found
func resolveCmd(cmd string) (string, error) {
	currProcess := os.Args[0]
	cwd := filepath.Dir(currProcess)

	osPathArr := strings.Split(os.Getenv("PATH"), ";")
	pathArr := append([]string{cwd}, osPathArr...)

	for _, path := range pathArr {
		path = strings.TrimSpace(path)
		if len(path) == 0 {
			continue
		}
		cmdPath := filepath.Join(path, cmd)
		if fileExists(cmdPath) {
			return cmdPath, nil
		}
	}

	return "", fmt.Errorf("Command not found in PATH: %v\n", cmd)
}

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
	registerSshOptions()
	argsWithCmd := os.Args
	var args = make([]string, len(argsWithCmd[1:]))
	copy(args, argsWithCmd[1:])
	flag.Parse()
	realArgs := flag.Args()

	fullPlinkCmd, err := resolveCmd(command)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Please make sure %v is in your current working directory or in PATH (%v)\n", command, os.Getenv("PATH"))
		return
	}

	if len(argsWithCmd) == 1 {
		fmt.Fprintf(os.Stderr, "ssh2plink %v\n", version)
		cmd := exec.Command(fullPlinkCmd, "-V")
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	if len(argsWithCmd) > 1 {

		if argsWithCmd[1] == "-V" {
			handleSshVersion()
		} else {
			// remove unsupported bool & string args
			args = removeSshOptionsUnsupportedByPlink(args)
			args = handleSshConfig(realArgs, args)

			args = replaceArgs("-p", "-P", args)
			// args = trimArgs("-T", args)

			sshCmd := exec.Command(fullPlinkCmd, args...)
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			sshCmd.Stdin = os.Stdin

			err := sshCmd.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to start sshCmd: %s %s\n", fullPlinkCmd, strings.Join(args, " "))
				fmt.Fprintf(os.Stderr, "sshCmd.Start faild with: %s\n", err)
			}

			err = sshCmd.Wait()
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to wait sshCmd: %s %s\n", fullPlinkCmd, strings.Join(args, " "))
				fmt.Fprintf(os.Stderr, "sshCmd.Wait failed with: %s\n", err)
			}
		}
	}
}
