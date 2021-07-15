package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/builder"
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
		graphTree builder.GraphTree
	}
)

func (a *runnableImpl) Run(ctx context.Context) (Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	a.graphTree.Run(ctx, cancel)
	return func() { <-ctx.Done() }, Cancel(cancel)
}
