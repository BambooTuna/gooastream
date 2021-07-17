package std

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"os"
)

type SinkConfig struct {
	FilePath string
	Buffer   int
}

func NewFileSink(conf *SinkConfig) stream.Sink {
	in := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newFileSinkWire(in, conf.FilePath))
	return stream.BuildSink(in, graphTree)
}

type fileSinkWire struct {
	from     queue.OutQueue
	filePath string
}

func newFileSinkWire(from queue.OutQueue, filePath string) stream.Wire {
	return &fileSinkWire{
		from:     from,
		filePath: filePath,
	}
}
func (a fileSinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	writer, err := os.OpenFile(a.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer func() {
		cancel()
		a.from.Close()
		_ = writer.Close()
	}()
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
			data, ok := v.([]byte)
			if !ok {
				continue
			}
			_, err = writer.Write(data)
			if err != nil {
				break T
			}
		}
	}
}

var _ stream.Wire = (*fileSinkWire)(nil)
