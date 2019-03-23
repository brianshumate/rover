// Package command for HashiCorp's Consul https://consul.io/
// VaultCommand executes commands with the consul command line internally
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

// ConsulCommand describes Consul related fields
type ConsulCommand struct {
	ConsulPID      string
	HostName       string
	HTTPCommand    string
	HTTPTokenValue string
	OS             string
	OutputPath     string
	UI             cli.Ui
}

// Help output
func (c *ConsulCommand) Help() string {
	helpText := `
Usage: rover consul
	Execute a series of Consul related commands and store output in text files
`

	return strings.TrimSpace(helpText)
}

// Run consul commands
func (c *ConsulCommand) Run(_ []string) int {
	c.OS = runtime.GOOS
	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)
		return 1
	}
	c.HostName = h
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

	logger.Info("consul", "hello from the Consul module at", c.HostName)
	logger.Info("consul", "our detected OS", c.OS)

	p, err = CheckProc("consul")
	if err != nil {
		logger.Info("consul", "consul process not detected:", err.Error())
		out := "Consul process not detected in this environment."
		c.UI.Warn(out)
		return 1
	}
	c.ConsulPID = p
	c.HTTPCommand = HTTPCmdCheck()
	c.HTTPTokenValue = os.Getenv("CONSUL_HTTP_TOKEN")
	// Command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/consul", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("cannot create directory %s", outPath)
		logger.Info("consul", out)
		c.UI.Error(out)
		return 1
	}
	// Drop a note about CONSUL_HTTP_TOKEN (zero length == unset)
	logger.Warn("vault", "CONSUL_HTTP_TOKEN length", hclog.Fmt("%v", len(c.HTTPTokenValue)))
	// Dump commands only if a running Consul process is detected
	if c.ConsulPID != "" {
		logger.Info("consul", "agent process identified", c.ConsulPID)

		// Shout out to Ye Olde School BSD spinner!
		roverSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
		s := spinner.New(roverSpinnerSet, 174*time.Millisecond)
		s.Writer = os.Stderr
		err := s.Color("fgHiCyan")
		if err != nil {
			logger.Warn("consul", "weird-error", err.Error())
		}
		s.Suffix = " Gathering Consul data ..."
		s.FinalMSG = "Gathered Consul data\n"
		s.Start()

		// Unauthenticated first...
		Dump("consul", "consul_version", "consul", "version")
		// These will fail most of the time unless running without ACL
		Dump("consul", "consul_info", "consul", "info")
		Dump("consul", "consul_members", "consul", "members")
		Dump("consul",
			"consul_operator_raft_list_peers",
			"consul",
			"operator",
			"raft",
			"list-peers")

		// Consul-specific operating system tasks based on host OS ID
		switch c.OS {
		case Darwin:
			logger.Info("attempt to extract consul log messages from system log (sudo required) ...")
			Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/system.log")
		case FreeBSD:
			// Grep for "consul" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("attempt to extact consul log messages from system logs (sudo required) ...")
			if FileExist("/var/log/syslog") {
				logger.Info("checking /var/log/syslog for consul entries (sudo required) ...")
				Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/syslog")
			} else {
				logger.Info("no /var/log/syslog found, checking /var/log/messages for consul entries (sudo required) ...")
				Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/messages")
			}
		case Linux:
			// Select process table information when Linux and PID determined
			Dump("consul", "proc_consul_limits", "cat", fmt.Sprintf("/proc/%s/limits", c.ConsulPID))
			Dump("consul", "proc_consul_status", "cat", fmt.Sprintf("/proc/%s/status", c.ConsulPID))
			Dump("consul", "proc_consul_open_file_count", "sh", "-c", fmt.Sprintf("ls /proc/%s/fd | wc -l", c.ConsulPID))
			// Grep for "consul" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("attempting to extract consul log messages from system logs (sudo required) ...")
			if FileExist("/var/log/syslog") {
				logger.Info("checking /var/log/syslog for consul entries (sudo required) ...")
				Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/syslog")
			} else {
				logger.Info("no /var/log/syslog found, checking /var/log/messages for consul entries (sudo required) ...")
				Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/messages")
			}
			if FileExist("/run/systemd/system") {
				logger.Info("consul", "attempting to gather Vault logging from systemd journal.")
				Dump("consul", "consul_journald", "journalctl", "-b", "--no-pager", "-u", "consul")
			}
		}
		// Attempt to grab full Consul gorountine stack dump and heap dump
		tokenHeader := fmt.Sprintf("X-Consul-Token: %s", c.HTTPTokenValue)
		if c.HTTPCommand == "curl" {
			logger.Info("attempting goroutine dump with curl (requires ACL token)")
			Dump("consul", "consul_goroutine", "curl", "-s", "--header", tokenHeader, "localhost:8500/debug/pprof/goroutine?debug=2")
			logger.Info("attempting heap dump with curl (requires ACL token)")
			Dump("consul", "consul_heap", "curl", "-s", "--header", tokenHeader, "localhost:8500/debug/pprof/heap?debug=1")
		} else if c.HTTPCommand == "wget" {
			logger.Info("attempting goroutine dump with wget (requires ACL token)")
			Dump("consul", "consul_goroutine", "wget", "--header", tokenHeader, "-qO-", "localhost:8500/debug/pprof/gorpoutine?debug=2")
			logger.Info("attempting heap dump with wget (requires ACL token)")
			Dump("consul", "consul_heap", "wget", "--header", tokenHeader, "-qO-", "localhost:8500/debug/pprof/heap?debug=1")
		} else {
			logger.Warn("cannot detect curl or wget in this environment")
		}
		s.Stop()
	} else {
		logger.Info("no consul details learned from this environment.")
	}

	return 0

}

// Synopsis output
func (c *ConsulCommand) Synopsis() string {
	return "Execute Consul related commands and store output"
}
