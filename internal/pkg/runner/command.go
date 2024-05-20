package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/huboh/gwatch/internal/pkg/utils"
)

// Command represents a command to be executed.
type Command struct {
	// cmd is the underlying exec.Cmd instance.
	cmd *exec.Cmd

	// cmdMemAccess Mutex prevent concurrent access to the underlying command.
	cmdMemAccess *sync.RWMutex

	// args is the command arguments.
	args []string

	// done is a channel to signal completion or termination of the command.
	done chan struct{}

	// outPrefix is the prefix to add to the commands output.
	outPrefix string
}

// NewCommand creates a new Command instance pointer with the provided arguments.
//
// Use Run method to execute the command.
// Use Kill method to terminate the running command.
func NewCommand(args []string, outPrefix string) *Command {
	return &Command{
		args:         args,
		outPrefix:    outPrefix,
		cmdMemAccess: new(sync.RWMutex),
	}
}

// Run starts the command and waits for it to finish.
func (c *Command) Run(stdout io.Writer, stderr io.Writer, onRun func()) error {
	// checks if there is an active cmd instance in a diff goroutine
	if c.IsActive() {
		utils.CloseSafely(c.done)
	}

	// prevent other goroutine from resetting cmd while we're still running
	c.cmdMemAccess.Lock()

	// new cmd
	c.cmd = exec.Command(c.args[0], c.args[1:]...)
	c.done = make(chan struct{})

	// pipe output from cmd process
	c.PipeStdErr(stderr)
	c.PipeStdOut(stdout)

	// only release mem access when we exit.
	defer func() {
		// reset
		c.cmd = nil
		c.done = nil
		c.cmdMemAccess.Unlock()
	}()

	if onRun != nil {
		onRun()
	}

	// start cmd
	if err := c.cmd.Start(); err != nil {
		return err
	}

	select {
	// kill cmd process
	case <-c.done:
		return c.Kill()

	// Wait for the cmd to finish or be interrupted.
	case err := <-utils.AsyncResult(c.cmd.Wait):
		if err != nil {
			// cmd was interrupted, so any exit error ignored
			if err, isExitErr := err.(*exec.ExitError); !isExitErr {
				return err
			}
		}
	}

	return nil
}

// Kill terminates the command and it's underlying process if it is still running.
func (c *Command) Kill() error {
	if c.IsActive() {
		if err := c.cmd.Process.Signal(os.Kill); err != nil {
			if !errors.Is(err, os.ErrProcessDone) {
				return err
			}
		}
	}

	return nil
}

// IsActive checks if the command is still running
func (c *Command) IsActive() bool {
	// this field is alway set when the process starts and nil when it exits
	return c.cmd != nil && c.cmd.Process != nil
}

func (c *Command) PipeStdOut(stdout io.Writer) error {
	stdOutPipe, err := c.cmd.StdoutPipe()

	if err != nil {
		return err
	}

	// continuously monitor launched stdout and prints to stdout
	go func() {
		defer stdOutPipe.Close()

		stdOutScanner := bufio.NewScanner(stdOutPipe)

		if c.outPrefix != "" && !strings.HasSuffix(c.outPrefix, ":") {
			c.outPrefix = c.outPrefix + ":"
		}

		for stdOutScanner.Scan() {
			fmt.Fprintln(stdout, c.outPrefix, stdOutScanner.Text())
		}
	}()

	return nil
}

func (c *Command) PipeStdErr(stderr io.Writer) error {
	stdErrPipe, err := c.cmd.StderrPipe()

	if err != nil {
		return err
	}

	// continuously monitor launched command stderr and print to stderr
	go func() {
		defer stdErrPipe.Close()

		stdErrScanner := bufio.NewScanner(stdErrPipe)

		if c.outPrefix != "" && !strings.HasSuffix(c.outPrefix, ":") {
			c.outPrefix = c.outPrefix + ":"
		}

		for stdErrScanner.Scan() {
			fmt.Fprintln(stderr, c.outPrefix, stdErrScanner.Text())
		}
	}()

	return nil
}
