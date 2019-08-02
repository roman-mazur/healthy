package healthy

import (
	"context"
	"fmt"
	"testing"
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
}
