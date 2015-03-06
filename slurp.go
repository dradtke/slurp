package slurp

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var help = flag.Bool("help", false, "show help")

// Run setups a build and runs the listed tasks.
func Run(client func(b *Build)) {
	//log.Flags = *level

	b := NewBuild()
	client(b)

	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-interrupts
		// stop watches and clean up.
		fmt.Println() //Next line
		b.Warnf("Captured %v, stopping build and exiting...", sig)
		b.Warn("Press ctrl+c again to force exit.")
		ret := 0
		select {
		case err := <-b.Cancel():
			if err != nil {
				b.Error(err)
				b.Error("Cleaning up anyways.")
				b.Cleanup()
				ret = 1
			}
		case <-interrupts:
			fmt.Println() //Next line
			b.Warn("Force exit.")
			ret = 1
		}
		os.Exit(ret)

	}()

	flag.Parse()
	tasks := flag.Args()

	if *help {
		if len(tasks) == 0 {
			HelpTemplate.ExecuteTemplate(os.Stdout, "build", b)
			return
		}

		for _, t := range tasks {
			if t, ok := b.Tasks[t]; ok {
				HelpTemplate.ExecuteTemplate(os.Stdout, "task", t)
				continue
			}
			b.Fatalf("No Such Task: %s", t)
		}

		return
	}

	if len(tasks) == 0 {
		tasks = []string{"default"}
	}

	b.Infof("Running: %s", strings.Join(tasks, ","))
	b.Start(b.C, tasks...).Wait()
	b.Cleanup()
}

// A stage where a series of files goes for transformation, manipulation.
// There is no correlation between a stages input and output, a stage may
// decided to pass the same files after transofrmation or generate new files
// based on the input.

type Stage func(<-chan File, chan<- File)

func (stage Stage) pipe(in <-chan File) Pipe {
	out := make(chan File)
	go func() {
		stage(in, out)
		close(out)
	}()

	return out
}

//Pipe is a channel of Files.
type Pipe <-chan File

// Pipes the current Channel to the give list of Stages and returns the
// last jobs otput pipe.
func (p Pipe) Pipe(stages ...Stage) Pipe {
	switch len(stages) {
	case 0:
		return p
	case 1:
		return stages[0].pipe(p)
	default:
		return stages[0].pipe(p).Pipe(stages[1:]...)
	}
}

// Waits for the end of channel and closes all the files.
func (p Pipe) Wait() error {
	var err error
	for f := range p {
		e := f.Close()
		if err == nil && e != nil {
			err = e
		}
	}
	return err
}

//This is a combination of p.Pipe(....).Wait()
func (p Pipe) Then(stages ...Stage) error {
	return p.Pipe(stages...).Wait()
}
