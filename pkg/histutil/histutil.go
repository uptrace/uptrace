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

func BuildHeatmap(tdig [][]float32, tm []time.Time) *Heatmap {
	if len(tm) == 0 {
		return &Heatmap{
			XAxis: make([]time.Time, 0),
			YAxis: make([][2]float64, 0),
			Data:  make([][3]uint32, 0),
		}
	}

	dmin := math.MaxFloat64
	var dmax float64
	for _, td := range tdig {
		for i := 0; i < len(td); i += 2 {
			n := float64(td[i])
			if n < dmin {
				dmin = n
			}
			if n > dmax {
				dmax = n
			}
		}
	}

	div := make([]float64, 17)
	if dmin <= 0 || dmax-dmin <= 1 {
		floats.Span(div, dmin, dmax)
	} else {
		floats.LogSpan(div, dmin, dmax)
	}
	div[0] = dmin
	if last := div[len(div)-1]; last <= dmax {
		div[len(div)-1] = math.Nextafter(dmax, 2*dmax)
	}

	heatmap := make([][3]uint32, 0, len(tm))

	hist := NewHist(div)
	for xi := range tm {
		td := tdig[xi]
		if len(td) == 0 {
			continue
		}
		cs := make([]uint32, hist.NumBin())
		for j := 0; j < len(td); j += 2 {
			n := uint32(td[j+1])
			if n == 0 {
				continue
			}
			cs[hist.Index(float64(td[j]))] += n
		}
		for yi, n := range cs {
			if n == 0 {
				continue
			}
			heatmap = append(heatmap, [3]uint32{uint32(xi), uint32(yi), n})
		}
	}

	return &Heatmap{
		XAxis: tm,
		YAxis: hist.Bins(),
		Data:  heatmap,
	}
}

type Hist struct {
	bins [][2]float64
}

func NewHist(div []float64) Hist {
	bins := make([][2]float64, 0, len(div)-1)

	for i := 1; i < len(div); i++ {
		bins = append(bins, [2]float64{
			div[i-1],
			div[i],
		})
	}

	return Hist{
		bins: bins,
	}
}

func (h Hist) NumBin() int {
	return len(h.bins)
}

func (h Hist) Index(x float64) int {
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
