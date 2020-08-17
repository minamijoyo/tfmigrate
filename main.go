package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/minamijoyo/tfmigrate/command"
	"github.com/mitchellh/cli"
)

// Version is a version number.
var version = "0.0.1"

func main() {
	log.SetOutput(logOutput())
	log.Printf("[INFO] CLI args: %#v", os.Args)

	ui := &cli.BasicUi{
		Writer: os.Stdout,
	}
	commands := initCommands(ui)

	args := os.Args[1:]

	c := &cli.CLI{
		Name:       "tfmigrate",
		Version:    version,
		Args:       args,
		Commands:   commands,
		HelpWriter: os.Stdout,
	}

	exitStatus, err := c.Run()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to execute CLI: %s", err))
	}

	os.Exit(exitStatus)
}

func logOutput() io.Writer {
	levels := []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}
	minLevel := os.Getenv("TFMIGRATE_LOG")

	// default log writer is null device.
	writer := ioutil.Discard
	if minLevel != "" {
		writer = os.Stderr
	}

	filter := &logutils.LevelFilter{
		Levels:   levels,
		MinLevel: logutils.LogLevel(minLevel),
		Writer:   writer,
	}

	return filter
}

func initCommands(ui cli.Ui) map[string]cli.CommandFactory {
	meta := command.Meta{
		UI: ui,
	}

	commands := map[string]cli.CommandFactory{
		"plan": func() (cli.Command, error) {
			return &command.PlanCommand{
				Meta: meta,
			}, nil
		},
		"apply": func() (cli.Command, error) {
			return &command.ApplyCommand{
				Meta: meta,
			}, nil
		},
	}

	return commands
}
