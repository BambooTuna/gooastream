package gooastream

import "context"

type (
	Graph interface {
		getGraphs() graphs
	}
	graph struct {
		id   string
		from OutQueue
		to   InQueue
		task func(interface{}) interface{}
	}
	graphs []*graph
)

func (a graphs) run(ctx context.Context, cancel func()) {
	for _, v := range a {
		go v.run(ctx, cancel)
	}
}
func (a *graph) run(ctx context.Context, cancel func()) {
	defer cancel()
T:
	for {
		select {
		case <-ctx.Done():
			break T
		default:
			v, err := a.from.Pop(ctx)
			if err != nil {
				break T
			}
			a.to.Push(ctx, a.task(v))
		}
	}
}
