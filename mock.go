package metric

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockCtrSnapshot is mock object of CounterSnapshot
type MockCtrSnapshot struct {
	mock.Mock
}

// SliceIn mocks SliceIn()
func (m *MockCtrSnapshot) SliceIn(dur time.Duration) []Bucket {
	return m.Called(dur).Get(0).([]Bucket)
}

// AggrIn mocks AggrIn()
func (m *MockCtrSnapshot) AggrIn(dur time.Duration) Bucket {
	return m.Called(dur).Get(0).(Bucket)
}

// MockHistSnapshot is mock object of HistogramSnapshot
type MockHistSnapshot struct {
	mock.Mock
}

// Bins mocks Bins()
func (m *MockHistSnapshot) Bins() []Bin {
	return m.Called().Get(0).([]Bin)
}

// Percentiles mocks Percentiles()
func (m *MockHistSnapshot) Percentiles(ps []float64) ([]float64, int64) {
	args := m.Called(ps)
	return args.Get(0).([]float64), args.Get(1).(int64)
}

// MockCounter is mock object of Counter
type MockCounter struct {
	mock.Mock
}

// Incr mocks Incr()
func (m *MockCounter) Incr(value float64) {
	m.Called(value)
}

// Snapshot mocks Snapshot()
func (m *MockCounter) Snapshot() CounterSnapshot {
	args := m.Called()
	return args.Get(0).(CounterSnapshot)
}

// MockHist is mock object of Histogram
type MockHist struct {
	mock.Mock
}

// Update mocks Update()
func (m *MockHist) Update(value float64) {
	m.Called(value)
}

// Snapshot mocks Snapshot()
func (m *MockHist) Snapshot() HistSnapshot {
	args := m.Called()
	return args.Get(0).(HistSnapshot)
}
