package metric

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestCase counter
type SuiteAPI struct {
	suite.Suite
	factory *mockFactory
}

func (s *SuiteAPI) SetupTest() {
	pkgClis = map[string]*pkgClient{}
	s.factory = &mockFactory{}

	newCounter = s.factory.NewCounter
	newHistogram = s.factory.NewHist
}

func (s *SuiteAPI) TestNewClient() {
	defer s.add("aaa", "ccc.ddd", 10, 2, 1)()

	c := NewClient("aaa", "")
	c.BumpAvg("ccc.ddd", 10)
	c.BumpHistogram("ccc.ddd", 10)

	ss := GetSnapshot("aaa", "ccc.ddd")
	s.Equal(len(ss), 1)
	s.Equal(ss[0].Pkg(), "aaa")
	s.Equal(ss[0].Name(), "ccc.ddd")
	s.True(ss[0].HasHistogram())

	ss = GetSnapshot("*", "*")
	s.Equal(len(ss), 1)
	s.True(ss[0].HasHistogram())
}

func (s *SuiteAPI) TestNewClientWithPrefix() {
	defer s.add("aaa", "ccc.ddd", 10, 6, 0)()

	c1 := NewClient("aaa", "ccc")
	c1.BumpAvg("ddd", 10)
	c1.BumpSum("ddd", 10)

	c2 := NewClient("aaa", "ccc.")
	c2.BumpAvg("ddd", 10)
	c2.BumpSum("ddd", 10)

	c3 := NewClient("aaa", "")
	c3.BumpAvg("ccc.ddd", 10)
	c3.BumpSum("ccc.ddd", 10)

	ss := GetSnapshot("aaa", "ccc.ddd")
	s.Equal(len(ss), 1)
	s.Equal(ss[0].Pkg(), "aaa")
	s.Equal(ss[0].Name(), "ccc.ddd")
	s.False(ss[0].HasHistogram())
}

func (s *SuiteAPI) TestGetPkgs() {
	s.add("aaa", "aaa.bbb", 10, 1, 0)
	s.add("bbb", "ccc.ddd", 20, 1, 0)

	c1 := NewClient("aaa", "")
	c2 := NewClient("bbb", "")
	_ = NewClient("ccc", "")

	c1.BumpSum("aaa.bbb", 10)
	c2.BumpSum("ccc.ddd", 20)

	s.Equal(GetPkgs(false), []string{
		"aaa",
		"bbb",
	})

	s.Equal(GetPkgs(true), []string{
		"aaa",
		"bbb",
		"ccc",
	})
}

func (s *SuiteAPI) add(pkg, name string, v float64, ct, ht int) func() {
	return addTestData(s.T(), s.factory, pkg, name, v, ct, ht)
}

func TestRunSuiteAPI(t *testing.T) {
	suite.Run(t, new(SuiteAPI))
}
