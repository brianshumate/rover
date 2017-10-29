package main

import (
	"fmt"
	"github.com/brianshumate/rover/command"
	"github.com/mitchellh/cli"
	"os"
)

func main() {

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c := cli.NewCLI("rover", "0.0.4")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"archive": func() (cli.Command, error) {
			return &command.ArchiveCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorGreen,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
		"consul": func() (cli.Command, error) {
			return &command.ConsulCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorGreen,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
		"info": func() (cli.Command, error) {
			return &command.InfoCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorNone,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
		"nomad": func() (cli.Command, error) {
			return &command.NomadCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorGreen,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
		"system": func() (cli.Command, error) {
			return &command.SystemCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorGreen,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
		// upload is a WIP
		"upload": func() (cli.Command, error) {
			return &command.UploadCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					OutputColor: cli.UiColorGreen,
				},
			}, nil
		},
		"vault": func() (cli.Command, error) {
			return &command.VaultCommand{
				UI: &cli.ColoredUi{
					Ui:          ui,
					ErrorColor:  cli.UiColorRed,
					InfoColor:   cli.UiColorCyan,
					OutputColor: cli.UiColorGreen,
					WarnColor:   cli.UiColorYellow,
				},
			}, nil
		},
	}

	// Initial subcommand autocompletion
	// use `rover -autocomplete-install` to activate then open a new shell
	c.Autocomplete = true

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	os.Exit(exitStatus)
}
