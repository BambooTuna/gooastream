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
		Run(ctx context.Context) (Left, Right, Done, Cancel)
		LeftOutput
		RightOutput
	}

	runnableImpl struct {
		graphs      graphs
		left, right interface{}
	}
)

func (a *runnableImpl) Run(ctx context.Context) (Left, Right, Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	a.graphs.run(ctx, cancel)
	return a.getLeft(), a.getRight(), func() { <-ctx.Done() }, Cancel(cancel)
}

func (a *runnableImpl) getLeft() interface{} {
	return a.left
}
func (a *runnableImpl) getRight() interface{} {
	return a.right
}
