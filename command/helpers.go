// helpers is the ironic junk drawer of random shared functions™
// that is just chilling all up in your business trying to be helpful

package command

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/pierrre/archivefile/zip"
)

// Internal describes fields which support the internal commands
type Internal struct {
	ConsulVersion string
	HostName      string
	NomadVersion  string
	OS            string
	LogFile       string
	TargetFile    string
	VaultVersion  string
}

// CheckProc checks for a running process by name with pgrep or ps and returns its PID
func CheckProc(name string) (string, error) {
	i := Internal{}
	// Internal logging
	l := "rover.log"
	h, err := GetHostName()
	if err != nil {
		return "", fmt.Errorf("Cannot get system hostname with error %v", err)
	}
	i.HostName = h
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
	pid := ""
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		return pid, err
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		return pid, err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})

	// If pgrep is around, use that...
	path, err := exec.LookPath("pgrep")
	if err != nil {
		logger.Info("check-proc", "pgrep not found in system PATH", path)
		// If no `pgrep`, check for running process with a POSIX-y `ps`
		out, err := exec.Command("sh", "-c", fmt.Sprintf("ps -A | grep -i %s | head -1 | awk '{print $1}'", name)).Output()
		if err != nil {
			logger.Error("check-proc", "cannot determine PID", name)
			return pid, err
		}
		if len(out) > 0 {
			logger.Info("check-proc", "process detected", name, "pid", pid)
		}
		pid := strings.TrimSpace(string(out))
		return pid, nil
	}
	logger.Debug("check-proc", "pgrep found in system PATH")
	// Check for running process with pgrep
	out, err := exec.Command("pgrep", name).Output()
	if err != nil {
		logger.Error("check-proc", "cannot determine PID", name)
		return pid, err
	}
	if len(out) > 0 {
		logger.Info("check-proc", "process detected", name, "pid", pid)
	}
	pid = strings.TrimSpace(string(out))
	return pid, nil
}

// CheckHashiVersion attempts to locate HashiCorp runtime tools and get
// their versions - Consul has slightly different version output style so
// it must be handled differently
func CheckHashiVersion(name string) string {
	i := Internal{}
	// Internal logging
	l := "rover.log"
	h, err := GetHostName()
	if err != nil {
		return fmt.Sprintf("Cannot get system hostname with error %v", err)
	}
	i.HostName = h
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
	v := ""
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})

	pid, err := CheckProc(name)
	if err != nil {
    	fmt.Println("Cannot check proc")
    	//os.Exit(1)
    }
	if pid != "" {
		logger.Info("check-hashi-version", "process identified", name, "pid", pid)
		path, err := exec.LookPath(name)
		if err != nil {
			logger.Info("check-hashi-version", "cannot find binary in PATH", name)
		}
		if name == "consul" {
			v, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | head -n 1 | awk '{print $2}'", path)).Output()
			if err != nil {
				logger.Error("check-hashi-version", "cannot execute binary", name, "error", err.Error())
			}
			return string(v)
		} else if name == "nomad" || name == "vault" {
			v, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s version | awk '{print $2}'", path)).Output()
			if err != nil {
				logger.Error("check-hashi-version", "cannot execute binary", name, "error", err.Error())
			}
			return string(v)
		}
	}
	return v
}

// Dump takes a type, output filename and command, which it then executes
// while also writing stdout + stderr to a file named for the command
// Inspired by debug-ninja! (https://github.com/fprimex/debug-ninja)
func Dump(dumpType string, outfile string, cmdName string, args ...string) int {
	i := Internal{}
	t := "%s/%s/%s.txt"
	h, err := GetHostName()
	if err != nil {
		fmt.Println("Cannot get system hostname")
		os.Exit(1)
	}
	i.HostName = h
	// Internal logging
	l := "rover.log"
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})

	path, err := exec.LookPath(cmdName)
	if err != nil {
		logger.Info("dump", "cannot find command in system PATH", cmdName)
	} else {
		logger.Debug("dump", "found command", cmdName, "location", path)
		// We audit all command parameters by specifying them explicitly as
		// Dump() invocations and not allowing any form of user input to
		// be passed in for most command parameter values, so subprocess
		// launching with strict parameters maybe not too terribly sad here
		cmd := exec.Command(cmdName, args...)
		f, err := os.Create(fmt.Sprintf(filepath.Join(".", t),
			i.HostName,
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
			logger.Error("dump", "cannot execute command with error", err.Error())
			panic(err)
		}
		// Not as cool as the dots and the Es, but it lets us know something
		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					cli := fmt.Sprintf("%s %s", cmdName, strings.Join(args[:], " "))
					logger.Error("dump", "command exited with non-zero status", cli, "exit-status", hclog.Fmt("%d",status.ExitStatus()))
				}
			} else {
				logger.Error("dump", "command wait error", err.Error())
			}
		}
	}
	return 0
}

// FileExist checks for a file's existence
func FileExist(fileName string) bool {
	i := Internal{}
 	// Internal logging
	l := "rover.log"
	h, err := GetHostName()
	if err != nil {
		fmt.Println("Cannot get system hostname")
		os.Exit(1)
	}
	i.HostName = h
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		logger.Info("file-exist", "file exists", fileName)
		return true
	}
	return false
}

// GetHostName gets the current system's network hostname
func GetHostName() (string, error) {
	h, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("cannot determine hostname with error %v", err)
	}
	return h, nil
}

// HTTPCmdCheck checks for curl or wget and use what is present
func HTTPCmdCheck() string {
	i := Internal{}
	// Internal logging
	l := "rover.log"
	h, err := GetHostName()
	if err != nil {
		fmt.Println("Cannot get system hostname")
		os.Exit(1)
	}
	i.HostName = h
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})
	path, err := exec.LookPath("curl")
	if err != nil {
		logger.Info("http-cmd-check", "curl was not found in system PATH")
		path, err = exec.LookPath("wget")
		if err != nil {
			logger.Info("http-cmd-check", "wget was not found in system PATH")
		} else {
			logger.Info("http-cmd-check", "wget is available at", path)
			httpCommand := "wget"
			return httpCommand
		}
	}
	logger.Info("http-cmd-check", "curl is available at", path)
	httpCommand := "curl"
	return httpCommand
}

// ZipIt archives rover results into a zip file suitable for tubing
func ZipIt(target string) {
	i := Internal{}
	// Internal logging
	l := "rover.log"
	h, err := GetHostName()
	if err != nil {
		fmt.Println("Cannot get system hostname")
		os.Exit(1)
	}
	i.HostName = h
	p := filepath.Join(fmt.Sprintf("%s", i.HostName), "log")
    if err := os.MkdirAll(p, os.ModePerm); err != nil {
		fmt.Println(fmt.Sprintf("Cannot create log directory %s.", p))
		os.Exit(1)
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to open log file %s with error: %v", f, err))
		os.Exit(1)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
    logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})
	outpath := filepath.Join(".", i.HostName)
	err = zip.ArchiveFile("output", outpath, nil)
	if err != nil {
		logger.Error("zipit", "cannot archive with error", err.Error())
		os.Exit(1)
	}
	// Remove the source directory after zip successfully created
	err = os.RemoveAll(i.HostName)
	if err != nil {
		logger.Error("zipit", "cannot clean up with error", err.Error())
		os.Exit(1)
	}
}
