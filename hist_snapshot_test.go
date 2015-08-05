package metric

import (
	"math"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	testPercentiles = []float64{0.5, 0.75, 1.0}
)

// TestCase counter
type SuiteHistSnapshot struct {
	suite.Suite
	sh *histSnapshot
}

func (s *SuiteHistSnapshot) SetupTest() {
	bound := &exponential{}

	s.sh = &histSnapshot{
		bins:  []binVal{},
		bound: bound,
	}
}

func (s *SuiteHistSnapshot) add(value float64, cnt int64) {
	sh := s.sh
	bins := sh.bins

	idx := sh.bound.Bin(value)
	for i := range bins {
		if bins[i].bin == idx {
			bins[i].count += cnt
			return
		}
	}
	sh.bins = append(sh.bins, binVal{bin: idx, count: cnt})
}

// TestEmpty test empty histogram
func (s *SuiteHistSnapshot) TestEmpty() {
	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(0))
	assertNearEqual(s.T(), 0, v[0])
	assertNearEqual(s.T(), 0, v[1])
	assertNearEqual(s.T(), 0, v[2])
}

func (s *SuiteHistSnapshot) TestAllZero() {
	s.add(0, 100)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(100))
	assertNearEqual(s.T(), 0, v[0])
	assertNearEqual(s.T(), 0, v[1])
	assertNearEqual(s.T(), 0, v[2])
}

// TestOneItem test only one item
func (s *SuiteHistSnapshot) TestOneItem() {
	s.add(100, 1)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(1))
	assertNearEqual(s.T(), 89.71641173621408, v[0])
	assertNearEqual(s.T(), 89.71641173621408, v[1])
	assertNearEqual(s.T(), 89.71641173621408, v[2])
}

func (s *SuiteHistSnapshot) TestSmallValue() {
	s.add(0.00001, 1)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(1))
	assertNearEqual(s.T(), 8.971641173621403e-06, v[0])
	assertNearEqual(s.T(), 8.971641173621403e-06, v[1])
	assertNearEqual(s.T(), 8.971641173621403e-06, v[2])
}

// TestOneItem test only one item
func (s *SuiteHistSnapshot) TestNegativeOneItem() {
	s.add(-100, 1)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(1))
	assertNearEqual(s.T(), -89.71641173621408, v[0])
	assertNearEqual(s.T(), -89.71641173621408, v[1])
	assertNearEqual(s.T(), -89.71641173621408, v[2])
}

// TestDuplicated test only one item
func (s *SuiteHistSnapshot) TestDuplicated() {
	s.add(100, 100)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(100))
	assertNearEqual(s.T(), 89.61357585357622, v[0])
	assertNearEqual(s.T(), 94.75536998546919, v[1])
	assertNearEqual(s.T(), 99.89716411736214, v[2])
}

func (s *SuiteHistSnapshot) TestUniformed() {
	for i := 0.0; i < 1000; i += 10 {
		s.add(i, 1)
	}

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(100))
	assertNearEqual(s.T(), 487.13086138993947, v[0])
	assertNearEqual(s.T(), 738.1694912028767, v[1])
	assertNearEqual(s.T(), 994.858205868107, v[2])
}

func (s *SuiteHistSnapshot) TestAllMin() {
	s.add(-maxValue-1, 100)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(100))
	assertNearEqual(s.T(), -maxValue, v[0])
	assertNearEqual(s.T(), -maxValue, v[1])
	assertNearEqual(s.T(), -maxValue, v[2])
}

func (s *SuiteHistSnapshot) TestAllMax() {
	s.add(maxValue+1, 100)

	v, count := s.sh.Percentiles(testPercentiles)
	s.Equal(count, int64(100))
	assertNearEqual(s.T(), maxValue, v[0])
	assertNearEqual(s.T(), maxValue, v[1])
	assertNearEqual(s.T(), maxValue, v[2])
}

func (s *SuiteHistSnapshot) TestInvalidPercentiles() {
	s.add(maxValue+1, 100)

	v, count := s.sh.Percentiles([]float64{-1, 3, -0.6})
	s.Equal(count, int64(100))
	s.True(math.IsNaN(v[0]))
	s.True(math.IsNaN(v[1]))
	s.True(math.IsNaN(v[2]))
}

func (s *SuiteHistSnapshot) TestBins() {
	s.add(0, 3)
	s.add(0.003, 5)
	s.add(100, 1)
	s.add(130, 2)
	s.add(12341234, 2)

	sbins := s.sh.Bins()
	for i, b := range s.sh.bins {
		s.Equal(sbins[i].Count, b.count)
		l, u := s.sh.bound.Bound(b.bin)
		assertNearEqual(s.T(), sbins[i].Lower, l)
		assertNearEqual(s.T(), sbins[i].Upper, u)
	}
}

func TestRunSuiteHistSnapshot(t *testing.T) {
	suite.Run(t, new(SuiteHistSnapshot))
}
