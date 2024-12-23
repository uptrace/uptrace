package idgen

import (
	"github.com/zeebo/xxh3"
	"os"
	"time"
)

func seed1() uint64 { return uint64(time.Now().UnixNano()) }
func seed2() uint64 {
	hostname, _ := os.Hostname()
	return xxh3.HashStringSeed(hostname, uint64(time.Now().UnixNano()))
}
