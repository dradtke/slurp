package main

import "text/template"

var runnerSrc = template.Must(template.New("main").Parse(`package main

import (
	"runtime"

	"github.com/omeid/slurp"

	client "{{ . }}"
)

func init() {
	maxprocs := runtime.NumCPU()
	if maxprocs > 2 {
		runtime.GOMAXPROCS(maxprocs / 2)
	}
}

func main() {
	slurp.Run(client.Slurp)
}
`))
