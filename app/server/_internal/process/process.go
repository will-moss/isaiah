package process

import "github.com/docker/docker/client"

// Represent a tri-channel holder for a long task to communicate
type LongTaskMonitor struct {
	Results chan string
	Errors  chan error
	Done    chan bool
}

// Represent a long-running function on a Docker resource
type LongTask struct {
	Function func(*client.Client, LongTaskMonitor, map[string]interface{})
	Args     map[string]interface{}
	OnStep   func(string)
	OnError  func(error)
	OnDone   func()
}

// Run task.Function in a goroutine, and update the Function monitor provided
// as the Function is executed
func (task LongTask) RunSync(docker *client.Client) {
	finished, results, errors, done := false, make(chan string), make(chan error), make(chan bool)
	go task.Function(docker, LongTaskMonitor{Results: results, Errors: errors, Done: done}, task.Args)

	for {
		if finished {
			break
		}

		select {
		case r := <-results:
			task.OnStep(r)
		case e := <-errors:
			task.OnError(e)
		case <-done:
			finished = true
		}
	}

	task.OnDone()
}
