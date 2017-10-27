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
	HostName string
	Uptime   string
	UI       cli.Ui
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

	columns := []string{}
	kvs := map[string]string{"OS": runtime.GOOS, "Architecture": runtime.GOARCH}
	for k, v := range kvs {
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
