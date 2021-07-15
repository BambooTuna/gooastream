package gooastream

type (
	SourceRef interface {
		InQueue
		CloserQueue
	}
	Source interface {
		Via(Flow) Source
		To(Sink) Runnable

		Outlet
		Graph
	}
	sourceImpl struct {
		out   Queue
		graph *graph
		left  interface{}
	}
)

var _ Source = (*sourceImpl)(nil)

func NewChannelSource(buffer int) (SourceRef, Source) {
	in := NewQueueEmpty(buffer)
	out := NewQueueEmpty(buffer)
	return in, &sourceImpl{
		out:   out,
		graph: passThrowGraph(in, out),
	}
}

func NewSource(list []interface{}, buffer int) Source {
	in := NewQueueSlice(list)
	out := NewQueueEmpty(buffer)
	return &sourceImpl{
		out:   out,
		graph: passThrowGraph(in, out),
	}
}
func (a *sourceImpl) Via(flow Flow) Source {
	return &sourceImpl{
		out: flow.Out(),
		graph: a.graph.
			Append(passThrowGraph(a.Out(), flow.In())).
			Append(flow.getGraph()),
	}
}
func (a *sourceImpl) To(sink Sink) Runnable {
	return &runnableImpl{
		graph: a.graph.
			Append(passThrowGraph(a.Out(), sink.In())).
			Append(sink.getGraph()),
	}
}
func (a *sourceImpl) Out() Queue {
	return a.out
}
func (a *sourceImpl) getGraph() *graph {
	return a.graph
}
