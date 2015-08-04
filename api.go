package metric

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/stats"
)

var (
	// pkgClis stores pkgClients
	pkgClis = map[string]*pkgClient{}
	// pkgClis protects pkgClis
	pkgClisLock = sync.RWMutex{}
	// default counter paramters
	counterParams = struct {
		window, bucket time.Duration
	}{window: 15 * time.Minute, bucket: time.Minute}
	// default histogram parameters
	histogramParams = struct {
		window, bucket time.Duration
	}{window: 5 * time.Minute, bucket: time.Minute}
)

// Counter defines interface for counter
type Counter interface {
	// Incr increments value to the counter
	Incr(value float64)
	// Snapshot return the snapshot of the counter
	Snapshot() CounterSnapshot
}

// Histogram defines interface for counter
type Histogram interface {
	// Update add value to the histogram
	Update(value float64)
	// Snapshot returns the snapshot of the histogrom
	Snapshot() HistSnapshot
}

// Snapshot includes CounterSnapshot, HistogramSnapshot and name
type Snapshot interface {
	CounterSnapshot
	HistSnapshot
	// HasHistogram returns whether this snapshot contains histogram
	HasHistogram() bool
	// Pkg returns the package of Histogram
	Pkg() string
	// Name returns the package of Histogram Name
	Name() string
}

// CounterSnapshot defines interface for accessing counter with following methods
type CounterSnapshot interface {
	// SliceIn returns statistics of each bucket in the given duration
	SliceIn(dur time.Duration) []Bucket
	// AggrIn returns aggregration statistics in the given duration
	AggrIn(dur time.Duration) Bucket
}

// HistSnapshot represents a snapshot of a histogram
type HistSnapshot interface {
	// List histogram bins
	Bins() []Bin
	// Return export percentail values of the histogram
	Percentiles([]float64) ([]float64, int64)
}

// Bucket represents the snapshot of a counter bucket, including
// statistics values like count, sum, average, min, max and bucket start/end
// time.
type Bucket struct {
	Count float64
	Sum   float64
	Min   float64
	Max   float64
	Avg   float64
	Start time.Time
	End   time.Time
}

// Bin represents the snapshot of a histogram bin, including bin couner and
// its lower and upper bound
type Bin struct {
	Count int64
	Lower float64
	Upper float64
}

// NewClient creates an instance of facebookgo/stats implementation with
// the given pkg name and preifx.
func NewClient(pkg, prefix string) stats.Client {
	pkgClisLock.Lock()
	defer pkgClisLock.Unlock()

	pc, ok := pkgClis[pkg]
	if !ok {
		pc = newClient(pkg)
		pkgClis[pkg] = pc
	}
	if prefix == "" {
		return pc
	}
	if !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}
	return stats.PrefixClient([]string{prefix}, pc)
}

// GetSnapshot returns shapshot of counters and histograms matched the
// given pkg and name
func GetSnapshot(qpkg string, qname string) []Snapshot {
	pkgClisLock.RLock()
	defer pkgClisLock.RUnlock()

	snapshots := []Snapshot{}
	for _, pc := range pkgClis {
		if qpkg != "*" && !strings.Contains(pc.pkg, qpkg) {
			continue
		}
		snapshots = append(snapshots, pc.get(qname)...)
	}
	return snapshots
}

// GetPkgs lista all package names
func GetPkgs(showEmpty bool) []string {
	pkgClisLock.RLock()
	result := make([]string, 0, len(pkgClis))
	for pkg, cli := range pkgClis {
		if showEmpty || cli.size() > 0 {
			result = append(result, pkg)
		}
	}
	pkgClisLock.RUnlock()

	sort.Strings(result)
	return result
}

// SetCounterParam sets the parameters of counter and histogram
func SetCounterParam(window, bucket time.Duration) error {
	if err := check(window, bucket); err != nil {
		return err
	}
	counterParams.window = window
	counterParams.bucket = bucket
	return nil
}

// SetHistogramParam sets the parameters of counter and histogram
func SetHistogramParam(window, bucket time.Duration) error {
	if err := check(window, bucket); err != nil {
		return err
	}
	histogramParams.window = window
	histogramParams.bucket = bucket
	return nil
}
