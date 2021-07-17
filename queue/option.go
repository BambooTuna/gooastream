package queue

const (
	Backpressure MailboxStrategy = iota
	DropNew
	DropHead
	DropBuffer
	Fail
)

var globalBuffer = 100
var globalStrategy = Backpressure

func SetGlobalBuffer(buffer int) {
	globalBuffer = buffer
}
func SetGlobalStrategy(strategy MailboxStrategy) {
	globalStrategy = strategy
}

type queueOption struct {
	size     int
	strategy MailboxStrategy
}

func defaultQueueOption() queueOption {
	return queueOption{
		size:     globalBuffer,
		strategy: globalStrategy,
	}
}

type Option func(*queueOption)

func WithSize(size int) Option {
	return func(option *queueOption) {
		option.size = size
	}
}

func WithStrategy(strategy MailboxStrategy) Option {
	return func(option *queueOption) {
		option.strategy = strategy
	}
}
