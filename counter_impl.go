package metric

import (
	"math"
	"sync"
	"time"
)

// bucket contains interval record of Start/End Count/Sum
type bucket struct {
	end   int64   // end time represent in unit nano-seconds
	count uint64  // count of increments during Start to End
	sum   float64 // sum of increment value during Start to End
	min   float64 // min of the values during Start to End
	max   float64 // max of the values during Start to End
}

// NewCounter creates a counter with the given paramters
func NewCounter(windowDur, bucketDur time.Duration) (Counter, error) {
	if err := check(windowDur, bucketDur); err != nil {
		return nil, err
	}
	// allocate extract bucket for proper cyclic reuse of bucket
	num := int(windowDur/bucketDur + 1)
	// initilaize a new counter
	return &counterImpl{
		buckets:   make([]bucket, num),
		windowDur: int64(windowDur),
		bucketDur: int64(bucketDur),
	}, nil
}

type counterImpl struct {
	buckets      []bucket // ring buffer of bucket
	windowDur    int64    // sliding windows duration
	bucketDur    int64    // bucket duration
	curIdx       int      // curIdx points to current working bucket
	sync.RWMutex          // embeded Read-Write lock to protect bucket ring buffer
}

func (c *counterImpl) Incr(value float64) {
	now := timeNow()
	c.Lock()
	defer c.Unlock()

	cur := &c.buckets[c.curIdx]
	// bucket range still valid
	if now < cur.end {
		cur.count++
		cur.sum += value
		cur.min = math.Min(cur.min, value)
		cur.max = math.Max(cur.max, value)
		return
	}
	// move to next bucket
	c.curIdx = (c.curIdx + 1) % len(c.buckets)
	cur = &c.buckets[c.curIdx]
	cur.end = now - now%c.bucketDur + c.bucketDur
	cur.count = 1
	cur.sum = value
	cur.min = value
	cur.max = value
}

func (c *counterImpl) Snapshot() CounterSnapshot {
	return &counterSnapshot{
		bucketDur: time.Duration(c.bucketDur),
		buckets:   c.getBuckets(),
	}
}

// getBuckets returns bucket list according to the current time in following step
func (c *counterImpl) getBuckets() []bucket {
	now := timeNow()

	// protect the process during bucket scanning from oldest to latest
	c.RLock()
	defer c.RUnlock()

	result := make([]bucket, 0, len(c.buckets))
	i := c.curIdx

	for range c.buckets {
		i = (i + 1) % len(c.buckets)
		// from latest bucket
		b := c.buckets[i]
		if b.end <= now && b.end+c.windowDur > now {
			result = append(result, b)
		}
	}
	return result
}
