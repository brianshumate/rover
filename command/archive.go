// Package command for archive
// Archive compresses command output into a zip file
package command

import (
	"flag"
	"fmt"
	"github.com/brianshumate/rover/internal"
	"github.com/mitchellh/cli"
	"github.com/pierrre/archivefile/zip"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	archivePathDefault = "/tmp"
	archivePathDescr   = "Archive file path"
)

// ArchiveCommand describes common zip file fields
type ArchiveCommand struct {
	ArchivePath string
	HostName    string
	KeepSrc     bool
	TargetFile  string
	UI          cli.Ui
}

// Help output
func (c *ArchiveCommand) Help() string {
	helpText := `
Usage: rover archive
	Archive the command output directory into a zip file with this
	filename format: rover-<hostname>-<timestamp>.zip

General Options:
  -keep-source	Whether to keep the archive source directory [default: false]
  -path		Path where archive file is written [default: "/tmp"]
`

	return strings.TrimSpace(helpText)
}

// Run command
func (c *ArchiveCommand) Run(args []string) int {

	// Internal logging
	internal.LogSetup()

	cmdFlags := flag.NewFlagSet("archive", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	cmdFlags.StringVar(&c.ArchivePath, "path", archivePathDefault, archivePathDescr)
	cmdFlags.BoolVar(&c.KeepSrc, "keep-source", false, "Remove the zipfile source directory?")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	c.HostName = internal.GetHostName()
	c.TargetFile = "rover-%s-%s.zip"
	t := time.Now().Format("20060102150405")
	archiveFileName := fmt.Sprintf(c.TargetFile, c.HostName, t)

	defer func() {
		// Remove the source directory after zip file created
		if !c.KeepSrc {
			err := os.RemoveAll(c.HostName)
			if err != nil {
				log.Println("[e] Could not remove source directory!")
				panic(err)
			}
		}
		log.Printf("[i] Preserved source directory in %s.", c.HostName)
	}()

	outPath := filepath.Join(c.ArchivePath, archiveFileName)
	err := zip.ArchiveFile(fmt.Sprintf("%s", c.HostName), outPath, nil)
	if err != nil {
		log.Println("[e] Could not archive data!")
		panic(err)
	}
	out := fmt.Sprintf("Data archived in %s", outPath)
	c.UI.Output(out)

	return 0
}

// Synopsis output
func (c *ArchiveCommand) Synopsis() string {
	return "Archive rover data into zip file"
}
