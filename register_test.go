package apphostgae

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/strongo/delaying"
	"google.golang.org/appengine/v2/taskqueue"
)

func TestMustRegisterDelayedFunc(t *testing.T) {
	ctx := context.Background()
	delayer := MustRegisterDelayedFunc("testRegisterDelayedFunc", func(_ context.Context, _ int) error {
		return nil
	})
	assert.NotNil(t, delayer)

	t.Run("EnqueueWork_task_err", func(t *testing.T) {
		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWork(ctx, params, "invalid_type_string")
		if err == nil {
			t.Fatal("expected error on invalid arg type")
		}
	})

	t.Run("EnqueueWork_with_delay", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		var gotDelay time.Duration
		taskqueueAdd = func(c context.Context, task *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			gotDelay = task.Delay
			return task, nil
		}

		params := delaying.With("my-queue", "sub", 10*time.Second)
		err := delayer.EnqueueWork(ctx, params, 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotDelay != 10*time.Second {
			t.Fatalf("expected 10s delay, got %v", gotDelay)
		}
	})

	t.Run("EnqueueWork_without_delay", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		var gotDelay time.Duration
		taskqueueAdd = func(c context.Context, task *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			gotDelay = task.Delay
			return task, nil
		}

		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWork(ctx, params, 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotDelay != 0 {
			t.Fatalf("expected 0 delay, got %v", gotDelay)
		}
	})

	t.Run("EnqueueWork_add_error", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		taskqueueAdd = func(c context.Context, task *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			return nil, errors.New("add_fail")
		}

		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWork(ctx, params, 123)
		if err == nil {
			t.Fatal("expected add error")
		}
	})

	t.Run("EnqueueWorkMulti_task_err", func(t *testing.T) {
		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWorkMulti(ctx, params, []any{"invalid_type_string"})
		if err == nil {
			t.Fatal("expected error on invalid arg type")
		}
	})

	t.Run("EnqueueWorkMulti_with_delay", func(t *testing.T) {
		origAddMulti := taskqueueAddMulti
		defer func() { taskqueueAddMulti = origAddMulti }()
		var gotDelays []time.Duration
		taskqueueAddMulti = func(c context.Context, tasks []*taskqueue.Task, queueName string) ([]*taskqueue.Task, error) {
			for _, task := range tasks {
				gotDelays = append(gotDelays, task.Delay)
			}
			return tasks, nil
		}

		params := delaying.With("my-queue", "sub", 5*time.Second)
		err := delayer.EnqueueWorkMulti(ctx, params, []any{1}, []any{2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(gotDelays) != 2 || gotDelays[0] != 5*time.Second || gotDelays[1] != 5*time.Second {
			t.Fatalf("expected 5s delays, got %v", gotDelays)
		}
	})

	t.Run("EnqueueWorkMulti_without_delay", func(t *testing.T) {
		origAddMulti := taskqueueAddMulti
		defer func() { taskqueueAddMulti = origAddMulti }()
		var gotDelays []time.Duration
		taskqueueAddMulti = func(c context.Context, tasks []*taskqueue.Task, queueName string) ([]*taskqueue.Task, error) {
			for _, task := range tasks {
				gotDelays = append(gotDelays, task.Delay)
			}
			return tasks, nil
		}

		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWorkMulti(ctx, params, []any{1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(gotDelays) != 1 || gotDelays[0] != 0 {
			t.Fatalf("expected 0 delay, got %v", gotDelays)
		}
	})

	t.Run("EnqueueWorkMulti_add_error", func(t *testing.T) {
		origAddMulti := taskqueueAddMulti
		defer func() { taskqueueAddMulti = origAddMulti }()
		taskqueueAddMulti = func(c context.Context, tasks []*taskqueue.Task, queueName string) ([]*taskqueue.Task, error) {
			return nil, errors.New("add_multi_fail")
		}

		params := delaying.With("my-queue", "sub", 0)
		err := delayer.EnqueueWorkMulti(ctx, params, []any{1})
		if err == nil {
			t.Fatal("expected add multi error")
		}
	})
}
