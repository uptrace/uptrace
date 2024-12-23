//go:build windows

package unixtime

import "time"

func now() int64 { return time.Now().UnixNano() }
