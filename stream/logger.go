package stream

import "fmt"

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

type simpleLoggerImpl struct{}

func NewSimpleLoggerImpl() Logger {
	return &simpleLoggerImpl{}
}

func (e simpleLoggerImpl) Tracef(format string, args ...interface{}) {
	fmt.Printf("[Trace] "+format, args...)
}

func (e simpleLoggerImpl) Debugf(format string, args ...interface{}) {
	fmt.Printf("[Debug] "+format, args...)
}

func (e simpleLoggerImpl) Printf(format string, args ...interface{}) {
	fmt.Printf("[Print] "+format, args...)
}

func (e simpleLoggerImpl) Infof(format string, args ...interface{}) {
	fmt.Printf("[Info] "+format, args...)
}

func (e simpleLoggerImpl) Warnf(format string, args ...interface{}) {
	fmt.Printf("[Warn] "+format, args...)
}

func (e simpleLoggerImpl) Errorf(format string, args ...interface{}) {
	fmt.Printf("[Error] "+format, args...)
}

var _ Logger = (*simpleLoggerImpl)(nil)
