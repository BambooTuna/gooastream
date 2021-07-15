package gooastream

type (
	Sink interface {
		Dummy()
		Inlet

		Graph
	}

	sinkImpl struct {
		id string
		in Queue

		graph *graph
	}
)

var _ Sink = (*sinkImpl)(nil)

func NewSink(task func(interface{}) error, buffer int) Sink {
	in := NewQueueEmpty(buffer)
	out := NewQueueSink(task)
	return &sinkImpl{
		id:    "sink",
		in:    in,
		graph: passThrowGraph(in, out),
	}
}
func (a *sinkImpl) Dummy() {
}
func (a *sinkImpl) In() Queue {
	return a.in
}
func (a *sinkImpl) getGraph() *graph {
	return a.graph
}
