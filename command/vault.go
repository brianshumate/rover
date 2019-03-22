// Package command for HashiCorp's Vault https://vaultproject.io/
// VaultCommand executes commands with the vault command line internally
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
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/cli"
)

// VaultCommand describes Vault related fields
type VaultCommand struct {
	HostName        string
	OS              string
	UI              cli.Ui
	VaultPID        string
	VaultTokenValue string
	VaultVersion    string
}

// Help output
func (c *VaultCommand) Help() string {
	helpText := `
Usage: rover consul
	Execute Vault related commands and store output in text files
`

	return strings.TrimSpace(helpText)
}

// Run vault commands
func (c *VaultCommand) Run(_ []string) int {
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
	logger.Info("vault", "hello from the Vault module at", c.HostName)
	logger.Info("vault", "our detected OS", c.OS)
	p, err = CheckProc("vault")
	if err != nil {
		logger.Warn("vault", "vault process not detected:", err.Error())
		out := "Vault process not detected in this environment."
		c.UI.Warn(out)
		return 1
	}
	c.VaultPID = p
	c.VaultTokenValue = os.Getenv("VAULT_TOKEN")
	// Command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/vault", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("Cannot create directory %s", outPath)
		logger.Error("vault", "cannot create directory", outPath)
		c.UI.Error(out)
		return 1
	}
	// Drop a note about VAULT_TOKEN (zero length == unset)
	logger.Warn("vault", "VAULT_TOKEN length", hclog.Fmt("%v", len(c.VaultTokenValue)))
	// Dump commands only if running Vault server process detected
	if c.VaultPID != "" {
		// Shout out to Ye Olde School BSD spinner!
		roverSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
		s := spinner.New(roverSpinnerSet, 174*time.Millisecond)
		s.Writer = os.Stderr
		err := s.Color("fgHiCyan")
		if err != nil {
			logger.Warn("vault", "weird-error", err.Error())
		}
		s.Suffix = " Gathering Vault information ..."
		s.FinalMSG = "Executed Vault related commands and stored output\n"
		s.Start()
		c.VaultVersion = CheckHashiVersion("vault")
		v1, err := version.NewVersion(c.VaultVersion)
		if err != nil {
			logger.Error("vault", "version compare issue with error", err.Error())
			out := fmt.Sprintf("Version compare error %v", err)
			c.UI.Error(out)
			return 1
		}
		// Compare CLI syntax for differences introduced by the GREAT RENAMINING!
		v2, err := version.NewVersion("0.9.2")
		if err != nil {
			logger.Error("vault", "version constraint parsing issue with error", err.Error())
			out := fmt.Sprintf("Version constraint parsing issue with error %v", err)
			c.UI.Error(out)
			return 1
		}
		// Unauthenticated stuff first
		Dump("vault", "vault_version", "vault", "version")
		Dump("vault", "vault_status", "vault", "status")
		// These require a token
		if v1.GreaterThan(v2) {
			Dump("vault", "vault_audit_list", "vault", "audit", "list")
			Dump("vault", "vault_auth_methods", "vault", "auth", "list")
			Dump("vault", "vault_mounts", "vault", "secrets", "list")
		} else {
			Dump("vault", "vault_audit_list", "vault", "audit-list")
			Dump("vault", "vault_auth_methods", "vault", "auth", "-methods")
			Dump("vault", "vault_mounts", "vault", "mounts")
		}
		// Perform Vault-specific operating system tasks based on host OS ID
		switch c.OS {
		case Darwin:
			logger.Info("vault", "attempting to extract vault log messages from system log.")
			Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/system.log")
		case FreeBSD:
			// Grep for "vault" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("vault", "attempt to extract vault log messages from system logs (sudo required).")
			if FileExist("/var/log/syslog") {
				logger.Info("vault", "checking /var/log/syslog for vault entries (sudo required).")
				Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/syslog")
			} else {
				logger.Info("vault", "no /var/log/syslog found, checking /var/log/messages for vault entries (sudo required).")
				Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/messages")
			}
		case Linux:
			// Select process table information when Linux and PID determined
			Dump("vault", "proc_vault_limits", "cat", fmt.Sprintf("/proc/%s/limits", c.VaultPID))
			Dump("vault", "proc_vault_status", "cat", fmt.Sprintf("/proc/%s/status", c.VaultPID))
			Dump("vault", "proc_vault_open_file_count", "sh", "-c", fmt.Sprintf("ls /proc/%s/fd | wc -l", c.VaultPID))
			// Grep for "vault" in /var/log/messages or /var/log/syslog (sudo required)
			logger.Info("vault", "attempt to extract vault log messages from system logs (sudo required).")
			if FileExist("/var/log/syslog") {
				logger.Info("vault", "checking /var/log/syslog for vault entries (sudo required).")
				Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/syslog")
			} else {
				logger.Info("vault", "no /var/log/syslog found, checking /var/log/messages for vault entries (sudo required).")
				Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/messages")
			}
			if FileExist("/run/systemd/system") {
				logger.Info("vault", "attempting to gather Vault logging from systemd journal.")
				Dump("vault", "vault_journald", "journalctl", "-b", "--no-pager", "-u", "vault")
			}
		}
		s.Stop()
	} else {
		logger.Info("no vault details learned from this environment.")
	}

	return 0
}

// Synopsis output
func (c *VaultCommand) Synopsis() string {
	return "Execute Vault related commands and store output"
}
