// Package command for archive
// Archive compresses command output into a zip file
package command

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	"github.com/pierrre/archivefile/zip"

)

const (
	// Default to storing archive files in the PWD vs. TMPDIR
	archivePathDefault = "./"
	archivePathDescr   = "Archive file path"
)

// ArchiveCommand describes common zip file fields
type ArchiveCommand struct {
	ArchivePath string
	HostName    string
	OS          string
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

	logger.Info("archive", "hello from", c.HostName)
    logger.Info("archive", "detected OS", c.OS)

	cmdFlags := flag.NewFlagSet("archive", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	cmdFlags.StringVar(&c.ArchivePath, "path", archivePathDefault, archivePathDescr)
	cmdFlags.BoolVar(&c.KeepSrc, "keep-source", false, "Remove the zipfile source directory?")
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	h, err := GetHostName()
	if err != nil {
		out := fmt.Sprintf("Cannot get system hostname with error %v", err)
		c.UI.Output(out)

		return 1
	}
	c.HostName = h
	c.TargetFile = "rover-%s-%s.zip"
	t := time.Now().Format("20060102150405")
	archiveFileName := fmt.Sprintf(c.TargetFile, c.HostName, t)

	defer func() {
		// Remove the source directory after zip file created
		if !c.KeepSrc {
			err := os.RemoveAll(c.HostName)
			if err != nil {
				logger.Error("cannot remove source directory with error", err.Error())
				fmt.Println(fmt.Sprintf("\nCannot remove source directory with error: %v", err))
				os.Exit(1)
			}
		}
		logger.Info("archive", "preserved source directory in", c.HostName)
	}()

	_, err = os.Stat(c.HostName)
	if os.IsNotExist(err) {
        fmt.Println(fmt.Sprintf("Cannot archive nonexistent directory '%s'; please use rover commands to generate data first.", c.HostName))
        os.Exit(1)
    }
	outPath := filepath.Join(c.ArchivePath, archiveFileName)

    // Shout out to Ye Olde School BSD spinner!
	roverSpinnerSet := []string{"/", "|", "\\", "-", "|", "\\", "-"}
	s := spinner.New(roverSpinnerSet, 174*time.Millisecond)
	s.Writer = os.Stderr
	err = s.Color("fgHiCyan")
	if err != nil {
		log.Printf("install", "weird-error", err.Error())
	}
	s.Suffix = " Archiving data, please wait ..."
	s.Start()
	err = zip.ArchiveFile(fmt.Sprintf("%s", c.HostName), outPath, nil)
	if err != nil {
		logger.Error("cannot archive data with error", err.Error())
		fmt.Println(fmt.Sprintf("\nCannot archive data with error: %v", err))
		os.Exit(1)
	}
	// out := fmt.Sprintf("Data archived in %s", outPath)
	// c.UI.Output(out)

	s.FinalMSG = fmt.Sprintf("Data archived in %s\n", outPath)
	s.Stop()

	return 0
}

// Synopsis output
func (c *ArchiveCommand) Synopsis() string {
	return "Archive rover data into zip file"
}
