package apphostgae

import (
	"context"
	"github.com/strongo/app/delaying"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/taskqueue"
)

func MustRegisterDelayedFunc(key string, i interface{}) delaying.Function {
	f := delay.Func(key, i)
	return delaying.NewFunction(key, f,
		func(c context.Context, params delaying.Params, args ...interface{}) error {
			task, err := f.Task(args...)
			if err != nil {
				return err
			}
			if d := params.Delay(); d > 0 {
				task.Delay = d
			}
			task, err = taskqueue.Add(c, task, params.Queue())
			return err
		},
		func(c context.Context, params delaying.Params, args ...[]interface{}) (err error) {
			tasks := make([]*taskqueue.Task, 0, len(args))

			for i, arg := range args {
				if tasks[i], err = f.Task(arg...); err != nil {
					return err
				}
				if d := params.Delay(); d > 0 {
					tasks[i].Delay = d
				}
			}
			tasks, err = taskqueue.AddMulti(c, tasks, params.Queue())
			return err
		},
	)
}
