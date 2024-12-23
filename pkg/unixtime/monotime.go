package unixtime

import _ "unsafe"

func Monotime() int64 { return nanotime() }

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64
