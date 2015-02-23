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

	//called bool

	lock sync.Mutex
}

type taskstack map[string]*task

type taskerror struct {
	name string
	err  error
}

func (t *task) run(c *C) error {

	t.lock.Lock()
	defer t.lock.Unlock()

	//if t.called {
	//		return nil
	//	}
	c.Info("Starting.")

	errs := make(chan taskerror)
	cancel := make(chan struct{}, len(t.deps))
	var wg sync.WaitGroup
	go func(errs chan taskerror) {
		defer close(errs)
		for name, t := range t.deps {
			select {
			case <-cancel:
				break
			default:

				wg.Add(1)
				go func(t *task, name string) {
					defer wg.Done()
					c.Infof("Waiting for %s", name)
					c := &C{c.New(fmt.Sprintf("%s: ", name))}
					err := t.run(c)
					if err != nil {
						c.Error(err)
					}
					errs <- taskerror{name, err}
				}(t, name)
			}
		}
		wg.Wait()
	}(errs)

	var failedjobs []string

	for err := range errs {
		if err.err != nil {
			cancel <- struct{}{}
			failedjobs = append(failedjobs, err.name)
		}
	}

	if failedjobs != nil {
		return fmt.Errorf("Task Canacled. Reason: Failed Dependency (%s).", strings.Join(failedjobs, ","))
	}

	//t.called = true
	err := t.task(c)
	if err == nil {
		c.Info("Done.")
	}

	return err
}
