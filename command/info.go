// Package command for info
// [WIP] Info executes commands and displays a mini dashboard
package command

import (
	"fmt"
	"github.com/brianshumate/rover/internal"
	"github.com/mitchellh/cli"
	"github.com/ryanuber/columnize"
	"runtime"
	"strings"
	"time"
)

// InfoCommand describes info dashboard related fields
type InfoCommand struct {
	ConsulVersion string
	HostName      string
	NomadVersion  string
	Uptime        string
	UI            cli.Ui
	VaultVersion  string
}

// Help output
func (c *InfoCommand) Help() string {
	helpText := `
Usage: rover info
	Provides current status on key details and versions for a system

Example output:

Basic factoids about this system:

OS:              darwin
Architecture:    amd64
Date/Time:       Sun Oct 29 12:53:43 2017
Consul version:  v0.9.3
Vault version:   v0.8.3
`

	return strings.TrimSpace(helpText)
}

// Run command
func (c *InfoCommand) Run(_ []string) int {
	// Internal logging
	internal.LogSetup()

	c.HostName = internal.GetHostName()
	c.ConsulVersion = internal.CheckHashiVersion("consul")
	c.NomadVersion = internal.CheckHashiVersion("nomad")
	c.VaultVersion = internal.CheckHashiVersion("vault")

	infoData := map[string]string{"OS": runtime.GOOS,
		"Architecture": runtime.GOARCH}
	t := time.Now()
	infoData["Date/Time"] = t.Format("Mon Jan _2 15:04:05 2006")

	if c.ConsulVersion != "ENOVERSION" {
		infoData["Consul version"] = c.ConsulVersion
	}

	if c.NomadVersion != "ENOVERSION" {
		infoData["Nomad version"] = c.NomadVersion
	}

	if c.VaultVersion != "ENOVERSION" {
		infoData["Vault version"] = c.VaultVersion
	}

	columns := []string{}
	for k, v := range infoData {
		columns = append(columns, fmt.Sprintf("%s: | %s ", k, v))
	}
	data := columnize.SimpleFormat(columns)
	out := fmt.Sprintf("Basic factoids about this system:\n\n%s", data)
	c.UI.Output(out)
	return 0
}

// Synopsis output
func (c *InfoCommand) Synopsis() string {
	return "Output basic system factoids"
}
