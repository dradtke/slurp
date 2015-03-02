package exec

// Package exec runs external commands. It wraps os/exec to make it
// easier to copy and kill a os/exec.Cmd

import (
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Cmd struct {
  *exec.Cmd
}

func Command(bin string, args ...string) *Cmd {
	return &Cmd{exec.Command(bin, args...)}
}

func (c *Cmd) New() *Cmd {
  return Command(c.Args[0], c.Args[1:]...)
}

func (c *Cmd) Kill() error {
	if c.Process != nil {
		done := make(chan error)
		go func() {
			c.Wait()
			close(done)
		}()
		//Trying a "soft" kill first
		var err error
		if runtime.GOOS == "windows" {
			err = c.Process.Kill()
		} else {
			err = c.Process.Signal(os.Interrupt)
		}
		if err != nil {
			return err
		}
		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			return c.Process.Kill()
		case <-done:
		}
	}
	return nil
}
