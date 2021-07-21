package stream

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"time"
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
	Done            func()
	DoneWithTimeout func(timeout time.Duration)
	Cancel          func()

	Runnable interface {
		Run(ctx context.Context) (Done, Cancel)
		RunWithTimeout(ctx context.Context) (DoneWithTimeout, Cancel)
		Merge(Runnable) Runnable
		getGraphTree() *GraphTree
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
	Returns a function that waits for the end of the context and cancel stream it.
*/
func (a *runnable) Run(ctx context.Context) (Done, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	go a.graphTree.Run(ctx, cancel)
	return func() { <-ctx.Done() }, Cancel(cancel)
}

/*
	RunWithTimeout
	Run Runnable Stream with context.Context.
	Returns a function that waits for the end of the all stream and cancel stream it.
*/
func (a *runnable) RunWithTimeout(ctx context.Context) (DoneWithTimeout, Cancel) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan interface{})
	go func() {
		a.graphTree.Run(ctx, cancel)
		close(done)
	}()
	return func(timeout time.Duration) {
		<-ctx.Done()
		select {
		case <-done:
		case <-time.After(timeout):
		}
	}, Cancel(cancel)
}

/*
	Merge
	Merge two Runnable to one.
*/
func (a *runnable) Merge(r Runnable) Runnable {
	return &runnable{graphTree: a.getGraphTree().Append(r.getGraphTree())}
}

func (a *runnable) getGraphTree() *GraphTree {
	return a.graphTree
}
