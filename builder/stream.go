package builder

import (
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
)

type (
	Balance interface {
		In() queue.Queue
		Out() []queue.Queue
		stream.GraphNode
	}

	Merge interface {
		In() []queue.Queue
		Out() queue.Queue
		stream.GraphNode
	}
)
