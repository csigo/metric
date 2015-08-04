package metric

// Histogram is a libary for recording data distribution.

import (
	"sort"
	"sync"
	"time"
)

// NewHistogram accepts StatsName, histogram parameter min, max slice, and export flag
func NewHistogram(windowDur, bucketDur time.Duration) (Histogram, error) {
	if err := check(windowDur, bucketDur); err != nil {
		return nil, err
	}
	return &histImpl{
		bucketDur: bucketDur,
		windowDur: windowDur,
		binMap:    map[int]*simpleCounter{},
		bins:      []int{},
		bound:     &exponential{},
	}, nil
}

type binVal struct {
	bin   int
	count int64
}

type histImpl struct {
	bucketDur    time.Duration
	windowDur    time.Duration
	bound        binBound               // binBound manage bucket range and size
	binMap       map[int]*simpleCounter // binMap maps bin to edc
	bins         []int                  // bins stored all used bin ids
	sync.RWMutex                        // rwMutext protects binMap and bins
}

// Update incr the corresponding bin in the histogram
func (h *histImpl) Update(value float64) {
	idx := h.bound.Bin(value)

	h.Lock()
	defer h.Unlock()
	b, ok := h.binMap[idx]
	if ok {
		b.incr()
		return
	}
	b = newSimpleCounter(h.windowDur, h.bucketDur)
	h.binMap[idx] = b
	// TODO: better data-struct
	h.bins = append(h.bins, idx)
	sort.Ints(h.bins)
	b.incr()
}

func (h *histImpl) Snapshot() HistSnapshot {
	return &histSnapshot{
		bound: h.bound,
		bins:  h.values(),
	}
}

func (h *histImpl) values() []binVal {
	h.RLock()
	defer h.RUnlock()

	values := make([]binVal, 0, len(h.bins))
	for _, i := range h.bins {
		values = append(values, binVal{
			bin:   i,
			count: h.binMap[i].get(),
		})
	}
	return values
}

type binBound interface {
	// ValueRange returns min and max value
	ValueRange() (float64, float64)
	// BinRange returns min and max bin ID
	BinRange() (int, int)
	// Bin returns bin ID of the given value
	Bin(v float64) int
	// Bound returns the upper and lower bound of given bin index
	Bound(bin int) (float64, float64)
}
