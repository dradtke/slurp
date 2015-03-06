// +build slurp

package run

import "github.com/omeid/slurp"

func Slurp(b *slurp.Build) {

	b.Task(
		slurp.Task{
			Name: "a",
			Action: func(c *slurp.C) error {
				return nil
			},
		},

		slurp.Task{
			Name: "b",
			Action: func(c *slurp.C) error {
				return nil
			},
		},

		slurp.Task{
			Name: "c",
			Action: func(c *slurp.C) error {
				return nil
			},
		},

		slurp.Task{
			Name: "default",
			Deps: []string{"a", "b", "c"},
			Action: func(c *slurp.C) error {

				b.Warn("Calling locally.")
				b.Run(c, "a", "b", "c")
				b.Warn("Calling One by one.")
				b.Run(c, "a")
				b.Run(c, "b")
				b.Run(c, "c")
				return nil
			},
		},
	)

}
