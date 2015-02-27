// +build slurp

package main

//Anything, even main.

import (
	"time"

	"github.com/omeid/slurp"
)

func Slurp(b *slurp.Build) {
	b.Task("turtle", nil, func(c *slurp.C) error {
		c.Info("Hello!")
		c.Warn("I will take at least 4 second unless cancled.")
		select {
		case <- c.Done():
		  c.Warn("Got cancel, leaving.")
		  return nil
		case <- time.Tick(4 * time.Second):
		  c.Info("I am done!")
		  return nil
		}

		c.Info("You shouldn't see this!")
		return nil
	})

	b.Task("medic", nil, func(c *slurp.C) error {
		c.Info("I might be able to help, Go ahead.")
		return nil
	})

	b.Task("rabbit", []string{"medic"}, func(c *slurp.C) error {
		c.Info("Hello, I am the the fast one.")

		c.Info("I take at least 4 seconds after cancel.")
		for i := 0; i < 4; i++ {
			c.Infof("This is the %d line of my work.", i)
			time.Sleep(500 * time.Millisecond)
		}
		<- c.Done() 
		c.Info("Got cancel but nope. I am not leaving.")
		time.Sleep(4*time.Second)
		return nil
	})

	b.Task("default", []string{"turtle", "rabbit"}, func(c *slurp.C) error {
		//This task is run when slurp is called with any task parameter.
		c.Info("Default task is running.")
		return nil
	})
}
