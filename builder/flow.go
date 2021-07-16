package builder

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	balanceImpl struct {
		in        queue.Queue
		out       []queue.Queue
		graphTree *stream.GraphTree
	}
	mergeImpl struct {
		in        []queue.Queue
		out       queue.Queue
		graphTree *stream.GraphTree
	}
)

var _ Balance = (*balanceImpl)(nil)
var _ Merge = (*mergeImpl)(nil)

func NewBalance(size int) Balance {
	in := queue.NewQueueEmpty(0)
	outs := make([]queue.Queue, size)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		out := queue.NewQueueEmpty(0)
		outs[i] = out
		graphTree.Add(stream.PassThrowGraph(in, out))
	}
	return &balanceImpl{
		in:        in,
		out:       outs,
		graphTree: graphTree,
	}
}

func (a *balanceImpl) In() queue.Queue {
	return a.in
}
func (a *balanceImpl) Out() []queue.Queue {
	return a.out
}
func (a *balanceImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}

func NewMerge(size int) Merge {
	ins := make([]queue.Queue, size)
	out := queue.NewQueueEmpty(0)
	graphTree := stream.EmptyGraph()
	for i := 0; i < size; i++ {
		in := queue.NewQueueEmpty(0)
		ins[i] = in
		graphTree.Add(stream.PassThrowGraph(in, out))
	}
	return &mergeImpl{
		in:        ins,
		out:       out,
		graphTree: graphTree,
	}
}

func (a *mergeImpl) In() []queue.Queue {
	return a.in
}
func (a *mergeImpl) Out() queue.Queue {
	return a.out
}
func (a *mergeImpl) GraphTree() *stream.GraphTree {
	return a.graphTree
}
