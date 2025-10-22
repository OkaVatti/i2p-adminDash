package i2psnark

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

var defaultSocket = "/var/run/i2psnark.sock" // example; adjust to your system

// RunCommand tries socket control (if present) then falls back to CLI
func RunCommand(cmd string, arg string) (string, error) {
	// Try socket
	out, err := runViaSocket(defaultSocket, cmd, arg)
	if err == nil {
		return out, nil
	}
	// Fallback to CLI
	return runViaCLI(cmd, arg)
}

func runViaSocket(sockPath string, cmd, arg string) (string, error) {
	// Connect to unix socket (if not present, return error)
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	// Write simple command protocol: "CMD ARG\n"
	_, err = fmt.Fprintf(conn, "%s %s\n", cmd, arg)
	if err != nil {
		return "", err
	}
	// read response
	sc := bufio.NewScanner(conn)
	var sb strings.Builder
	for sc.Scan() {
		sb.WriteString(sc.Text())
		sb.WriteString("\n")
	}
	if sc.Err() != nil {
		return "", sc.Err()
	}
	return sb.String(), nil
}

func runViaCLI(cmd, arg string) (string, error) {
	switch cmd {
	case "list":
		out, err := exec.Command("i2psnark", "--list").CombinedOutput()
		return string(out), err
	case "start":
		return runWithArgs("i2psnark", arg)
	case "stop":
		return runWithArgs("i2psnark", "--stop", arg)
	default:
		return "", errors.New("unknown command")
	}
}

func runWithArgs(base string, args ...string) (string, error) {
	cmd := exec.Command(base, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
