package metric

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestCase counter
type SuiteSimpleCounter struct {
	suite.Suite
	counter *simpleCounter
}

func (s *SuiteSimpleCounter) SetupSuite() {
	timeNow = func() int64 {
		return curTimestamp
	}
}

func (s *SuiteSimpleCounter) SetupTest() {
	s.counter = newSimpleCounter(defaultWindow, defaultBucket)
}

func (s *SuiteSimpleCounter) TestGetBucket() {
	c := s.counter
	sum := int64(0)

	for i := 1; i <= defaultBucketNum; i++ {
		for j := 1; j <= i; j++ {
			c.incr()
		}
		sum += int64(i)
		s.Equal(sum, c.get())
		tick(defaultBucket)
	}
	// start to expire
	for i := 1; i <= defaultBucketNum; i++ {
		for j := 1; j <= i; j++ {
			c.incr()
			c.incr()
		}
		// add 2, but expire 1
		sum += int64(i)
		s.Equal(sum, c.get())
		tick(defaultBucket)
	}
	for i := 1; i <= defaultBucketNum; i++ {
		tick(defaultBucket)
	}
	s.Equal(c.get(), int64(0))
}

func TestRunSuiteSimpleCounter(t *testing.T) {
	suite.Run(t, new(SuiteSimpleCounter))
}
