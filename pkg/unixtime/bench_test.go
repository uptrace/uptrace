package unixtime

import (
	"testing"
	"time"
)

func BenchmarkNowSyscall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}
func BenchmarkNowTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}
func BenchmarkMonotime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Monotime()
	}
}
func BenchmarkNowFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NowFast()
	}
}
