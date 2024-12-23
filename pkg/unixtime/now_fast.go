//go:build aix || darwin || dragonfly || freebsd || (js && wasm) || linux || nacl || netbsd || openbsd || solaris

package unixtime

import (
	"syscall"
	"time"
)

func now() int64 {
	var tv syscall.Timeval
	if err := syscall.Gettimeofday(&tv); err != nil {
		return time.Now().UnixNano()
	}
	return syscall.TimevalToNsec(tv)
}
