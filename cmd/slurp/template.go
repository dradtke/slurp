package main

import "text/template"

var runnerSrc = template.Must(template.New("main").Parse(`package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

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

	flag.Parse()
	slurp := slurp.NewBuild()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-interrupts
		// stop watches and clean up.
		slurp.Warnf("Captured %v, stopping build and exiting..\n", sig)
		go func() {
		  err := slurp.Stop() 
		  if err != nil {
			slurp.Error(err)
			os.Exit(1)
		  }
		  os.Exit(0)
		}()
		slurp.Warn("Press ctrl+c again to force exit.")
		<-interrupts
		os.Exit(1)

	}()

	client.Slurp(slurp)

	tasks := flag.Args()
	if len(tasks) == 0 {
		tasks = []string{"default"}
	}

	slurp.Infof("Running: %s", strings.Join(tasks, ","))
	slurp.Run(slurp.C, tasks...)
	slurp.Stop()

}
`))
