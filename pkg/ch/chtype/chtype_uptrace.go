package chtype

import "math"

type BFloat16 uint16

func ToBFloat16(f float64) BFloat16 {
	return BFloat16(math.Float32bits(float32(f)) >> 16)
}

func (f BFloat16) Float32() float32 {
	return math.Float32frombits(uint32(f) << 16)
}
