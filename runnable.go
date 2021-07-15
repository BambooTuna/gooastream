package gooastream

import (
	"context"
)

type (
	Left  interface{}
	Right interface{}
	Done  func()

	Cancel   func()
	Runnable interface {
		Run(ctx context.Context) (Done, Cancel)
	}

	runnableImpl struct {
		graph *graph
	}
)

func (a *runnableImpl) Run(ctx context.Context) (Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	a.graph.Run(ctx, cancel)
	return func() { <-ctx.Done() }, Cancel(cancel)
}
