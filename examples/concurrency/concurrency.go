// +build slurp

package main

//Anything, even main.

import (
	"errors"
	"time"

	"github.com/omeid/slurp"
)

func Slurp(b *slurp.Build) {

	b.Task(
		slurp.Task{
			Name: "turtle",
			Action: func(c *slurp.C) error {
				c.Info("Hello!")
				c.Warn("I will take at least 3 seconds.")
				time.Sleep(3 * time.Second)
				c.Info("Well, here is a line.")
				return errors.New("I died.")
			},
		},

		slurp.Task{
			Name: "medic",
			Action: func(c *slurp.C) error {
				c.Info("I might be able to help, Go ahead.")
				return nil
			},
		},

		slurp.Task{
			Name: "rabbit",
			Deps: []string{"medic"},
			Action: func(c *slurp.C) error {
				c.Info("Hello, I am the the fast one.")
				for i := 0; i < 4; i++ {
					c.Infof("This is the %d line of my work.", i)
					time.Sleep(500 * time.Millisecond)
				}
				return nil
			},
		},

		slurp.Task{
			Name: "default",
			Deps: []string{"turtle", "rabbit"},
			Action: func(c *slurp.C) error {
				//This task is run when slurp is called with any task parameter.
				c.Info("Default task is running.")
				return nil
			},
		},
	)
}
