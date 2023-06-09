package apphostgae

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/appengine/v2/delay"
	"google.golang.org/appengine/v2/taskqueue"
	"net/url"
	"time"
)

// CallDelayFunc - TODO: Document why whe need this
// Obsolete - use EnqueueWork instead
var CallDelayFunc = func(c context.Context, queueName, subPath string, f *delay.Function, args ...interface{}) error {
	return CallDelayFuncWithDelay(c, 0, queueName, subPath, f, args...)
}

// CallDelayFuncWithDelay - TODO: Document why whe need this
// Obsolete - use EnqueueWork instead
var CallDelayFuncWithDelay = func(c context.Context, delay time.Duration, queueName, subPath string, f *delay.Function, args ...interface{}) error {
	if task, err := CreateDelayTask(queueName, subPath, f, args...); err != nil {
		return err
	} else {
		task.Delay = delay
		_, err = AddTaskToQueue(c, task, queueName)
		return err
	}
}

const failToCreateDelayTask = "failed to create delay task"
const failToCreateDelayTaskPrefix = failToCreateDelayTask + ": "

// CreateDelayTask creates a delay task TODO: Document why whe need this
func CreateDelayTask(queueName, subPath string, f *delay.Function, args ...interface{}) (*taskqueue.Task, error) {
	if queueName == "" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "queueName is empty")
	}
	if queueName == "default" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "queueName is 'default'")
	}
	if subPath == "" {
		return nil, errors.New(failToCreateDelayTaskPrefix + "subPath is empty")
	}
	if task, err := f.Task(args...); err != nil {
		return task, fmt.Errorf("%s: queue=%v, subPath=%v: %w", failToCreateDelayTask, queueName, subPath, err)
	} else {
		task.Path += fmt.Sprintf("?task=%v&queue=%v", url.QueryEscape(subPath), url.QueryEscape(queueName))
		return task, nil
	}
}

// EnqueueWorkMulti - is obsolete
// Obsolete
func EnqueueWorkMulti(ctx context.Context, queueName, subPath string, delay time.Duration, f *delay.Function, args ...[]interface{}) (err error) {
	tasks := make([]*taskqueue.Task, len(args))
	for i, arg := range args {
		if tasks[i], err = CreateDelayTask(queueName, subPath, f, arg...); err != nil {
			return fmt.Errorf("faield to create task for work # %d: %w", i, err)
		}
		tasks[i].Delay = delay
	}
	_, err = taskqueue.AddMulti(ctx, tasks, queueName)
	return err
}

// EnqueueWork - is obsolete
// Obsolete
func EnqueueWork(ctx context.Context, queueName, subPath string, delay time.Duration, f *delay.Function, args ...interface{}) (err error) {
	var task *taskqueue.Task
	task, err = CreateDelayTask(queueName, subPath, f, args...)
	if err == nil {
		return fmt.Errorf("failed to create delay task: %w", err)
	}
	task.Delay = delay
	_, err = taskqueue.Add(ctx, task, queueName)
	return err
}

const failedToAddTaskToQueue = "failed to add task to queue"
const failedToAddTaskToQueuePrefix = failedToAddTaskToQueue + ": "

// AddTaskToQueue - adds tasks to a queue TODO: Document why whe need this
var AddTaskToQueue = func(c context.Context, t *taskqueue.Task, queueName string) (task *taskqueue.Task, err error) {
	if queueName == "" {
		return nil, errors.New(failedToAddTaskToQueuePrefix + "queueName is empty")
	}
	if queueName == "default" {
		return nil, errors.New(failedToAddTaskToQueuePrefix + "queueName is 'default'")
	}
	if task, err = taskqueue.Add(c, t, queueName); err != nil {
		err = fmt.Errorf("%s: %w", failedToAddTaskToQueue, err)
		//} else {
		//	isInTransaction := gaedb.NewDatabase().IsInTransaction(c)
		//	log.Debugf(c, "Added task to queue '%v', tx=%v): path: %v", queueName, isInTransaction, task.Path)
	}
	return
}
