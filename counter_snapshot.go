package metric

import (
	"math"
	"time"
)

// counterSnapshot represents a counter snapshot
type counterSnapshot struct {
	bucketDur time.Duration
	buckets   []bucket
}

// SliceIn returns statistics of each bucket in the given duration
func (c *counterSnapshot) SliceIn(dur time.Duration) []Bucket {
	result := make([]Bucket, 0, len(c.buckets))
	lowerBound := timeNow() - int64(dur)
	for _, b := range c.buckets {
		if b.end < lowerBound {
			continue
		}
		r := Bucket{
			Count: float64(b.count),
			Sum:   b.sum,
			Min:   b.min,
			Max:   b.max,
			Start: time.Unix(0, int64(b.end)).Add(-c.bucketDur),
			End:   time.Unix(0, int64(b.end)),
		}
		if r.Count > 0 {
			r.Avg = r.Sum / float64(r.Count)
		}
		result = append(result, r)
	}
	return result
}

// AggrIn returns aggregration statistics in the given duration
func (c *counterSnapshot) AggrIn(dur time.Duration) Bucket {
	lowerBound := timeNow() - int64(dur)
	var cnt uint64
	var sum, min, max, avg float64
	var minEnd, maxEnd int64 = math.MaxInt64, 0

	for _, b := range c.buckets {
		if b.end < lowerBound {
			continue
		}
		if cnt > 0 {
			min = math.Min(min, b.min)
			max = math.Max(max, b.max)
		} else {
			min = b.min
			max = b.max
		}
		cnt += b.count
		sum += b.sum
		if b.end < minEnd {
			minEnd = b.end
		}
		if b.end > maxEnd {
			maxEnd = b.end
		}
	}
	if minEnd > maxEnd {
		return Bucket{}
	}
	if cnt > 0 {
		avg = sum / float64(cnt)
	}
	return Bucket{
		Count: float64(cnt),
		Sum:   sum,
		Min:   min,
		Max:   max,
		Avg:   avg,
		Start: time.Unix(0, int64(minEnd)).Add(-c.bucketDur),
		End:   time.Unix(0, int64(maxEnd)),
	}
}
