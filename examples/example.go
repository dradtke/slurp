// +build slurp

package main //Anything, even main.

import (
	"time"

	"github.com/omeid/slurp"
)

func Slurp(b *slurp.Build) {
	b.Task("slow", nil, func(c *slurp.C) error {
		c.Info("Hello!")
		c.Warn("I will take at least 3 seconds.")
		time.Sleep(3 * time.Second)
		c.Error("This is a phony error.")
		return nil
	})

	b.Task("rabbit", nil, func(c *slurp.C) error {
		c.Info("Hello, I am the the fast one.")
		for i := 0; i < 6; i++ {
			c.Printf("This is the %d line of my work.", i)
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	})

	b.Task("default", []string{"slow", "rabbit"}, func(c *slurp.C) error {
		//This task is run when slurp is called with any task parameter.
		c.Info("Hello!")
		return nil
	})
}
