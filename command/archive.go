// Package command for archive
// Archive compresses command output into a zip file
package command

import (
	"github.com/brianshumate/rover/internal"
	//"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/pierrre/archivefile/zip"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ArchiveCommand describes common zip file fields
type ArchiveCommand struct {
	HostName   string
	RemoveSrc  bool
	TargetFile string
	UI         cli.Ui
}

// Help output
func (c *ArchiveCommand) Help() string {
	helpText := `
Usage: rover archive
	Archive the command output directory into a zip file with this filename format: rover-<hostname>-<timestamp>.zip

`

	return strings.TrimSpace(helpText)
}

// Run command
func (c *ArchiveCommand) Run(_ []string) int {

	// Internal logging
	internal.LogSetup()

	/*
		    Beginning of a remove source dir opt out flag

			cmdFlags := flag.NewFlagSet("zip", flag.ContinueOnError)
			cmdFlags.Usage = func() { c.UI.Output(c.Help()) }

			cmdFlags.BoolVar(&c.RemoveSrc, "remove-source", false, "Remove the zipfile source directory?")
			if err := cmdFlags.Parse(args); err != nil {
				return 1
			}
	*/

	c.HostName = internal.GetHostName()
	c.TargetFile = "rover-%s-%s.zip"
	t := time.Now().Format("20060102150405")
	archiveFileName := fmt.Sprintf(c.TargetFile, c.HostName, t)

	defer func() {
		// Remove the source directory after zip file created
		err := os.RemoveAll(c.HostName)
		if err != nil {
			log.Println("[e] Could not remove source directory!")
			panic(err)
		}
	}()

	outPath := filepath.Join(".", archiveFileName)
	err := zip.ArchiveFile(fmt.Sprintf("%s", c.HostName), outPath, nil)
	if err != nil {
		log.Println("[e] Could not archive data!")
		panic(err)
	}

	out := fmt.Sprintf("Data archived in %s", archiveFileName)
	c.UI.Output(out)

	return 0
}

// Synopsis output
func (c *ArchiveCommand) Synopsis() string {
	return "Archive rover data into zip file"
}
