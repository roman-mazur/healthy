package healthy // import "rmazur.io/healthy"
import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Task represents any task run by the checker.
type Task interface {
	// Name returns the descriptive task name.
	Name() string
	// Run executes the task returning an error if it fails.
	Run(ctx context.Context) error
}

type retryableTask struct {
	task       Task
	maxRetries int
}

// WithRetries wraps another task to retry it in the case of failure.
func WithRetries(task Task, max int) Task {
	if task == nil {
		panic(fmt.Errorf("task must not be nil"))
	}
	if max <= 0 {
		max = 1
	}
	return &retryableTask{task: task, maxRetries: max}
}

func (rt *retryableTask) Name() string {
	return rt.task.Name()
}

func (rt *retryableTask) Run(ctx context.Context) error {
	var err error
	for a := 0; a < rt.maxRetries; a++ {
		err = rt.task.Run(ctx)
		if err == nil {
			break
		}
	}
	return err
}

type scheduler func() <-chan struct{}

type execRule struct {
	task                       Task
	scheduleTicks              scheduler
	failuresCount              int
	cancellation               chan struct{}
	failureOptions 			   *FailureOptions
}

// Notifier can be set on a Checker to report failed tasks.
type Notifier interface {
	Notify(e error)
}

// FailureOptions define when to report a detected task failure.
type FailureOptions struct {
	// Number of consequent task failures required to report a failure.
	ReportFailuresCount int
}

// DefaultFailureOptions are used if Checker of Task level options (with AddXXX methods) are not set.
var DefaultFailureOptions = FailureOptions{
	ReportFailuresCount: 3,
}

// Checker runs configured tasks with a specified schedule.
type Checker struct {
	rules []execRule
	stopSync sync.WaitGroup

	Notifier Notifier
	DefaultFailureOptions *FailureOptions
}

func generateDelay(period, flex time.Duration) time.Duration {
	if flex == 0 {
		return period
	}
	return period + time.Duration(rand.Int63n(int64(flex)*2)) - flex
}

var tickMessage = struct{}{}

func createPeriodSchedule(period, flex time.Duration, cancel <-chan struct{}) <-chan struct{} {
	ticks := make(chan struct{}, 1)
	var timer *time.Timer

	go func() {
		<-cancel
		if timer != nil {
			timer.Stop()
		}
		close(ticks)
	}()

	timer = time.AfterFunc(generateDelay(period, flex), func() {
		ticks <- tickMessage
		timer.Reset(generateDelay(period, flex))
	})

	return ticks
}

// AddTaskWithPeriod adds a new task to be run with the specified period.
// Flex parameter allows adding random component to task scheduling:
// next task invocation will happen at now + period +/- flex.
func (c *Checker) AddTaskWithPeriod(task Task, period, flex time.Duration) {
	c.addTaskWithPeriodWithOptions(task, period, flex, c.DefaultFailureOptions)
}

func (c *Checker) addTaskWithPeriodWithOptions(task Task, period, flex time.Duration, options *FailureOptions) {
	fo := options
	if fo == nil {
		fo = &DefaultFailureOptions
	}
	cancel := make(chan struct{}, 1)
	c.rules = append(c.rules, execRule{
		task: task,
		scheduleTicks: func() <-chan struct{} {
			return createPeriodSchedule(period, flex, cancel)
		},
		cancellation:               cancel,
		failureOptions: 			fo,
	})
}

// Run schedules all the tasks.
// AddXXX() methods must not be called once Run() is invoked.
func (c *Checker) Run(ctx context.Context) {
	c.stopSync.Add(len(c.rules))

	for _, r := range c.rules {
		go func() {
			for range r.scheduleTicks() {
				if err := r.task.Run(ctx); err != nil {
					reportCount := r.failureOptions.ReportFailuresCount
					if r.failuresCount < reportCount {
						r.failuresCount++
						if r.failuresCount == reportCount {
							c.notify(err)
						}
					}
				} else {
					r.failuresCount = 0
				}
			}
			c.stopSync.Done()
		}()
	}
}

func (c *Checker) notify(err error) {
	n := c.Notifier
	if n != nil {
		n.Notify(err)
	}
}

// Stop cancels currently scheduled task invocations.
// No tasks will be fired after this method completes.
// Run can be called again after calling Stop.
func (c *Checker) Stop() {
	for _, r := range c.rules {
		r.cancellation <- struct{}{}
	}
	c.stopSync.Wait()
}
