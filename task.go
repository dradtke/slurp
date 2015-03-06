package slurp

import (
	"fmt"
	"strings"
	"sync"
)

// The type of function to call when a task is invoked.
type Action func(*C) error

type Task struct {
	// Name of the task
	Name string
	// A short description of the task.
	Usage string
	// A long explanation of how the task works/
	Description string
	// List of dependencies.
	Deps []string
	// The function to call when the task is invoked.
	Action Action
}

type task struct {
	Task

	deps  taskstack

	lock sync.Mutex

	done    <-chan struct{}
	running bool
}

type taskstack map[string]*task

func (t *task) run(c *C) error {

	t.lock.Lock()
	c.Notice(t.Name)
	defer func() {
		t.running = false
		t.lock.Unlock()
	}()

	if t.Name != "default" {
		c = c.New(fmt.Sprintf("%s: ", t.Name))
		c.Notice("Starting.")
	}

	failed := make(chan string)
	cancel := make(chan struct{}, len(t.deps))
	done := make(chan struct{})
	var wg sync.WaitGroup
	go func(failed chan string) {
		defer close(failed)
		for _, t := range t.deps {
			select {
			case <-cancel:
				break
			default:
				wg.Add(1)
				go func(t *task) {
					defer wg.Done()
					c.Infof("Waiting for %s", t.Name)
					err := t.run(c)
					if err != nil {
						c.Error(err)
						failed <- t.Name
					}
				}(t)
			}
		}
		wg.Wait()
		close(done)
	}(failed)

	var failedjobs []string

	select {
	case <-t.done:
		cancel <- struct{}{}
		c.Warn("Task Canacled. Reasons: Canacled build.")
		return nil
	case fail, ok := <-failed:
		if ok {
			cancel <- struct{}{}
			failedjobs = append(failedjobs, fail)
			//Collect all the errors.
			for fail = range failed {
				failedjobs = append(failedjobs, fail)
			}
			return fmt.Errorf("Task Canacled. Reason: Failed Dependency (%s).", strings.Join(failedjobs, ","))
		}
	case <-done:
	}

	t.running = true
	err := t.Action(c)
	if err == nil {
		c.Notice("Done.")
	}
	return err
}
