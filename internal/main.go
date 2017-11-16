// Package internal is the ironic junk drawer of random shared functions™
// This is where dreams go to dream gooder and do other stuff good too
// You should not do this because it's bad programming, but I'm a human
// with flaws and at least it is not named "util" ¯\\_(ツ)_/¯
package internal

import (
	"fmt"
	"github.com/pierrre/archivefile/zip"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// Internal describes fields which support the internal commands
// that are required of other commands
type Internal struct {
	ConsulVersion string
	HostName      string
	NomadVersion  string
	OS            string
	LogFile       string
	TargetFile    string
	VaultVersion  string
}

// CheckProc checks for a running process by name with pgrep or ps
func CheckProc(name string) (bool, string) {
	// If `pgrep` is around, use that...
	path, err := exec.LookPath("pgrep")
	if err != nil {
		log.Printf("[i] pgrep not found in system PATH: %s", path)
		// If no `pgrep`, check for running process with a POSIX-y `ps`
		out, err := exec.Command("sh", "-c", fmt.Sprintf("ps -A | grep -i %s | head -1 | awk '{print $1}'", name)).Output()
		pid := strings.TrimSpace(string(out))
		if err != nil {
			log.Printf("[e] Error determining PID of %s", name)
			pid = "ENOIDEA"
			return false, pid
		}
		if len(out) > 0 {
			log.Printf("[i] Process detected! PID of %s is: %s", name, pid)
		}
		return true, pid
	}

	log.Println("[i] pgrep found in system PATH.")
	// Check for running process with `pgrep`
	out, err := exec.Command("pgrep", name).Output()
	pid := strings.TrimSpace(string(out))
	if err != nil {
		log.Printf("[e] Error determining PID of %s", name)
		pid = "ENOIDEA"
		return false, pid
	}
	if len(out) > 0 {
		log.Printf("[i] Process detected! PID of %s is: %s", name, pid)
	}
	return true, pid
}

// CheckHashiVersion attempts to locate HashiCorp runtime tools and get
// their versions - Consul has slightly different version output style so
// it must be handled differently
func CheckHashiVersion(name string) string {
	active, pid := CheckProc(name)
	if active {
		log.Printf("[i] %s process identified as %s", name, pid)
		path, err := exec.LookPath(name)
		if err != nil {
			log.Printf("[i] Did not find a %s binary in PATH", name)
		}
		if name == "consul" {
			version, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | head -n 1 | awk '{print $2}'", path)).Output()
			if err != nil {
				log.Printf("[e] Error executing %s binary! Error: %v", name, err)
			}
			return string(version)
		} else if name == "nomad" || name == "vault" {
			version, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | awk '{print $2}'", path)).Output()
			if err != nil {
				log.Printf("[e] Error executing %sl binary! Error: %v", name, err)
			}
			return string(version)
		}
	}
	return "ENOVERSION"
}

// Dump takes a type, output filename and command, which it then executes
// while also writing stdout + stderr to a file named for the command
// Inspired by debug-ninja! (https://github.com/fprimex/debug-ninja)
func Dump(dumpType string, outfile string, cmdName string, args ...string) int {

	s := Internal{HostName: GetHostName(),
		TargetFile: "%s/%s/%s.txt"}

	path, err := exec.LookPath(cmdName)
	if err != nil {
		log.Printf("[w] Command %s was not found in this system's $PATH", cmdName)
	} else {
		log.Printf("[i] Command %s is available at %s", cmdName, path)

		// We audit all command parameters by specifying them explicitly as
		// Dump() invocations and not allowing any form of user input to
		// be passed in for most command parameter values, so subprocess
		// launching with strict parameters maybe not too terribly sad here
		cmd := exec.Command(cmdName, args...)

		f, err := os.Create(fmt.Sprintf(filepath.Join(".", s.TargetFile),
			s.HostName,
			dumpType,
			outfile))
		if err != nil {
			panic(err)
		}

		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()

		cmd.Stdout = f
		cmd.Stderr = f
		err = cmd.Start()

		if err != nil {
			panic(err)
		}

		// Not as cool as the dots and the Es, but it lets us know something
		if err := cmd.Wait(); err != nil {

			if exiterr, ok := err.(*exec.ExitError); ok {

				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					log.Printf("[e] Command %s %s exited with non-zero status: %d",
						cmdName,
						strings.Join(args[:], " "), status.ExitStatus())

				}

			} else {

				log.Fatalf("cmd.Wait: %v", err)
			}

		}
	}
	return 0
}

// FileExist checks for a file's existence
func FileExist(fileName string) bool {
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		log.Printf("[i] File %s exists", fileName)
		return true
	}
	return false
}

// GetHostName gets the current system's network hostname
func GetHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		log.Printf("[e] Cannot determine hostname!")
		hostName = "unknown"
	}
	return hostName
}

// HTTPCmdCheck checks for curl or wget and use what is present
func HTTPCmdCheck() string {
	path, err := exec.LookPath("curl")
	if err != nil {
		log.Printf("[i] Command curl was not found in this system's $PATH")
		path, err = exec.LookPath("wget")
		if err != nil {
			log.Printf("[i] Command wget was not found in this system's $PATH")
		} else {
			log.Printf("[i] Command wget is available at %s", path)
			httpCommand := "wget"
			return httpCommand
		}
	}
	log.Printf("[i] Command curl is available at %s", path)
	httpCommand := "curl"
	return httpCommand
}

// LogSetup ensures presence of the log path and initializes
// some rudimentary rover self-logging ye olde schoole action
func LogSetup() {

	s := Internal{HostName: GetHostName(), LogFile: "rover.log"}
	logPath := filepath.Join(fmt.Sprintf("%s", s.HostName), "log")

	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		log.Fatalf("[e] Cannot create log directory %s.", logPath)
	}

	f, err := os.OpenFile(filepath.Join(logPath, s.LogFile),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)

	if err != nil {
		log.Fatalf("[e] Error opening log file: %v", err)
	}

	log.SetOutput(f)
}

// ZipIt archives rover results into a zip file suitable for tubing
func ZipIt(target string) {

	hostName := GetHostName()
	outpath := filepath.Join(".", hostName)
	err := zip.ArchiveFile("output", outpath, nil)
	if err != nil {
		panic(err)
	}
	// Remove the source directory after zip successfully created
	err = os.RemoveAll(hostName)
	if err != nil {
		panic(err)
	}
}
