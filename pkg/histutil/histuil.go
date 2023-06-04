package histutil

import (
	"math"
	"time"

	"gonum.org/v1/gonum/floats"
)

type Heatmap struct {
	XAxis []time.Time  `json:"xAxis"`
	YAxis [][2]float64 `json:"yAxis"`
	Data  [][3]uint32  `json:"data"`
}

func BuildHeatmap(tdigest [][]float32, timeCol []time.Time) *Heatmap {
	const numBin = 16

	if len(timeCol) == 0 {
		return &Heatmap{
			XAxis: make([]time.Time, 0),
			YAxis: make([][2]float64, 0),
			Data:  make([][3]uint32, 0),
		}
	}

	dataMin := math.MaxFloat64
	var dataMax float64

	for _, td := range tdigest {
		for i := 0; i < len(td); i += 2 {
			mean := float64(td[i])
			if mean < dataMin {
				dataMin = mean
			}
			if mean > dataMax {
				dataMax = mean
			}
		}
	}

	dividers := Dividers(numBin, dataMin, dataMax)
	heatmap := make([][3]uint32, 0, len(timeCol))
	hist := NewHist(dividers)

	for xIndex := range timeCol {
		td := tdigest[xIndex]
		if len(td) == 0 {
			continue
		}

		counts := make([]uint32, hist.NumBin())

		for j := 0; j < len(td); j += 2 {
			count := uint32(td[j+1])
			if count == 0 {
				continue
			}

			mean := float64(td[j])
			index := hist.BinIndex(mean)
			counts[index] += count
		}

		for yIndex, count := range counts {
			if count == 0 {
				continue
			}
			heatmap = append(heatmap, [3]uint32{uint32(xIndex), uint32(yIndex), count})
		}
	}

	return &Heatmap{
		XAxis: timeCol,
		YAxis: hist.Bins(),
		Data:  heatmap,
	}
}

type Hist struct {
	bins [][2]float64
}

func NewHist(dividers []float64) Hist {
	bins := make([][2]float64, 0, len(dividers)-1)

	for i := 1; i < len(dividers); i++ {
		bins = append(bins, [2]float64{
			dividers[i-1],
			dividers[i],
		})
	}

	return Hist{
		bins: bins,
	}
}

func (h Hist) NumBin() int {
	return len(h.bins)
}

func (h Hist) BinIndex(x float64) int {
	for i, bin := range h.bins {
		if x >= bin[0] && x < bin[1] {
			return i
		}
	}
	return len(h.bins) - 1
}

func (h Hist) Bins() [][2]float64 {
	return h.bins
}

func Dividers(n int, l, u float64) []float64 {
	fs := span(n, l, u)

	fs[0] = l
	if last := fs[len(fs)-1]; last <= u {
		fs[len(fs)-1] = math.Nextafter(u, 2*u)
	}

	return fs
}

func span(n int, l, u float64) []float64 {
	fs := make([]float64, n+1)
	if l <= 0 || u-l <= 1 {
		return floats.Span(fs, l, u)
	}
	return floats.LogSpan(fs, l, u)
}
