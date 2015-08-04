package metric

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// SuiteImpl is test suite for histogram percentile
type SuiteImpl struct {
	suite.Suite
	hist *histImpl
}

// TestRunSuiteImpl run SuiteImpl
func TestRunSuiteImpl(t *testing.T) {
	suite.Run(t, new(SuiteImpl))
}

func (s *SuiteImpl) SetupTest() {
	hist, _ := NewHistogram(defaultWindow, defaultBucket)
	s.hist = hist.(*histImpl)
}

func (s *SuiteImpl) TestCreate() {
	_, err := NewHistogram(defaultWindow, defaultBucket)
	s.NoError(err)

	// invalid decay period
	_, err = NewHistogram(time.Millisecond, time.Second)
	s.Error(err)
}

func (s *SuiteImpl) TestBins() {
	s.hist.Update(150)
	s.hist.Update(160)
	s.hist.Update(100000)
	s.hist.Update(100003)
	s.hist.Update(100103)
	s.hist.Update(101010101011010) // > max

	exp := []binVal{
		binVal{bin: 83, count: 1},
		binVal{bin: 84, count: 1},
		binVal{bin: 111, count: 1},
		binVal{bin: 112, count: 2},
		binVal{bin: 162, count: 1},
	}

	s.Equal(exp, s.hist.values())
}

func (s *SuiteImpl) TestSnapshot() {
	hist := s.hist
	hist.Update(1000)
	hist.Update(100)
	hist.Update(100000)

	sh := hist.Snapshot().(*histSnapshot)
	s.Equal(sh.bound, &exponential{})
	s.Equal(sh.bins, []binVal{
		binVal{bin: 81, count: 1},
		binVal{bin: 91, count: 1},
		binVal{bin: 111, count: 1},
	})
}

func assertNearEqual(t *testing.T, a float64, b float64) {
	precision := -3.0
	switch {
	case math.Abs(a) < math.Pow(10, -6):
		precision = -8.0
	case math.Abs(a) < math.Pow(10, -3):
		precision = -6.0
	}
	assert.True(t, math.Abs(a-b) < math.Pow(10, precision), fmt.Sprintf("%v != %v", a, b))
}
