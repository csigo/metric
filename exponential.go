package metric

import (
	"fmt"
	"math"
)

/*
exponential implements exponential-sized bin with range [-10^10, 10^10]
here uses (10th root of 10 = 1.258925412) as exponentail base, which means
bucket size will grow 10 times every 10 bins.
*/
const (
	maxValue      = 10000000000 // 10^10
	minValue      = -maxValue   // -10^10
	minFloatValue = 0.000001    // 10^-6
)

var (
	// expCutoff stores upper bound of positive bin ids. Negative bin ids are
	// symmetric to positive ones
	expCutoff = map[int]float64{}
	base      = 0.1                              // math.Log(10^0.1, 10) = 0.1
	minLogV   = math.Log10(minFloatValue) / base // log(10^-6, 10^0.1) = -60
	maxLogV   = math.Log10(maxValue) / base      // log(10^10, 10^0.1)  = 100
	maxBin    = int(maxLogV-minLogV+1) + 1
)

func init() {
	i := 0
	expCutoff[i] = 0
	for i = 0; i < maxBin; i++ {
		expCutoff[i+1] = math.Pow(10, (float64(i)+minLogV)*0.1)
	}
	// last bin stores value exceed maxValue
	expCutoff[maxBin] = math.MaxFloat64
}

// exponential implements exponential-sized buckets with range [0, 10^8]
// here uses (10th root of 10 = 1.258925412) as exponentail base, which means
// bucket size will grow up 10 times every 10 buckets
type exponential struct {
}

// ValueRange returns min and max value
func (*exponential) ValueRange() (float64, float64) {
	return minValue, maxValue
}

// BinRange returns min and max bin ID
func (*exponential) BinRange() (int, int) {
	return -maxBin, maxBin
}

// Bin returns bin id the given value belong to
// def r = 10^0.1
//   61: [-1, -1/r)
//   11: [-10^-5, -10^-5/r)
//   2:  [-r^2*10^-6, -r*10^-6)
//   -1: [-10^-6, 0)
//   0:  [0, 0]
//   1:  (0, 10^-6]
//   2:  (r*10^-6, r^2*10^-6]
//   11: (10^-5/r, 10^-5)
//   61: (1/r, 1]
func (*exponential) Bin(v float64) int {
	if v >= 0 {
		return bin(v)
	}
	return -bin(-v)
}

// Bound returns the upper and lower bound of given bin index
func (*exponential) Bound(bin int) (float64, float64) {
	switch {
	case bin > maxBin || bin < -maxBin:
		panic(fmt.Errorf("index is out of range %d", bin))
	case bin == 0:
		return 0, 0
	case bin > 0:
		return expCutoff[bin-1], expCutoff[bin]
	default:
		bin = -bin
		return -expCutoff[bin], -expCutoff[bin-1]
	}
}

func bin(v float64) int {
	// do log10(v) first to target a smaller bin range for binary search
	switch {
	case v == 0:
		return 0
	case v <= minFloatValue:
		return 1
	case v > maxValue:
		return maxBin
	default:
		return int(math.Ceil(math.Log10(v)/base-minLogV)) + 1
	}
}
