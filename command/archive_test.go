package command

import (
	"github.com/mitchellh/cli"
	"testing"
)

func TestArchive(t *testing.T) {
	ui := new(cli.MockUi)
	testHost := "rover-test"
	testFile := "rover-test-%s-%s.zip"

	args := []string{}

	c := &ArchiveCommand{
		UI:         ui,
		HostName:   testHost,
		TargetFile: testFile,
	}

	if code := c.Run(args); code != 0 {
		t.Fatalf("archive failed: %d\n\n%s", code, ui.ErrorWriter.String())
	}
}
