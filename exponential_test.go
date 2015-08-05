package metric

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/suite"
)

// SuiteExponential is test suite for histogram percentile
type SuiteExponential struct {
	suite.Suite
}

// TestRunSuiteExponential run SuiteExponential
func TestRunSuiteExponential(t *testing.T) {
	suite.Run(t, new(SuiteExponential))
}

func (s *SuiteExponential) TestMinMax() {
	exp := &exponential{}

	minV, maxV := exp.ValueRange()
	minBin, maxBin := exp.BinRange()
	s.Equal(minV, float64(-maxValue))
	s.Equal(maxV, float64(maxValue))
	s.Equal(minBin, -maxBin)
	s.Equal(maxBin, maxBin)
}

func (s *SuiteExponential) TestRange() {
	exp := &exponential{}
	// min & max
	s.Equal(exp.Bin(0), 0)
	s.Equal(exp.Bin(10*maxValue), maxBin)
	s.Equal(exp.Bin(-10*maxValue), -maxBin)

	for v := float64(1000); v > minFloatValue/10.0; v /= 3.0 {
		idx := exp.Bin(v)
		// check always in buckets
		l, u := exp.Bound(idx)
		s.True(v > l, fmt.Sprintf("%f > %f", v, l))
		s.True(v <= u, fmt.Sprintf("%f <= %f", v, u))
	}

	for v := float64(1); v < 5000; v += 0.5 {
		idx := exp.Bin(v)
		// check always in buckets
		l, u := exp.Bound(idx)
		s.True(v > l, fmt.Sprintf("%f > %f", v, l))
		s.True(v <= u, fmt.Sprintf("%f <= %f", v, u))
	}

	for v := float64(-1000); v < -minFloatValue/10.0; v /= 3.0 {
		idx := exp.Bin(v)
		// check always in buckets
		l, u := exp.Bound(idx)
		s.True(v >= l, fmt.Sprintf("%f >= %f", v, l))
		s.True(v < u, fmt.Sprintf("%f < %f", v, u))
	}

	for v := float64(-1); v > -5000; v -= 0.5 {
		idx := exp.Bin(v)
		// check always in buckets
		l, u := exp.Bound(idx)
		s.True(v >= l, fmt.Sprintf("%f >= %f", v, l))
		s.True(v < u, fmt.Sprintf("%f < %f", v, u))
	}

}

func (s *SuiteExponential) TestBound() {
	exp := &exponential{}
	r := math.Pow(10, 0.1)

	s.Panics(func() {
		exp.Bound(-maxBin - 1)
	})
	s.Panics(func() {
		exp.Bound(maxBin + 1)
	})

	// first bin
	l, u := exp.Bound(0)
	assertNearEqual(s.T(), l, 0)
	assertNearEqual(s.T(), u, 0)

	l, u = exp.Bound(1)
	assertNearEqual(s.T(), l, 0)
	assertNearEqual(s.T(), u, minFloatValue)

	l, u = exp.Bound(-1)
	assertNearEqual(s.T(), l, -minFloatValue)
	assertNearEqual(s.T(), u, 0)

	l, u = exp.Bound(2)
	assertNearEqual(s.T(), l, minFloatValue)
	assertNearEqual(s.T(), u, minFloatValue*r)

	l, u = exp.Bound(-2)
	assertNearEqual(s.T(), l, -minFloatValue*r)
	assertNearEqual(s.T(), u, -minFloatValue)

	// max value bin
	l, u = exp.Bound(maxBin - 1)
	assertNearEqual(s.T(), l, maxValue/r)
	assertNearEqual(s.T(), u, maxValue)

	l, u = exp.Bound(-maxBin + 1)
	assertNearEqual(s.T(), l, -maxValue)
	assertNearEqual(s.T(), u, -maxValue/r)

	// last bin
	l, u = exp.Bound(maxBin)
	assertNearEqual(s.T(), l, maxValue)
	assertNearEqual(s.T(), u, math.MaxFloat64)

	l, u = exp.Bound(-maxBin)
	assertNearEqual(s.T(), l, -math.MaxFloat64)
	assertNearEqual(s.T(), u, -maxValue)
}
