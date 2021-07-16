package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
)

// GraphNode
// An interface that represents having a GraphTree.
type GraphNode interface {
	GraphTree() *GraphTree
}

// Source
// An interface that represents a stream with a single output.
type Source interface {
	Via(Flow) Source
	To(Sink) Runnable

	Out() queue.Queue
	GraphNode
}

// Flow
// An interface that represents a stream with one input and one output.
type Flow interface {
	Via(Flow) Flow
	To(Sink) Sink

	In() queue.Queue
	Out() queue.Queue
	GraphNode
}

// Sink
// An interface that represents a stream with a single input.
type Sink interface {
	Dummy()
	In() queue.Queue

	GraphNode
}

type (
	Done   func()
	Cancel func()

	Runnable interface {
		Run(ctx context.Context) (Done, Cancel)
	}
	runnable struct {
		graphTree *GraphTree
	}
)

var _ Runnable = (*runnable)(nil)

/*
	NewRunnable
	Create a Runnable from a GraphTree.
*/
func NewRunnable(graphTree *GraphTree) Runnable {
	return &runnable{
		graphTree: graphTree,
	}
}

/*
	Run
	Run Runnable Stream with context.Context.
	Returns a function that waits for the end of the stream and cancel stream it.
*/
func (a *runnable) Run(ctx context.Context) (Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	a.graphTree.Run(ctx, cancel)
	return func() { <-ctx.Done() }, Cancel(cancel)
}
