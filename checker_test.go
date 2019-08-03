package healthy

import (
	"context"
	"fmt"
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
