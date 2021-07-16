package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
)

type (
	GraphNode interface {
		GraphTree() *GraphTree
	}

	Source interface {
		Via(Flow) Source
		To(Sink) Runnable

		Out() queue.Queue
		GraphNode
	}

	Flow interface {
		Via(Flow) Flow
		To(Sink) Sink

		In() queue.Queue
		Out() queue.Queue
		GraphNode
	}

	Sink interface {
		Dummy()
		In() queue.Queue

		GraphNode
	}

	Done   func()
	Cancel func()

	Runnable interface {
		Run(ctx context.Context) (Done, Cancel)
	}
	runnable struct {
		graphTree *GraphTree
	}
)

func NewRunnable(graphTree *GraphTree) Runnable {
	return &runnable{
		graphTree: graphTree,
	}
}
func (a *runnable) Run(ctx context.Context) (Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	a.graphTree.Run(ctx, cancel)
	return func() { <-ctx.Done() }, Cancel(cancel)
}
