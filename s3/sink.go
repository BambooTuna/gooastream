package s3

import (
	"context"
	"github.com/BambooTuna/gooastream/queue"
	"github.com/BambooTuna/gooastream/stream"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type SinkConfig struct {
	UploadInput *s3manager.UploadInput
}

func NewS3Sink(conf *SinkConfig, conn *s3manager.Uploader, options ...queue.Option) stream.Sink {
	in := queue.NewQueueEmpty(options...)
	graphTree := stream.EmptyGraph()
	graphTree.AddWire(newS3SinkWire(in, conn, conf.UploadInput))
	return stream.BuildSink(in, graphTree)
}

type s3SinkWire struct {
	from   queue.OutQueue
	reader *io.PipeReader
	writer *io.PipeWriter

	conn        *s3manager.Uploader
	uploadInput *s3manager.UploadInput
}

func newS3SinkWire(from queue.OutQueue, conn *s3manager.Uploader, uploadInput *s3manager.UploadInput) stream.Wire {
	reader, writer := io.Pipe()
	return &s3SinkWire{
		from:   from,
		reader: reader,
		writer: writer,

		conn:        conn,
		uploadInput: uploadInput,
	}
}
func (a s3SinkWire) Run(ctx context.Context, cancel context.CancelFunc) {
	var err error
	uploadCtx := context.Background()
	defer func() {
		cancel()
		a.from.Close()
		_ = a.writer.Close()
		if err != nil {
			stream.Log().Errorf("%v", err)
		}
	}()
	go func() {
		input := a.uploadInput
		input.Body = a.reader
		_, _ = a.conn.UploadWithContext(uploadCtx, input)
		cancel()
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
			_, err = a.writer.Write(data)
			if err != nil {
				break T
			}
		}
	}
}

var _ stream.Wire = (*s3SinkWire)(nil)
