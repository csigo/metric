package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	defaultPkg = "test"
)

// TestCase counter
type SuiteClient struct {
	suite.Suite
	factory *mockFactory
	client  *pkgClient
	now     time.Time
}

func (s *SuiteClient) SetupSuite() {
	s.now = time.Now()
	timeNow = func() int64 {
		return s.now.UnixNano()
	}
}

func (s *SuiteClient) SetupTest() {
	s.factory = &mockFactory{}
	s.client = newClient(defaultPkg)

	newCounter = s.factory.NewCounter
	newHistogram = s.factory.NewHist
}

func (s *SuiteClient) TestOneCounter() {
	defer s.add(defaultPkg, "aaa.bbb", 10, 2, 0)()

	p := s.client

	p.BumpAvg("aaa.bbb", 10)
	p.BumpAvg("aaa.bbb", 10)

	ss := p.get("aaa.bbb")

	s.Equal(len(ss), 1)
	s.Equal(ss[0].Name(), "aaa.bbb")
	s.False(ss[0].HasHistogram())
}

func (s *SuiteClient) TestTwoCounters() {
	defer s.add(defaultPkg, "aaa.bbb", 10, 2, 0)()
	defer s.add(defaultPkg, "ccc.ddd", 20, 1, 0)()

	p := s.client

	p.BumpAvg("aaa.bbb", 10)
	p.BumpAvg("aaa.bbb", 10)
	p.BumpAvg("ccc.ddd", 20)

	ss := p.get("aaa.bbb")
	s.Equal(len(ss), 1)
	s.Equal(ss[0].Name(), "aaa.bbb")
	s.False(ss[0].HasHistogram())

	ss = p.get("ccc.ddd")
	s.Equal(len(ss), 1)
	s.Equal(ss[0].Name(), "ccc.ddd")
	s.False(ss[0].HasHistogram())

	ss = p.get("*")
	s.Equal(len(ss), 2)
	s.False(ss[0].HasHistogram())
	s.False(ss[1].HasHistogram())
}

func (s *SuiteClient) TestCounterHist() {
	defer s.add(defaultPkg, "aaa.bbb", 10, 2, 2)()

	p := s.client

	p.BumpHistogram("aaa.bbb", 10)
	p.BumpHistogram("aaa.bbb", 10)

	ss := p.get("aaa.bbb")
	s.Equal(len(ss), 1)
	s.Equal(ss[0].Name(), "aaa.bbb")
	s.True(ss[0].HasHistogram())

	ss = p.get("*")
	s.Equal(len(ss), 1)
	s.True(ss[0].HasHistogram())
}

func (s *SuiteClient) TestCounterThenHist() {
	defer s.add(defaultPkg, "aaa.bbb", 10, 4, 2)()

	p := s.client

	p.BumpAvg("aaa.bbb", 10)
	p.BumpHistogram("aaa.bbb", 10)
	p.BumpAvg("aaa.bbb", 10)
	p.BumpHistogram("aaa.bbb", 10)

	ss := p.get("aaa.bbb")
	s.Equal(len(ss), 1)
	s.True(ss[0].HasHistogram())
}

func (s *SuiteClient) TestBumSum() {
	defer s.add(defaultPkg, "aaa.bbb", 10, 2, 0)()

	p := s.client

	p.BumpAvg("aaa.bbb", 10)
	p.BumpSum("aaa.bbb", 10)

	ss := p.get("aaa.bbb")
	s.Equal(len(ss), 1)
	s.False(ss[0].HasHistogram())
}

func (s *SuiteClient) TestBumpTime() {
	dur := time.Minute
	defer s.add(defaultPkg, "aaa.bbb", float64(dur), 2, 2)()

	p := s.client

	e := p.BumpTime("aaa.bbb")
	s.tick(dur)
	e.End()

	e = p.BumpTime("aaa.bbb")
	s.tick(dur)
	e.End()

	ss := p.get("aaa.bbb")
	s.Equal(len(ss), 1)
	s.True(ss[0].HasHistogram())
}

func (s *SuiteClient) add(pkg, name string, v float64, ct, ht int) func() {
	return addTestData(s.T(), s.factory, pkg, name, v, ct, ht)
}

func (s *SuiteClient) tick(dur time.Duration) {
	s.now = s.now.Add(dur)
}

func TestRunSuiteClient(t *testing.T) {
	suite.Run(t, new(SuiteClient))
}

// mock and fake ---------------------------------------------

func addTestData(t *testing.T, fac *mockFactory, pkg, name string, v float64, ct, ht int) func() {
	mcs := &MockCtrSnapshot{}
	mc := &MockCounter{}
	mc.On("Incr", v).Return().Times(ct)
	mc.On("Snapshot").Return(mcs)
	fac.On("NewCounter", counterParams.window, counterParams.bucket).Return(mc, nil).Once()

	mh := &MockHist{}
	if ht > 0 {
		mhs := &MockHistSnapshot{}
		mh.On("Update", v).Return().Times(ht)
		mh.On("Snapshot").Return(mhs)
		fac.On("NewHist", histogramParams.window, histogramParams.bucket).Return(mh, nil).Once()
	}
	return func() {
		mc.AssertExpectations(t)
		mh.AssertExpectations(t)
	}
}

type mockFactory struct {
	mock.Mock
}

func (m *mockFactory) NewCounter(w, b time.Duration) (Counter, error) {
	args := m.Called(w, b)
	return args.Get(0).(Counter), args.Error(1)
}

func (m *mockFactory) NewHist(w, b time.Duration) (Histogram, error) {
	args := m.Called(w, b)
	return args.Get(0).(Histogram), args.Error(1)
}
