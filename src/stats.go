package main

import (
	"math"
	"sort"
)

type Stats struct {
	Min, Q1, Median, Q3, Max float64
	Mean, Stddev             float64
}

func stats(xs []float64) Stats {
	n := len(xs)
	if n == 0 {
		return Stats{}
	}
	cp := make([]float64, n)
	copy(cp, xs)
	sort.Float64s(cp)

	mean := 0.0
	for _, x := range cp {
		mean += x
	}
	mean /= float64(n)

	ss := 0.0
	for _, x := range cp {
		d := x - mean
		ss += d * d
	}
	std := 0.0
	if n > 1 {
		std = math.Sqrt(ss / float64(n-1))
	}

	return Stats{
		Min:    cp[0],
		Q1:     cp[n/4],
		Median: cp[n/2],
		Q3:     cp[(3*n)/4],
		Max:    cp[n-1],
		Mean:   mean,
		Stddev: std,
	}
}

// Harmonic mean used for TEPS as specified by the Graph500 benchmark
// (rates are averaged using the harmonic mean)
func harmonic_mean(xs []float64) float64 {
	sumInv := 0.0
	k := 0
	for _, x := range xs {
		if x > 0 {
			sumInv += 1.0 / x
			k++
		}
	}
	if k == 0 {
		return 0
	}
	return float64(k) / sumInv
}
