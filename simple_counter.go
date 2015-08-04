package metric

import (
	"time"
)

func newSimpleCounter(windowDur, bucketDur time.Duration) *simpleCounter {
	return &simpleCounter{
		bucketDur: int64(bucketDur),
		buckets:   make([]buckets, int(windowDur/bucketDur)),
	}
}

type buckets struct {
	end   int64
	count int64
}

// simpleCounter is a none thread-safe implementation of
// bucket-window counter and only records count
type simpleCounter struct {
	bucketDur int64 // bucket duration
	buckets   []buckets
	index     int
}

// incr incrases the count by 1
func (c *simpleCounter) incr() {
	now := timeNow()

	cur := &c.buckets[c.index]
	if now < cur.end {
		cur.count++
		return
	}
	// cylically initilaize next bucket
	c.index = (c.index + 1) % len(c.buckets)
	cur = &c.buckets[c.index]
	cur.end = now - now%c.bucketDur + c.bucketDur
	cur.count = 1
}

// get returns total count of all buckets
func (c *simpleCounter) get() int64 {
	now := timeNow()

	sum := int64(0)
	w := int64(len(c.buckets)) * c.bucketDur
	for _, b := range c.buckets {
		if b.end+w > now {
			sum += b.count
		}
	}
	return sum
}
