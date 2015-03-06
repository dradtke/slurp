// build slurp

//Bare minimum Slurp file.
package minimal

import "github.com/omeid/slurp"

func Slurp(b *slurp.Build) {
	b.Task(slurp.Task{
	  Name: "default", 
	  Action: func(c *slurp.C) error {
		return nil
	}})
}
