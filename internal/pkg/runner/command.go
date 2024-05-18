package runner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/huboh/gwatch/internal/pkg/utils"
)

// Command represents a command to be executed.
type Command struct {
	// cmd is the underlying exec.Cmd instance.
	cmd *exec.Cmd

	// args is the command arguments.
	args []string

	// done is a channel to signal completion or termination of the command.
	done chan struct{}

	// cmdLock Mutex prevent concurrent access to the underlying command.
	cmdLock *sync.RWMutex
}

// NewCommand creates a new Command instance pointer with the provided arguments.
//
// Use Run method to execute the command.
// Use Kill method to terminate the running command.
func NewCommand(args []string) *Command {
	return &Command{
		args:    args,
		cmdLock: new(sync.RWMutex),
	}
}

// Run starts the command and waits for it to finish.
func (c *Command) Run(stdout io.Writer, stderr io.Writer) error {
	var err error

	// checks if theres an active cmd instance in a diff goroutine
	if c.IsActive() {
		close(c.done)
	}

	// prevent other goroutine from resetting cmd while we're still active
	c.cmdLock.Lock()

	// new cmd
	c.cmd = exec.Command(c.args[0], c.args[1:]...)
	c.done = make(chan struct{})

	// pipe output from cmd process
	c.PipeStdErr(stderr)
	c.PipeStdOut(stdout)

	// only unlock when we exit. we exit either the process
	// ran to completion or we kill it explicitly
	defer func() {
		// reset
		c.cmd = nil
		c.done = nil
		c.cmdLock.Unlock()
	}()

	// start cmd
	if err = c.cmd.Start(); err != nil {
		return err
	}

	select {
	// TODO: add timeout case if Wait takes too long

	// Wait for the command to finish.
	case err = <-utils.AsyncResult(c.cmd.Wait):
		break

	// If the command is unexpectedly killed, send a kill signal to the process.
	case <-c.done:
		if err = c.cmd.Process.Signal(os.Kill); err != nil {
			return err
		}

		if err = c.cmd.Wait(); err != nil {
			// we killed it so any exit code of not 0 is ignored
			if _, isExitErr := err.(*exec.ExitError); !isExitErr {
				return err
			}
		}
	}

	return err
}

// Kill terminates the command and it's underlying process if it is still running.
func (c *Command) Kill() error {
	if c.IsActive() {
		return c.cmd.Process.Signal(os.Kill)
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

		for stdOutScanner.Scan() {
			fmt.Fprintln(stdout, "stdout:", stdOutScanner.Text())
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

		for stdErrScanner.Scan() {
			fmt.Fprintln(stderr, "stderr:", stdErrScanner.Text())
		}
	}()

	return nil
}
