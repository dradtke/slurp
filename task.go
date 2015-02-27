package slurp

import (
	"fmt"
	"strings"
	"sync"
)

type Task func(*C) error

type task struct {
	name string
	deps taskstack
	task Task

	lock sync.Mutex

	end     chan struct{}
	running bool
}
type taskstack map[string]*task

func (t *task) run(c *C) error {

	t.lock.Lock()
	defer func() {
		t.running = false
		t.lock.Unlock()
	}()

	if t.name != "default" {
		c = c.New(fmt.Sprintf("%s: ", t.name))
		c.Bold("Starting.")
	}

	failed := make(chan string)
	cancel := make(chan struct{}, len(t.deps))
	done := make(chan struct{})
	var wg sync.WaitGroup
	go func(failed chan string) {
		defer close(failed)
		for name, t := range t.deps {
			select {
			case <-cancel:
				break
			default:
				wg.Add(1)
				go func(t *task, name string) {
					defer wg.Done()
					c.Infof("Waiting for %s", name)
					err := t.run(c)
					if err != nil {
						c.Error(err)
						failed <- name
					}
				}(t, name)
			}
		}
		wg.Wait()
		close(done)
	}(failed)

	var failedjobs []string

	select {
	case <-t.end:
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
	err := t.task(c)
	if err == nil {
		c.Bold("Done.")
	}
	return err
}
