package filter

import (

	"github.com/omeid/slurp"
	"github.com/omeid/slurp/tools/glob"
)

//A Filter stage, will either close or pass files to the next
// slurp.Stage based on the output of the `filter` function.
func FilterFunc(c *slurp.C, filter func(slurp.File) bool) slurp.Stage {
	return func(files <-chan slurp.File, out chan<- slurp.File) {
		for f := range files {
			if filter(f) {
				f.Close()
			} else {
				out <- f
			}
		}
	}
}

// A simple transformation slurp.Stage, sends the file to output
// channel after passing it through the the "do" function.
func DoFunc(c *slurp.C, do func(*slurp.C, slurp.File) slurp.File) slurp.Stage {
	return func(files <-chan slurp.File, out chan<- slurp.File) {
		for f := range files {
			out <- do(c, f)
		}
	}
}

//For The Glory of Debugging.
func List(c *slurp.C) slurp.Stage {
	return DoFunc(c, func(c *slurp.C, f slurp.File) slurp.File {
		s, err := f.Stat()
		if err != nil {
			c.Print("Can't get file name.")
		} else {
			c.Printf("slurp.File: %+v Name: %s", f, s.Name())
		}
		return f
	})
}

//Filters out files based on a pattern, if they match,
// they will be closed, otherwise sent to the output channel.
func Filter(c *slurp.C, pattern string) slurp.Stage {
	return FilterFunc(c, func(f slurp.File) bool {
		s, err := f.Stat()
		if err != nil {
			c.Print("Can't get file name.")
			return false
		}
		m, err := glob.Match(pattern, s.Name())
		if err != nil {
			c.Println(err)
		}
		return m
	})
}
