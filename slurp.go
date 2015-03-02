package slurp

import (
	"github.com/omeid/slurp/log"
)

type C struct {
	log.Log
	done <-chan struct{}
}

func (c *C) New(prefix string) *C {
  return &C{c.Log.New(prefix), c.done}
}

// Done returns a channel that's closed when the current build is
// canceled. You should return as soon as possible.
// Successive calls to Done return the same value.
func (c  *C) Done() <-chan struct{} {
  return c.done
}

// A stage where a series of files goes for transformation, manipulation.
// There is no correlation between a stages input and output, a stage may
// decided to pass the same files after transofrmation or generate new files
// based on the input.

type Stage func(<-chan File, chan<- File)

func (stage Stage) pipe(in <-chan File) Pipe {
	out := make(chan File)
	go func() {
		stage(in, out)
		close(out)
	}()

	return out
}

//Pipe is a channel of Files.
type Pipe <-chan File

// Pipes the current Channel to the give list of Stages and returns the
// last jobs otput pipe.
func (p Pipe) Pipe(stages ...Stage) Pipe {
	switch len(stages) {
	case 0:
		return p
	case 1:
		return stages[0].pipe(p)
	default:
		return stages[0].pipe(p).Pipe(stages[1:]...)
	}
}

// Waits for the end of channel and closes all the files.
func (p Pipe) Wait() error {
	var err error
	for f := range p {
		e := f.Close()
		if err == nil && e != nil {
			err = e
		}
	}
	return err
}

//This is a combination of p.Pipe(....).Wait()
func (p Pipe) Then(stages ...Stage) error {
	return p.Pipe(stages...).Wait()
}
