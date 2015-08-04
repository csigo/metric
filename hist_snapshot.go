package metric

import (
	"math"
	"sort"
)

type histSnapshot struct {
	bins  []binVal
	bound binBound
}

// Values list current histogram totaol values
func (h *histSnapshot) Bins() []Bin {
	result := make([]Bin, len(h.bins))
	for i, b := range h.bins {
		l, u := h.bound.Bound(b.bin)
		result[i] = Bin{
			Count: b.count,
			Lower: l,
			Upper: u,
		}
	}
	return result
}

// Percentiles returns exported percentile with give window size
func (h *histSnapshot) Percentiles(ps []float64) ([]float64, int64) {
	result := make([]float64, len(ps))
	if len(ps) == 0 {
		return result, 0
	}
	// calculate cumulated rank position range for each bin
	// the smallest value is rank first.
	cumCount := h.bins
	if len(cumCount) == 0 {
		return result, 0
	}
	// cal cumulate count
	count := int64(0)
	for i := range cumCount {
		count += cumCount[i].count
		if i > 0 {
			cumCount[i].count += cumCount[i-1].count
		}
	}
	// no value
	if count == 0 {
		return result, 0
	}

	// get range
	minBin, maxBin := h.bound.BinRange()
	for pIdx, p := range ps {
		if p > 1 || p < 0 {
			result[pIdx] = math.NaN()
			continue
		}
		// calculate the number of count matches p percentile
		// refer http://en.wikipedia.org/wiki/Percentile
		pCount := math.Ceil(float64(count) * float64(p))
		// find out pCount locate in which bin (cumulated slice imply sorted)
		cumIdx := sort.Search(len(cumCount),
			func(k int) bool { return cumCount[k].count >= int64(pCount) })
		pBin := cumCount[cumIdx].bin
		// Handle min & max
		if pBin == minBin {
			result[pIdx], _ = h.bound.ValueRange()
			continue
		}
		if pBin == maxBin {
			_, result[pIdx] = h.bound.ValueRange()
			continue
		}
		// assume values are evenly distributed in each bin
		// value range of the bin is [lValue, uValue] with rank [lCount, uCount]
		// Ranks are descreted, Values are continous, so need to pad 0.5 in left and right)
		//   (pCount - lCount + 0.5)       result - lValue
		// --------------------------- = -------------------
		//    (uCount - lCount + 1)        uValue - lValue
		lCount := 1.0
		if cumIdx > 0 {
			lCount = float64(cumCount[cumIdx-1].count) + 1
		}
		uCount := float64(cumCount[cumIdx].count)
		lValue, uValue := h.bound.Bound(pBin)
		result[pIdx] = (pCount-lCount+0.5)*(uValue-lValue)/(uCount-lCount+1.0) + lValue
	}
	return result, count
}
