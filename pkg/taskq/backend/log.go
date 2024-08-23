package backend

import (
	"sync/atomic"
	"unsafe"

	"github.com/go-logr/logr"
)

var globalLogger unsafe.Pointer

func SetLogger(l logr.Logger) {
	atomic.StorePointer(&globalLogger, unsafe.Pointer(&l))
}

func getLogger() logr.Logger {
	return *(*logr.Logger)(atomic.LoadPointer(&globalLogger))
}

func Info(msg string, keysAndValues ...interface{}) {
	getLogger().V(4).Info(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	getLogger().Error(err, msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	getLogger().V(1).Info(msg, keysAndValues...)
}
