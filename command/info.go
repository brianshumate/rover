// Package command for info
// [WIP] Info executes commands and displays a mini dashboard
package command

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

    "github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"github.com/ryanuber/columnize"
)

// InfoCommand describes info dashboard related fields
type InfoCommand struct {
	ConsulVersion string
	HostName      string
	NomadVersion  string
	OS 			  string
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
	l := "rover.log"
	p := filepath.Join(fmt.Sprintf("%s", c.HostName), "log")
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

	logger.Info("system", "hello from", c.HostName)
    logger.Info("system", "detected OS", c.OS)

	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)

		return 1
	}
	c.HostName = h
	c.ConsulVersion = CheckHashiVersion("consul")
	c.NomadVersion = CheckHashiVersion("nomad")
	c.VaultVersion = CheckHashiVersion("vault")

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
