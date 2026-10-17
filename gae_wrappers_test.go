package apphostgae

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"google.golang.org/appengine/v2/delay"
	"google.golang.org/appengine/v2/taskqueue"
)

var testDelayFunc = delay.MustRegister("testGaeWrapperDelayFunc", func(_ context.Context, _ int) error {
	return nil
})

func TestCreateDelayTask(t *testing.T) {
	t.Run("empty_queue_name", func(t *testing.T) {
		_, err := CreateDelayTask("", "sub", testDelayFunc, 1)
		if err == nil || !strings.Contains(err.Error(), "queueName is empty") {
			t.Fatalf("expected queueName is empty, got %v", err)
		}
	})

	t.Run("default_queue_name", func(t *testing.T) {
		_, err := CreateDelayTask("default", "sub", testDelayFunc, 1)
		if err == nil || !strings.Contains(err.Error(), "queueName is 'default'") {
			t.Fatalf("expected queueName is 'default', got %v", err)
		}
	})

	t.Run("empty_sub_path", func(t *testing.T) {
		_, err := CreateDelayTask("my-queue", "", testDelayFunc, 1)
		if err == nil || !strings.Contains(err.Error(), "subPath is empty") {
			t.Fatalf("expected subPath is empty, got %v", err)
		}
	})

	t.Run("task_args_error", func(t *testing.T) {
		// testDelayFunc expects an int, passing string will error in f.Task()
		_, err := CreateDelayTask("my-queue", "sub", testDelayFunc, "invalid-arg")
		if err == nil {
			t.Fatal("expected error from f.Task()")
		}
	})

	t.Run("success", func(t *testing.T) {
		task, err := CreateDelayTask("my-queue", "sub/action", testDelayFunc, 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(task.Path, "task=sub%2Faction") || !strings.Contains(task.Path, "queue=my-queue") {
			t.Fatalf("expected query params in task.Path, got %s", task.Path)
		}
	})
}

func TestAddTaskToQueue(t *testing.T) {
	ctx := context.Background()
	task := &taskqueue.Task{Path: "/test"}

	t.Run("empty_queue", func(t *testing.T) {
		_, err := AddTaskToQueue(ctx, task, "")
		if err == nil || !strings.Contains(err.Error(), "queueName is empty") {
			t.Fatalf("expected queueName is empty error, got %v", err)
		}
	})

	t.Run("default_queue", func(t *testing.T) {
		_, err := AddTaskToQueue(ctx, task, "default")
		if err == nil || !strings.Contains(err.Error(), "queueName is 'default'") {
			t.Fatalf("expected queueName is 'default' error, got %v", err)
		}
	})

	t.Run("add_error", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			return nil, errors.New("fail_add")
		}

		_, err := AddTaskToQueue(ctx, task, "my-queue")
		if err == nil || !strings.Contains(err.Error(), "failed to add task to queue") {
			t.Fatalf("expected failed to add task to queue error, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			return t, nil
		}

		res, err := AddTaskToQueue(ctx, task, "my-queue")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != task {
			t.Fatalf("expected returned task, got %v", res)
		}
	})
}

func TestCallDelayFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("create_task_error", func(t *testing.T) {
		err := CallDelayFunc(ctx, "", "sub", testDelayFunc, 1)
		if err == nil {
			t.Fatal("expected error on invalid queue")
		}
	})

	t.Run("add_task_error", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			return nil, errors.New("fail_add")
		}

		err := CallDelayFunc(ctx, "my-queue", "sub", testDelayFunc, 1)
		if err == nil {
			t.Fatal("expected error on queue add fail")
		}
	})

	t.Run("success", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		var queuedDelay time.Duration
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			queuedDelay = t.Delay
			return t, nil
		}

		err := CallDelayFunc(ctx, "my-queue", "sub", testDelayFunc, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if queuedDelay != 0 {
			t.Fatalf("expected delay 0, got %v", queuedDelay)
		}
	})
}

func TestEnqueueWork(t *testing.T) {
	ctx := context.Background()

	t.Run("create_task_error", func(t *testing.T) {
		err := EnqueueWork(ctx, "", "sub", time.Minute, testDelayFunc, 1)
		if err == nil || !strings.Contains(err.Error(), "failed to create delay task") {
			t.Fatalf("expected failed to create delay task, got %v", err)
		}
	})

	t.Run("add_error", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			return nil, errors.New("fail_add")
		}

		err := EnqueueWork(ctx, "my-queue", "sub", time.Minute, testDelayFunc, 1)
		if err == nil {
			t.Fatal("expected error on add")
		}
	})

	t.Run("success", func(t *testing.T) {
		origAdd := taskqueueAdd
		defer func() { taskqueueAdd = origAdd }()
		var queuedDelay time.Duration
		taskqueueAdd = func(c context.Context, t *taskqueue.Task, queueName string) (*taskqueue.Task, error) {
			queuedDelay = t.Delay
			return t, nil
		}

		err := EnqueueWork(ctx, "my-queue", "sub", time.Minute, testDelayFunc, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if queuedDelay != time.Minute {
			t.Fatalf("expected 1m delay, got %v", queuedDelay)
		}
	})
}

func TestEnqueueWorkMulti(t *testing.T) {
	ctx := context.Background()

	t.Run("create_task_error", func(t *testing.T) {
		err := EnqueueWorkMulti(ctx, "", "sub", time.Minute, testDelayFunc, []interface{}{1})
		if err == nil || !strings.Contains(err.Error(), "faield to create task for work # 0") {
			t.Fatalf("expected failed to create task error, got %v", err)
		}
	})

	t.Run("add_multi_error", func(t *testing.T) {
		origAddMulti := taskqueueAddMulti
		defer func() { taskqueueAddMulti = origAddMulti }()
		taskqueueAddMulti = func(c context.Context, tasks []*taskqueue.Task, queueName string) ([]*taskqueue.Task, error) {
			return nil, errors.New("fail_add_multi")
		}

		err := EnqueueWorkMulti(ctx, "my-queue", "sub", time.Minute, testDelayFunc, []interface{}{1})
		if err == nil {
			t.Fatal("expected add_multi error")
		}
	})

	t.Run("success", func(t *testing.T) {
		origAddMulti := taskqueueAddMulti
		defer func() { taskqueueAddMulti = origAddMulti }()
		var addedCount int
		var delays []time.Duration
		taskqueueAddMulti = func(c context.Context, tasks []*taskqueue.Task, queueName string) ([]*taskqueue.Task, error) {
			addedCount = len(tasks)
			for _, tk := range tasks {
				delays = append(delays, tk.Delay)
			}
			return tasks, nil
		}

		err := EnqueueWorkMulti(ctx, "my-queue", "sub", 2*time.Minute, testDelayFunc,
			[]interface{}{1},
			[]interface{}{2},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if addedCount != 2 {
			t.Fatalf("expected 2 tasks added, got %d", addedCount)
		}
		if delays[0] != 2*time.Minute || delays[1] != 2*time.Minute {
			t.Fatalf("expected 2m delay on all tasks, got %v", delays)
		}
	})
}
