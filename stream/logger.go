package stream

var logger Logger

func init() {
	logger = &emptyLoggerImpl{}
}

func Log() Logger {
	return logger
}

func SetLog(log Logger) {
	logger = log
}

type Logger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type emptyLoggerImpl struct{}

func (e emptyLoggerImpl) Tracef(format string, args ...interface{}) {
}

func (e emptyLoggerImpl) Debugf(format string, args ...interface{}) {
}

func (e emptyLoggerImpl) Printf(format string, args ...interface{}) {
}

func (e emptyLoggerImpl) Infof(format string, args ...interface{}) {
}

func (e emptyLoggerImpl) Warnf(format string, args ...interface{}) {
}

func (e emptyLoggerImpl) Errorf(format string, args ...interface{}) {
}

var _ Logger = (*emptyLoggerImpl)(nil)
