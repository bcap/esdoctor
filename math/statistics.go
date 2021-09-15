package math

import (
	"sort"
)

// Calculates percentiles for a given int slice. The buckets argument controls
// how precise the percentile calculations are. Additionally, the first bucket
// value is the minimum value in the dataset (p0) while the last bucket value
// is the maximum value in the dataset (p100). For that reason ranges are
// inclusive (eg [0,100]) and the returned slice will have its length being buckets+1
//
// NOTE on input correctness:
//  - the input slice needs to have at least one element
//  - buckets needs to be a value greater than zero
// In case any passed input is invalid, a nil value is returned
//
// Examples:
// For instance, to calculate percentiles in steps of 10 points (p10, p20, p30, ...):
//   out := PercentilesInt(slice, 10) // out will have length of 11
//   out[0]  // min
//   out[1]  // p10
//   out[5]  // p50
//   out[10] // max
// Now for steps of 1 point (p1, p2, p3, etc):
//   out := PercentilesInt(slice, 100) // out will have length of 101
//   out[0]   // min
//   out[2]   // p2
//   out[50]  // p50
//   out[100] // max
// Now for steps of .1 point (p1.1, p99.9, etc):
//   out := PercentilesInt(slice, 1000) // out will have length of 1001
//   out[0]    // min
//   out[20]   // p2
//   out[501]  // p50.1
//   out[999]  // p99.9
//   out[1000] // max

func PercentilesInt(slice []int, buckets int) []int {
	if len(slice) == 0 || buckets < 1 {
		return nil
	}
	cp := make([]int, len(slice))
	copy(cp, slice)
	slice = cp
	sort.Ints(slice)
	result := make([]int, buckets+1)
	for i := 0; i < buckets; i++ {
		result[i] = slice[percentileIdx(i, buckets, len(slice))]
	}
	result[buckets] = slice[len(slice)-1]
	return result
}

// Same as PercentilesInt, but for int64
func PercentilesInt64(slice []int64, buckets int) []int64 {
	if len(slice) == 0 || buckets < 1 {
		return nil
	}
	cp := make([]int64, len(slice))
	copy(cp, slice)
	slice = cp
	sort.Slice(slice, func(i, j int) bool { return slice[i] < slice[j] })
	result := make([]int64, buckets+1)
	for i := 0; i < buckets; i++ {
		result[i] = slice[percentileIdx(i, buckets, len(slice))]
	}
	result[buckets] = slice[len(slice)-1]
	return result
}

// Same as PercentilesInt, but for float64
func PercentilesFloat64(slice []float64, buckets int) []float64 {
	if len(slice) == 0 || buckets < 1 {
		return nil
	}
	cp := make([]float64, len(slice))
	copy(cp, slice)
	slice = cp
	sort.Float64s(slice)
	result := make([]float64, buckets+1)
	for i := 0; i < buckets; i++ {
		result[i] = slice[percentileIdx(i, buckets, len(slice))]
	}
	result[buckets] = slice[len(slice)-1]
	return result
}

func percentileIdx(idx int, buckets int, length int) int {
	return int(float64(idx) / float64(buckets) * float64(length))
}
