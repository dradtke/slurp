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
	tasks    taskstack
	cleanups []func()

	end chan struct{}

	lock sync.Mutex
}

func NewBuild() *Build {
	end := make(chan struct{})
	return &Build{C: &C{log.New(), end}, tasks: make(taskstack), end: end, lock: sync.Mutex{}}
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

func waitForTasks(c *C, tasks taskstack, done chan struct{}) {

	running := make(map[string]struct{})
	for _, t := range tasks {
		if t.running {
			running[t.name] = struct{}{}
		}
	}

	count := len(running)
	for count > 0 {
		for name, _ := range running {
			if tasks[name].running {
				c.Boldf("Waiting for %s to finish.", name)
			} else {
				delete(running, name)
			}
		}
		count = len(running)
		if count > 0 {
			time.Sleep(time.Second)
		}
	}
	done <- struct{}{}
}

// Stop a build, it will call all the cleanup functions.
// Returns an error on timeout.
func (b *Build) Stop() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.end == nil {
		return errors.New("Already Cancelled.")
	}

	var err error
	close(b.end)
	done := make(chan struct{})
	go waitForTasks(b.C, b.tasks, done)
	select {
	case <-done:
		break
	case <-time.Tick(time.Second * 30):
		b.Error("Timed out.")
		err = errors.New("Timeout.")
		for _, t := range b.tasks {
			if t.running && t.name != "default" {
				b.Errorf("Task '%s' didn't honour cancellation.", t.name)
			}
		}
		b.Error("Cleaning up anyways.")
	}

	for _, cleanup := range b.cleanups {
		cleanup()
	}
	return err
}
