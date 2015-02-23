package util

import (
	"bytes"

	"github.com/omeid/slurp"
)

// Concatenates all the files from the input channel
// and passes them to output channel with the given name.
func Concat(c *slurp.C, name string) slurp.Stage {
	return func(files <-chan slurp.File, out chan<- slurp.File) {

		var (
			size    int64
			bigfile = new(bytes.Buffer)
		)

		for f := range files {
			n, err := bigfile.ReadFrom(f)
			if err != nil {
				c.Println(err)
				return
			}
			bigfile.WriteRune('\n')
			size += n + 1

			f.Close()
		}

		sf := slurp.File{
			Reader: bigfile,
			Dir:    "",
			Path:   name,
		}
		stat := &slurp.FileInfo{}
		stat.SetSize(size)
		stat.SetName(name)

		sf.SetStat(stat)

		out <- sf
	}
}
