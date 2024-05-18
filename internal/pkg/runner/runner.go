// Package runner provides functionality for executing our build and run commands.
package runner

import (
	"os"
	"strings"

	"github.com/huboh/gwatch/internal/pkg/config"
)

// Runner represents a runner for building and running go applications.
type Runner struct {
	// buildCmd is the build command to be executed
	buildCmd *Command

	// runBuildCmd is the command to run the compiled binary
	runBuildCmd *Command
}

// New creates a new `*Runner` instance with the given configuration.
func New(config config.Config) *Runner {
	return &Runner{
		buildCmd:    NewCommand(strings.Split(config.Build.Cmd, "\x20")),
		runBuildCmd: NewCommand(append([]string{config.Run.Bin}, config.Run.Args...)),
	}
}

// Kill kills builds and runs process.
func (r *Runner) Kill() error {
	if err := r.buildCmd.Kill(); err != nil {
		return err
	}

	if err := r.runBuildCmd.Kill(); err != nil {
		return err
	}

	return nil
}

// Launch builds and runs the application.
func (r *Runner) Launch() error {
	if err := r.buildCmd.Run(os.Stdout, os.Stderr); err != nil {
		return err
	}

	if err := r.runBuildCmd.Run(os.Stdout, os.Stderr); err != nil {
		return err
	}

	return nil
}
