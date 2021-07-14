package gooastream

type (
	Sink interface {
		Dummy()

		SingleInputToClose
		Graph
	}

	sinkImpl struct {
		id     string
		in     Queue
		graphs graphs

		right interface{}
	}
)

func NewSink(task func(interface{}), buffer int) Sink {
	in := NewQueueEmpty(buffer)
	out := NewQueueSink(task)
	return &sinkImpl{
		id: "sink",
		in: in,
		graphs: graphs{
			&graph{
				id:   "sink",
				from: in,
				to:   out,
				task: emptyTaskFunc,
			},
		},
	}
}
func (a *sinkImpl) Dummy() {
}
func (a *sinkImpl) getIn() Queue {
	return a.in
}
func (a *sinkImpl) getGraphs() graphs {
	return a.graphs
}
func (a *sinkImpl) getRight() interface{} {
	return a.right
}
