package slurp


import (
  "github.com/omeid/slurp/log"
)

type C struct {
	log.Log
	done <-chan struct{}
}

func (c *C) New(prefix string) *C {
  return &C{Log: c.Log.New(prefix), done: c.done}
}

// Done returns a channel that's closed when the current build is
// canceled. You should return as soon as possible.
// Successive calls to Done return the same value.
func (c  *C) Done() <-chan struct{} {
  return c.done
}
