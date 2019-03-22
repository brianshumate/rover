// Package command for HashiCorp's Nomad https://nomadproject.io/
// VaultCommand executes commands with the nomad command line internally
// and stores the output in plain text files
package command

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
)

// NomadCommand describes Nomad related fields
type NomadCommand struct {
	HostName string
	OS       string
	UI       cli.Ui
	NomadPID string
}

// Help output
func (c *NomadCommand) Help() string {
	helpText := `
Usage: rover nomad
	Execute a series of Nomad related commands and store output in text files
`

	return strings.TrimSpace(helpText)
}

// Run nomad commands
func (c *NomadCommand) Run(_ []string) int {
	n := NomadCommand{}
	c.OS = runtime.GOOS
	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)
		return 1
	}
	n.HostName = h
	// Internal logging
	l := "rover.log"
	p := filepath.Join(fmt.Sprintf("%s", c.HostName), "log")
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		out := fmt.Sprintf("Cannot create log directory %s.", p)
		c.UI.Error(out)
		return 1
	}
	f, err := os.OpenFile(filepath.Join(p, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		out := fmt.Sprintf("Failed to open log file %s with error: %v", f, err)
		c.UI.Error(out)
		return 1
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	logger := hclog.New(&hclog.LoggerOptions{Name: "rover", Level: hclog.LevelFromString("INFO"), Output: w})
	logger.Info("nomad", "hello from the Nomad module at", c.HostName)
	logger.Info("nomad", "our detected OS", c.OS)
	p, err = CheckProc("nomad")
	if err != nil {
		logger.Warn("nomad", "nomad process not detected:", err.Error())
		out := "Nomad process not detected in this environment."
		c.UI.Warn(out)
		return 1
	}
	c.NomadPID = p
	// Handle creating the command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/nomad", n.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		logger.Error("nomad", "cannot create directory", outPath, "error", err.Error())
		out := fmt.Sprintf("Cannot create directory %s with error %v", outPath, err)
		c.UI.Error(out)
		return 1
	}
	// Dump commands only if running Nomad server process detected
	if c.NomadPID != "" {
		// Shout out to Ye Olde School BSD spinner!
		roverSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
		s := spinner.New(roverSpinnerSet, 174*time.Millisecond)
		s.Writer = os.Stderr
		err := s.Color("fgHiCyan")
		if err != nil {
			logger.Warn("nomad", "weird-error", err.Error())
		}
		s.Suffix = " Gathering Nomad information ..."
		s.FinalMSG = "Executed Nomad related commands and stored output\n"
		s.Start()
		Dump("nomad", "nomad_status", "nomad", "status")
		Dump("nomad", "nomad_version", "nomad", "version")
		// Perform Nomad-specific operating system tasks based on host OS ID
		switch c.OS {
		case Darwin:
			logger.Info("attempt to extract nomad log messages from system log (sudo required) ...")
			Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/system.log")
		case FreeBSD:
			// Grep for "nomad" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("attempt to extact nomad log messages from system logs (sudo required) ...")
			if FileExist("/var/log/syslog") {
				logger.Info("checking /var/log/syslog for nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/syslog")
			} else {
				logger.Info("no /var/log/syslog found, checking /var/log/messages for nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/messages")
			}
		case Linux:
			// Select process table information when Linux and PID determined
			Dump("nomad", "proc_nomad_limits", "cat", fmt.Sprintf("/proc/%s/limits", c.NomadPID))
			Dump("nomad", "proc_nomad_status", "cat", fmt.Sprintf("/proc/%s/status", c.NomadPID))
			Dump("nomad", "proc_nomad_open_file_count", "sh", "-c", fmt.Sprintf("ls /proc/%s/fd | wc -l", c.NomadPID))
			// Grep for "nomad" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("attempting to extract nomad log messages from system logs (sudo required) ...")
			if FileExist("/var/log/syslog") {
				logger.Info("checking /var/log/syslog for nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/syslog")
			} else {
				logger.Info("no /var/log/syslog found, checking /var/log/messages for nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/messages")
			}
			if FileExist("/run/systemd/system") {
				logger.Info("nomad", "attempting to gather Vault logging from systemd journal.")
				Dump("nomad", "nomad_journald", "journalctl", "-b", "--no-pager", "-u", "nomad")
			}
		}
		s.Stop()
	} else {
		logger.Info("no nomad details learned from this environment")
	}

	return 0
}

// Synopsis output
func (c *NomadCommand) Synopsis() string {
	return "Execute Nomad related commands and store output"
}
