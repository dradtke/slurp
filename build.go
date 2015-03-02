package slurp

import (
	"errors"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/omeid/slurp/log"
)

// Waiter interface implementsion Wait() function.
type Waiter interface {
	Wait()
}

// Build is a simple build harness that you can register tasks and their
// dependencies and then run them. You usually don't need to create your
// own Build and instead use the one passed by Slurp runner.
type Build struct {
	*C
	tasks    taskstack
	cleanups []func()

	done chan struct{}

	lock sync.Mutex
}

func NewBuild() *Build {
	done := make(chan struct{})
	return &Build{C: &C{log.New(), done}, tasks: make(taskstack), done: done, lock: sync.Mutex{}}
}

// Register a task and it's dependencies.
// When running the task, the dependencies will be run in parallel.
// Circular Dependencies are not allowed and will result into error.
func (b *Build) Task(name string, deps []string, Task Task) {

	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.tasks[name]; ok {
		b.Fatalf("Duplicate task: %s", name)
	}

	Deps := make(taskstack)
	t := task{name: name, deps: Deps, task: Task, running: false}

	for _, dep := range deps {
		d, ok := b.tasks[dep]
		if !ok {
			b.Fatalf("Missing Task %s. Required by Task %s.", dep, name)
		}
		_, ok = d.deps[name]
		if ok {
			b.Fatalf("Circular dependency %s requies %s and around.", d.name, name)
		}

		t.deps[dep] = d
	}

	b.tasks[name] = &t
}

// Run Starts a task and waits for it to finish.
func (b *Build) Run(c *C, tasks ...string) {
	b.Start(c, tasks...).Wait()
}

// Start a task but doesn't wait for it to finish.
func (b *Build) Start(c *C, tasks ...string) Waiter {
	var wg sync.WaitGroup
	for _, name := range tasks {
		task, ok := b.tasks[name]
		if !ok {
			b.Fatalf("No Such Task: %s", name)
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := task.run(c)
			if err != nil {
				b.Error(err)
			}
		}()
	}

	return &wg
}

// Register a function to be called when Slurp exists.
func (b *Build) Defer(fn func()) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.cleanups = append(b.cleanups, fn)
}

// Depracted use slurp.C.Done
func (b Build) Wait() {
	b.Warn("slrup.Build.Wait is Depracted. Use slurp.C.Wait instead.")
	<-make(chan struct{})
}

// Stop a build, it will call all the cleanup functions.
// Returns an error on timeout.
func (b *Build) Cancel() <-chan error {

	errs := make(chan error)
	go func() {
		b.lock.Lock()
		defer b.lock.Unlock()
		close(b.done)

		if b.done == nil {
			errs <- errors.New("Already Cancelled.")
		}

		running := make(map[string]struct{})
		for _, t := range b.tasks {
			if t.running {
				running[t.name] = struct{}{}
			}
		}

		count := len(running)
		for count > 0 {
			time.Sleep(time.Second)
			for name, _ := range running {
				if b.tasks[name].running {
					b.Noticef("Waiting for %s to finish.", name)
				} else {
					delete(running, name)
					count--
				}
			}
		}
	}()
	return errs
}

func (b *Build) Cleanup() {
	<-b.done
	for _, cleanup := range b.cleanups {
		cleanup()
	}
}

// Nothing to see here, move on.
func (b *Build) End() {
	b.lock.Lock()
	defer b.lock.Unlock()
	close(b.done)
}

var help = template.Must(template.New("help").Parse(`
USAGE:
slurp [flags] [tasks]

TASKS:{{ range .tasks }}
{{ .name }} {{ index .help 0 }}
{{ end }}
`))

var taskhelp = template.Must(template.New("taskhelp").Parse(`
HELP: {{ .name }}
{{ range .help }} {{ . }}
{{ end }}
`))

func (b *Build) PrintHelp(task string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if task == "" {
		help.Execute(os.Stdout, b)
		return
	}
	t, ok := b.tasks[task]
	if ok {
		taskhelp.Execute(os.Stdout, t)
		return
	}
	b.Errorf("No Such Task: %s", task)
}
