// Package command for Nomad
// NomadCommand executes commands with the nomad command line internality
// and stores their output for HashiCorp's Nomad https://nomadproject.io/
package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/cli"
)

// NomadCommand describes Nomad related fields
type NomadCommand struct {
	HostName 	string
	OS       	string
	UI       	cli.Ui
	NomadPID 	string
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
	// Internal logging
	// LogSetup()

	p, err := CheckProc("nomad")
	if err != nil {
    	fmt.Println("Cannot find a consul")
    	os.Exit(1)
    }
    c.NomadPID = p
	c.OS = runtime.GOOS
	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)

		return 1
	}
	n.HostName = h

	log.Printf("[i] Hello from the rover Nomad module on %s!", n.HostName)

	// Handle creating the command output directory
	outPath := filepath.Join(".", fmt.Sprintf("%s/nomad", n.HostName))
	if err := os.MkdirAll(outPath, os.ModePerm); err != nil {
		log.Fatalf("[e] Cannot create directory %s.", outPath)
		panic(err)
	}

	// Drop a note about missing token
	nomadTokenValue := os.Getenv("VAULT_TOKEN")
	if len(nomadTokenValue) == 0 {
		log.Println("[i] No VAULT_TOKEN value detected in this environment")
	}

	// Dump commands only if running Nomad server process detected
	if c.NomadPID != "" {

		Dump("nomad", "nomad_status", "nomad", "status")
		Dump("nomad", "nomad_version", "nomad", "version")

		//Check syslog output locations for supported systems
		switch c.OS {

		case "darwin":
			log.Println("[i] Attempting extraction of Nomad log messages from system log ...")
			Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/system.log")

		case "freebsd", "linux":
			// Grep for "nomad" in /var/log/messages or /var/log/syslog (sudo required)
			log.Println("[i] Attempting extraction of Nomad log messages from system logs (sudo required) ...")
			if FileExist("/var/log/syslog") {
				log.Println("[i] Checking /var/log/syslog for Nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/syslog")
			} else {
				log.Println("[i] No /var/log/syslog found, checking /var/log/messages for Nomad entries (sudo required) ...")
				Dump("nomad", "nomad_syslog", "grep", "-w", "nomad", "/var/log/messages")
			}
		}
	} else {
		log.Println("[w] No nomad process detected in this environment")
	}

	out := "Executed Nomad commands and stored output"
	c.UI.Output(out)

	return 0
}

// Synopsis output
func (c *NomadCommand) Synopsis() string {
	return "Execute Nomad related commands and store output"
}
