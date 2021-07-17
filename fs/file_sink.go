package fs

import (
	"bufio"
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"io"
	"os"
)

type FileSourceConfig struct {
	FilePath   string
	BufferSize int
	Buffer     int
}

func NewFileSource(conf *FileSourceConfig) stream.Source {
	out := queue.NewQueueEmpty(conf.Buffer)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newFileSourceWire(out, conf.FilePath, conf.BufferSize))
	return stream.BuildSource(out, graphTree)
}

type fileSourceWire struct {
	to         queue.InQueue
	filePath   string
	bufferSize int
}

func newFileSourceWire(to queue.InQueue, filePath string, bufferSize int) stream.Wire {
	return &fileSourceWire{
		to:         to,
		filePath:   filePath,
		bufferSize: bufferSize,
	}
}
func (a fileSourceWire) Run(ctx context.Context, cancel context.CancelFunc) {
	file, err := os.Open(a.filePath)
	if err != nil {
		return
	}
	reader := bufio.NewReader(file)
	defer func() {
		cancel()
		a.to.Close()
		_ = file.Close()
	}()
	buf := make([]byte, a.bufferSize)
T:
	for {
		select {
		case <-ctx.Done():
			break T
		default:
			n, err := reader.Read(buf)
			if err != nil {
				if err == io.EOF {
					<-ctx.Done()
				}
				break T
			}
			err = a.to.Push(ctx, buf[:n])
			if err != nil {
				break T
			}
		}
	}
}

var _ stream.Wire = (*fileSourceWire)(nil)

type FileSinkConfig struct {
	FilePath string
	Buffer   int
}

func NewFileSink(conf *FileSinkConfig) stream.Sink {
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
