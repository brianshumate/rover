// Package command for Vault
// VaultCommand executes commands with the vault command line internality
// and stores their output for HashiCorp's Vault https://vaultproject.io/
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

// VaultCommand describes Vault related fields
type VaultCommand struct {
	HostName        string
	OS              string
	UI              cli.Ui
	VaultDa         bool
	VaultTokenValue string
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

	// Internal logging
	internal.LogSetup()

	c.VaultDa, _ = internal.CheckProc("consul")
	c.OS = runtime.GOOS
	c.HostName = internal.GetHostName()
	c.VaultTokenValue = os.Getenv("VAULT_TOKEN")

	log.Printf("[i] Hello from the rover Vault module on %s!", c.HostName)

	// Command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/vault", c.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		out := fmt.Sprintf("[e] Cannot create directory %s!", outPath)
		log.Println(out)
		c.UI.Error(out)
		os.Exit(-1)
	}

	// Drop a note about VAULT_TOKEN (zero length == unset)
	log.Printf("[i] VAULT_TOKEN length: %v", len(c.VaultTokenValue))

	// Dump commands only if running Vault server process detected
	if c.VaultDa {

		internal.Dump("vault", "vault_version", "vault", "version")
		internal.Dump("vault", "vault_audit_list", "vault", "audit-list")
		internal.Dump("vault", "vault_auth_methods", "vault", "auth", "-methods")
		internal.Dump("vault", "vault_mounts", "vault", "mounts")
		internal.Dump("vault", "vault_status", "vault", "status")

		// Check syslog output locations for supported systems
		switch c.OS {

		case Darwin:
			log.Println("[i] Attempting extraction of Vault log messages from system log ...")
			internal.Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/system.log")

		case FreeBSD, Linux:
			// Grep for "vault" in /var/log/messages or /var/log/syslog (sudo required)
			log.Println("[i] Attempting extraction of Vault log messages from system logs (sudo required) ...")
			if internal.FileExist("/var/log/syslog") {
				log.Println("[i] Checking /var/log/syslog for Vault entries (sudo required) ...")
				internal.Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/syslog")
			} else {
				log.Println("[i] No /var/log/syslog found, checking /var/log/messages for Vault entries (sudo required) ...")
				internal.Dump("vault", "vault_syslog", "grep", "-w", "vault", "/var/log/messages")
			}
		}
	} else {
		log.Println("[w] No vault process detected in this environment")
	}

	out := "Executed Vault commands and stored output"
	c.UI.Output(out)

	return 0
}

// Synopsis output
func (c *VaultCommand) Synopsis() string {
	return "Execute Vault related commands and store output"
}
