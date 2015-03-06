package slurp

import (
	"errors"
	"sync"
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
	// The name of the program. Defaults to os.Args[0]
	Name string
	// Description of the program.
	Usage string
	// Version of the program
	Version string
	// Author
	Author string
	// Author e-mail
	Email string

	Tasks taskstack

	shortnames map[string]string

	cleanups    []func()
	runcleanups bool

	done chan struct{}
	lock sync.Mutex
}

func NewBuild() *Build {
	done := make(chan struct{})
	return &Build{C: &C{Log: log.New(), done: done}, Tasks: make(taskstack), done: done, lock: sync.Mutex{}}
}

// Register Tasks.
// When running the task, the dependencies will be run in parallel.
// Circular Dependencies are not allowed and will result into error.
func (b *Build) Task(tasks ...Task) {

	b.lock.Lock()
	defer b.lock.Unlock()
	for i, T := range tasks {
		if T.Name == "" {
			b.Error("Task %d Missing Name.", i)
		}

		if T.Action == nil {
			b.Fatalf("Task %s Missing Action.", T.Name)
		}

		if T.Usage == "" {
			b.Fatalf("Task %s Missing Usage.", T.Name)
		}
		
		if _, ok := b.Tasks[T.Name]; ok {
			b.Fatalf("Duplicate task: %s", T.Name)
		}
		t := &task{Task: T, deps: make(taskstack), done: b.done, running: false}

		for _, dep := range t.Deps {
			d, ok := b.Tasks[dep]
			if !ok {
				b.Fatalf("Missing Task %s. Required by Task %s.", dep, t.Name)
			}
			_, ok = d.deps[t.Name]
			if ok {
				b.Fatalf("Circular dependency %s requies %s and around.", d.Name, t.Name)
			}
			t.deps[dep] = d
		}

		b.Tasks[t.Name] = t
	}
}

func (b *Build) Start(c *C, tasks ...string) Waiter {
	var wg sync.WaitGroup
	for _, name := range tasks {
		task, ok := b.Tasks[name]
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

// Run Starts a task and waits for it to finish.
func (b *Build) Run(c *C, tasks ...string) {
	b.Start(c, tasks...).Wait()
}

// Register a function to be called when Slurp exists.
func (b *Build) Defer(fn func()) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.cleanups = append(b.cleanups, fn)
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
		for _, t := range b.Tasks {
			if t.running {
				running[t.Name] = struct{}{}
			}
		}

		count := len(running)
		for count > 0 {
			time.Sleep(time.Second)
			for name, _ := range running {
				if b.Tasks[name].running {
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
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.runcleanups {
		return
	}
	b.runcleanups = true
	for _, cleanup := range b.cleanups {
		cleanup()
	}
}
