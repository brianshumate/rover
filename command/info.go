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
`

	return strings.TrimSpace(helpText)
}

// Run command
func (c *InfoCommand) Run(_ []string) int {
	// Internal logging
	internal.LogSetup()

	c.ConsulVersion = internal.CheckHashiVersion("consul")
	c.NomadVersion = internal.CheckHashiVersion("nomad")
	c.VaultVersion = internal.CheckHashiVersion("vault")

	infoData := map[string]string{"OS": runtime.GOOS,
		"Architecture": runtime.GOARCH}

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
	out := fmt.Sprintf("Handy factoids about this system:\n\n%s", data)
	c.UI.Output(out)
	return 0
}

// Synopsis output
func (c *InfoCommand) Synopsis() string {
	return "Output installation information"
}
