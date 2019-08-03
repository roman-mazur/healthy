package healthy

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type funcTask func(ctx context.Context) error

func (f funcTask) Name() string {
	return "test func"
}

func (f funcTask) Run(ctx context.Context) error {
	return f(ctx)
}

func TestWithRetries(t *testing.T) {
	counter := 0
	failure := func(ctx context.Context) error {
		counter++
		return fmt.Errorf("test")
	}

	err := WithRetries(funcTask(failure), 5).Run(context.Background())
	if err == nil || err.Error() != "test" {
		t.Errorf("An error was expected, got %s", err)
	} else if counter != 5 {
		t.Errorf("Expected 5 retries, got %d", counter)
	}

	counter = 0
	handler := func(ctx context.Context) error {
		counter++
		if counter == 2 {
			return nil
		}
		return fmt.Errorf("test")
	}
	err = WithRetries(funcTask(handler), 5).Run(context.Background())
	if err != nil {
		t.Errorf("Got an error %s", err)
	} else if counter != 2 {
		t.Errorf("Expected 2 retries, got %d", counter)
	}
}

func TestChecker_AddTaskWithPeriod(t *testing.T) {
	var handlerTicks []time.Time

	checker := &Checker{}
	checker.AddTaskWithPeriod(
		funcTask(func(ctx context.Context) error {
			handlerTicks = append(handlerTicks, time.Now())
			return nil
		}),
		200*time.Millisecond,
		100*time.Millisecond,
	)
	checker.Run(context.Background())

	testSync := make(chan int)

	time.AfterFunc(500*time.Millisecond, func() {
		if len(handlerTicks) < 1 || len(handlerTicks) > 5 {
			t.Errorf("Unexpected ticks %s", handlerTicks)
		} else {
			t.Logf("Ticks: %s", handlerTicks)
		}
		checker.Stop()
		testSync <- 1
	})

	select {
	case <-time.After(2 * time.Second):
		t.Errorf("Timeout testing checker run!")
	case <-testSync:
	}
}

type fn func(e error)
func (f fn) Notify(e error) {
	f(e)
}

func TestChecker_AddTaskWithPeriod_Notifier(t *testing.T) {
	taskResults := []error{
		fmt.Errorf("failure1"),
		fmt.Errorf("failure2"),
		fmt.Errorf("failure3"),
		nil,
		fmt.Errorf("failure4"),
		fmt.Errorf("failure5"),
		fmt.Errorf("failure6"),
		nil,
	}
	period := 50 * time.Millisecond
	checker := &Checker{}
	checker.DefaultFailureOptions = &FailureOptions{ReportFailuresCount: 2}
	resultIndex := 0
	checker.AddTaskWithPeriod(
		funcTask(func(ctx context.Context) error {
			res := taskResults[resultIndex]
			resultIndex = (resultIndex + 1) % len(taskResults)
			return res
		}),
		period,
		0,
	)

	var reportedErrors []error
	checker.Notifier = fn(func (e error) {
		reportedErrors = append(reportedErrors, e)
	})

	checker.Run(context.Background())
	time.Sleep(period * time.Duration(len(taskResults)) + period / 4)
	checker.Stop()

	if len(reportedErrors) != 2 {
		t.Errorf("Unexpected errors count. Expected 2, got %s", reportedErrors)
	} else {
		if !strings.Contains(reportedErrors[0].Error(), "failure2") {
			t.Errorf("Unexpected error 1 %s", reportedErrors[0])
		}
		if !strings.Contains(reportedErrors[1].Error(), "failure5") {
			t.Errorf("Unexpected error 1 %s", reportedErrors[1])
		}
	}
}

func ExampleChecker() {
	counter := 1

	var checker Checker
	checker.AddTaskWithPeriod(
		funcTask(func (ctx context.Context) error {
			fmt.Printf("Running check %d\n", counter)
			counter++
			return nil
		}),
		500 * time.Millisecond,
		0,
	)

	checker.Run(context.Background())
	time.Sleep(1100 * time.Millisecond)
	checker.Stop()

	// Output:
	// Running check 1
	// Running check 2
}
