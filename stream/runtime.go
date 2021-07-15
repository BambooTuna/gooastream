package stream

import "github.com/BambooTuna/gooastream/queue"

type (
	Inlet interface {
		In() queue.Queue
	}
	Outlet interface {
		Out() queue.Queue
	}
)
