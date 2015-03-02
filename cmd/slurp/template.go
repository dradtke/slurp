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
	"fmt"

	"github.com/omeid/slurp"
	"github.com/omeid/slurp/log"

	client "{{ . }}"
)

func init() {

	maxprocs := runtime.NumCPU()
	if maxprocs > 2 {
		runtime.GOMAXPROCS(maxprocs / 2)
	}
}


var (
  level = flag.Int("timestamp", 0, "Log timestamp: 1-6.")
)

func main() {

	flag.Parse()
	log.Flags = *level

	build := slurp.NewBuild()

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-interrupts
		// stop watches and clean up.
		fmt.Println() //Next line
		build.Warnf("Captured %v, stopping build and exiting...", sig)
		build.Warn("Press ctrl+c again to force exit.")
		ret := 0
		select {
		case err := <- build.Cancel():
		  if err != nil {
			build.Error(err)
			build.Error("Cleaning up anyways.")
			ret = 1
		  }
		case <-interrupts:
		  fmt.Println() //Next line
		  build.Warn("Force exit.")
		  ret = 1
		}
		build.Cleanup()
		os.Exit(ret)

	}()

	client.Slurp(build)

	tasks := flag.Args()
	if len(tasks) == 0 {
		tasks = []string{"default"}
	}

	build.Infof("Running: %s", strings.Join(tasks, ","))
	build.Run(build.C, tasks...)
	build.Cleanup()
}
`))
