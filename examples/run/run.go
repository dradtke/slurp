// +build slurp

package run

import "github.com/omeid/slurp"

func Slurp(b *slurp.Build) {

	b.Task("a", nil, func(c *slurp.C) error {
		return nil
	})

	b.Task("b", nil, func(c *slurp.C) error {
		return nil
	})

	b.Task("c", nil, func(c *slurp.C) error {
		return nil
	})

	b.Task("default", []string{"a", "b", "c"} , func(c *slurp.C) error {
		b.Warn("Calling locally.")
		b.Run(c, "a", "b", "c")
		b.Warn("Calling One by one.")
		b.Run(c,"a")
		b.Run(c,"b")
		b.Run(c,"c")

		return nil
	})
}
