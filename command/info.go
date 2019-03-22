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
	OS            string
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

	// System data / random system factoids
	systemData := map[string]string{"OS": runtime.GOOS,
		"Architecture": runtime.GOARCH}
	t := time.Now()
	systemData["Date/Time"] = t.Format("Mon Jan _2 15:04:05 2006")

	systemCols := []string{}
	for k, v := range systemData {
		systemCols = append(systemCols, fmt.Sprintf("%s: | %s ", k, v))
	}

	systemOut := columnize.SimpleFormat(systemCols)

	// Version data for actively running binaries
	versionData := map[string]string{}
	if c.ConsulVersion != "" {
		versionData["Consul version"] = c.ConsulVersion
	}

	if c.NomadVersion != "" {
		versionData["Nomad version"] = c.NomadVersion
	}

	if c.VaultVersion != "" {
		versionData["Vault version"] = c.VaultVersion
	}

	versionCols := []string{}
	for k, v := range versionData {
		versionCols = append(versionCols, fmt.Sprintf("%s: | %s ", k, v))
	}

	versionOut := columnize.SimpleFormat(versionCols)

	so := fmt.Sprintf("Basic factoids about this system:\n\n%s\n\n", systemOut)
	vo := fmt.Sprintf("Active running versions:\n\n%s\n\n", versionOut)

	if versionOut == "" {
		out := fmt.Sprintf("%s", so)
		c.UI.Output(out)
	} else {
		out := fmt.Sprintf("%s%s", so, vo)
		c.UI.Output(out)
	}

	return 0
}

// Synopsis output
func (c *InfoCommand) Synopsis() string {
	return "Output basic system factoids"
}
