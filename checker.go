package healthy // import "rmazur.io/healthy"
import (
	"context"
	"fmt"
)

// task represents any task run by the checker.
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
