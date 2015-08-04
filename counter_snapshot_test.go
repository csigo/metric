package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TestCase counter
type SuiteCounterSnapshot struct {
	suite.Suite
	sh  *counterSnapshot
	now *time.Time
}

func (s *SuiteCounterSnapshot) SetupTest() {
	s.sh = &counterSnapshot{
		bucketDur: defaultBucket,
		buckets: []bucket{
			bucket{
				end:   timeNow(),
				count: 10,
				sum:   20,
				min:   2,
				max:   8,
			},
			bucket{
				end:   timeNow() - int64(defaultBucket),
				count: 20,
				sum:   70,
				min:   7,
				max:   12,
			},
			bucket{
				end:   timeNow() - 2*int64(defaultBucket),
				count: 30,
				sum:   70,
				min:   6,
				max:   11,
			},
		},
	}
}

func (s *SuiteCounterSnapshot) TestSliceIn() {
	now := time.Unix(0, timeNow())
	slice := s.sh.SliceIn(defaultBucket)
	s.Equal([]Bucket{
		Bucket{
			Count: 10,
			Sum:   20,
			Min:   2,
			Max:   8,
			Avg:   2,
			Start: now.Add(-defaultBucket),
			End:   now,
		},
		Bucket{
			Count: 20,
			Sum:   70,
			Min:   7,
			Max:   12,
			Avg:   3.5,
			Start: now.Add(-2 * defaultBucket),
			End:   now.Add(-defaultBucket),
		},
	}, slice)
}

func (s *SuiteCounterSnapshot) TestAvgIn() {
	now := time.Unix(0, timeNow())
	slice := s.sh.AggrIn(defaultBucket)
	s.Equal(Bucket{
		Count: 30,
		Sum:   90,
		Min:   2,
		Max:   12,
		Avg:   3,
		Start: now.Add(-2 * defaultBucket),
		End:   now,
	}, slice)
}

func TestRunSuiteCounterSnapshot(t *testing.T) {
	suite.Run(t, new(SuiteCounterSnapshot))
}
