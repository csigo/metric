package metric

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	// customized curTimestamp to control the timing
	curTimestamp int64
)

func tick(d time.Duration) {
	curTimestamp += int64(d)
}

// Test releated default value
const (
	defaultWindow    = time.Hour
	defaultBucket    = time.Minute
	defaultBucketNum = int(defaultWindow / defaultBucket)
)

// TestCase counter
type SuiteCounter struct {
	suite.Suite
	counter *counterImpl
}

func (s *SuiteCounter) SetupSuite() {
	timeNow = func() int64 {
		return curTimestamp
	}
}

func (s *SuiteCounter) SetupTest() {
	c, _ := NewCounter(defaultWindow, defaultBucket)
	s.counter = c.(*counterImpl)
}

func (s *SuiteCounter) TestCreate() {
	_, err := NewCounter(time.Minute, 5*time.Second)
	s.NoError(err)

	// window must >= 1 minute
	_, err = NewCounter(10*time.Second, 5*time.Second)
	s.Error(err)

	// bucket must >= 2 second
	_, err = NewCounter(time.Minute, time.Second)
	s.Error(err)

	// window is multiple of bucket
	_, err = NewCounter(time.Minute, 7*time.Second)
	s.Error(err)

	// window must >= bucket
	_, err = NewCounter(time.Minute, 5*time.Minute)
	s.Error(err)
}

func (s *SuiteCounter) TestConcurrent() {
	c := s.counter
	wg := sync.WaitGroup{}
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			c.Incr(3)
			c.getBuckets()
			wg.Done()
		}()
	}

	wg.Wait()
	s.Equal(c.buckets[c.curIdx].count, 100)
	s.Equal(c.buckets[c.curIdx].sum, 300)
}

type offset struct {
	start int64
	end   int64
	m     float64
}

func (s *SuiteCounter) TestGetBucket() {
	c := s.counter
	bucketInterval := int64(defaultBucket)
	now := timeNow()
	end := now - now%bucketInterval + bucketInterval
	record := []offset{}
	for i := 0; i < defaultBucketNum; i++ {
		// Tick of next bucket interval
		m := float64(i + 1)
		c.Incr(1 * m)
		c.Incr(4 * m)
		c.Incr(6 * m)
		tick(defaultBucket)
		record = append(record, offset{m: m, end: end})
		end += bucketInterval
	}
	b := c.getBuckets()
	s.Equal(len(b), defaultBucketNum)
	for i := 0; i < len(b); i++ {
		s.Equal(b[i].end, record[i].end)
		s.Equal(b[i].count, 3)
		s.Equal(b[i].sum, 11*record[i].m)
		s.Equal(b[i].min, 1*record[i].m)
		s.Equal(b[i].max, 6*record[i].m)
	}
}

func (s *SuiteCounter) TestGetBucketReadonly() {
	c := s.counter
	c.Incr(1)
	tick(defaultBucket)
	c.getBuckets()[0].count = 8
	// change value of GetBucket() won't affect real counter value
	s.Equal(1, c.getBuckets()[0].count)
}

func (s *SuiteCounter) TestSnapshot() {
	c := s.counter
	c.Incr(1)
	tick(defaultBucket)
	c.Incr(2)
	tick(defaultBucket)
	c.Incr(3)
	tick(defaultBucket)

	sh := c.Snapshot().(*counterSnapshot)
	s.Equal(sh.buckets, c.getBuckets())
}

func TestRunSuiteCounter(t *testing.T) {
	suite.Run(t, new(SuiteCounter))
}
