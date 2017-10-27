// Package command for Consul
// ConsulCommand executes commands with the consul command and stores output
package command

import (
	"fmt"
	"github.com/brianshumate/rover/internal"
	"github.com/mitchellh/cli"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConsulCommand describes Consul related fields
type ConsulCommand struct {
	ConsulDa       bool
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

// Run consl commands
func (c *ConsulCommand) Run(_ []string) int {

	// Internal logging
	internal.LogSetup()

	c.ConsulDa, c.ConsulPID = internal.CheckProc("consul")
	c.OS = runtime.GOOS
	c.HostName = internal.GetHostName()
	c.HTTPCommand = internal.HTTPCmdCheck()
	c.HTTPTokenValue = os.Getenv("CONSUL_HTTP_TOKEN")

	log.Printf("[i] Hello from the rover Consul module on %s!", c.HostName)

	// Command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/consul", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("[e] Cannot create directory %s!", outPath)
		log.Println(out)
		c.UI.Error(out)
		os.Exit(-1)
	}

	// Drop a note about CONSUL_HTTP_TOKEN (zero length == unset)
	log.Printf("[i] CONSUL_HTTP_TOKEN length: %v", len(c.HTTPTokenValue))

	// Dump commands only if a running Consul process is detected
	if c.ConsulDa {
		log.Printf("[i] Consul processs identified as %s", c.ConsulPID)
		internal.Dump("consul", "consul_version", "consul", "version")
		internal.Dump("consul", "consul_info", "consul", "info")
		internal.Dump("consul", "consul_members", "consul", "members")
		internal.Dump("consul",
			"consul_operator_raft_list_peers",
			"consul",
			"operator",
			"raft",
			"list-peers")

		// Current process limit information and open files for process
		// when Linux and PID determined
		if c.OS == Linux && c.ConsulPID != "ENOIDEA" {
			internal.Dump("consul", "proc_consul_limits", "cat", fmt.Sprintf("/proc/%s/limits", c.ConsulPID))
			internal.Dump("consul", "proc_consul_open_file_count", "sh", "-c", fmt.Sprintf("ls /proc/%s/fd | wc -l", c.ConsulPID))
		}

		// Check syslog output locations for supported systems
		switch c.OS {

		case Darwin:
			log.Println("[i] Attempting extraction of Consul log messages from system log (sudo required) ...")
			internal.Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/system.log")

		case FreeBSD, Linux:
			// Grep for "consul" in /var/log/messages or /var/log/syslog (sudo required)
			log.Println("[i] Attempting extraction of Consul log messages from system logs (sudo required) ...")
			if internal.FileExist("/var/log/syslog") {
				log.Println("[i] Checking /var/log/syslog for Consul entries (sudo required) ...")
				internal.Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/syslog")
			} else {
				log.Println("[i] No /var/log/syslog found, checking /var/log/messages for Consul entries (sudo required) ...")
				internal.Dump("consul", "consul_syslog", "grep", "-w", "consul", "/var/log/messages")
			}
		}

		// Attempt to grab full Consul gorountine stack dump and heap dump
		tokenHeader := fmt.Sprintf("X-Consul-Token: %s", c.HTTPTokenValue)

		if c.HTTPCommand == "curl" {
			log.Print("[i] Attempting goroutine dump with curl (requires ACL token)")
			internal.Dump("consul", "consul_goroutine", "curl", "-s", "--header", tokenHeader, "localhost:8500/debug/pprof/goroutine?debug=2")
			log.Print("[i] Attempting heap dump with curl (requires ACL token)")
			internal.Dump("consul", "consul_heap", "curl", "-s", "--header", tokenHeader, "localhost:8500/debug/pprof/heap?debug=1")
		} else if c.HTTPCommand == "wget" {
			log.Print("[i] Attempting goroutine dump with wget (requires ACL token)")
			internal.Dump("consul", "consul_goroutine", "wget", "--header", tokenHeader, "-qO-", "localhost:8500/debug/pprof/gorpoutine?debug=2")
			log.Print("[i] Attempting heap dump with wget (requires ACL token)")
			internal.Dump("consul", "consul_heap", "wget", "--header", tokenHeader, "-qO-", "localhost:8500/debug/pprof/heap?debug=1")
		} else {
			log.Print("[w] Could not detect curl or wget in this environment")
		}

	} else {
		// Danger Will Robinson! No Consul detected!
		log.Println("[w] No consul process detected in this environment")
	}

	c.UI.Output("Executed Consul commands and stored output")

	return 0
}

// Synopsis output
func (c *ConsulCommand) Synopsis() string {
	return "Execute Consul related commands and store output"
}
