package metric

import (
	"strings"
	"sync"
)

var (
	// making functions as variable for testing
	newCounter   = NewCounter
	newHistogram = NewHistogram
)

// newClient creates an instance of facebookgo/stats implementation
// with the given pkg
func newClient(pkg string) *pkgClient {
	return &pkgClient{
		pkg:   pkg,
		pairs: map[string]*pair{},
	}
}

// pkgClient implements interface of facebookgo/stats
// It maintains counter and histogram instances of a certain package
// Counters and histograms are created on demain
// It is safe for concurrent use by multiple goroutines.
type pkgClient struct {
	sync.RWMutex
	pkg   string
	pairs map[string]*pair
}

// pair contains counter and histogrm of the same name
type pair struct {
	counter Counter
	hist    Histogram
}

// endable is for BumpTime return values
type endable struct {
	end func()
}

func (e *endable) End() {
	e.end()
}

// BumpAvg implements interface of facebookgo/stats
func (p *pkgClient) BumpAvg(key string, val float64) {
	p.BumpSum(key, val)
}

// BumpSum implements interface of facebookgo/stats
func (p *pkgClient) BumpSum(key string, val float64) {
	c, _ := p.ensure(key, false)
	c.Incr(val)
}

// BumpTime implements interface of facebookgo/stats
func (p *pkgClient) BumpTime(key string) interface {
	End()
} {
	start := timeNow()
	return &endable{
		end: func() {
			end := timeNow()
			p.BumpHistogram(key, float64(end-start))
		},
	}
}

// BumpHistogram implements interface of facebookgo/stats
func (p *pkgClient) BumpHistogram(key string, val float64) {
	c, h := p.ensure(key, true)
	c.Incr(val)
	h.Update(val)
}

// size returns the number of counters
func (p *pkgClient) size() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.pairs)
}

// get returns counter and histogram shapshots matched the given qname.
// if qname is "*", client will return all counters and histograms
func (p *pkgClient) get(qname string) []Snapshot {
	p.RLock()
	defer p.RUnlock()
	snapshots := make([]Snapshot, 0, len(p.pairs))
	for name, r := range p.pairs {
		if qname != "*" && !strings.Contains(name, qname) {
			continue
		}
		c := r.counter.Snapshot()
		h := HistSnapshot(nil)
		if r.hist != nil {
			h = r.hist.Snapshot()
		}
		snapshots = append(snapshots, &snapshot{
			pkg:             p.pkg,
			name:            name,
			CounterSnapshot: c,
			HistSnapshot:    h,
		})
	}
	return snapshots
}

// ensure returns the keeped counter and histogram of the given name, and creates them
// if they are not in the paris map
func (p *pkgClient) ensure(name string, hist bool) (Counter, Histogram) {
	p.RLock()
	r, ok := p.pairs[name]
	p.RUnlock()
	if ok && (!hist || r.hist != nil) {
		return r.counter, r.hist
	}
	// modify pair
	p.Lock()
	defer p.Unlock()
	// need to check again
	r, ok = p.pairs[name]
	if ok && (!hist || r.hist != nil) {
		return r.counter, r.hist
	}
	if !ok {
		ctr, _ := newCounter(counterParams.window, counterParams.bucket)
		r = &pair{counter: ctr}
	}
	// create histogram if needed
	if hist {
		r.hist, _ = newHistogram(histogramParams.window, histogramParams.bucket)
	}
	p.pairs[name] = r
	return r.counter, r.hist
}

// snapshot implements Snapshot interface
type snapshot struct {
	pkg  string
	name string
	CounterSnapshot
	HistSnapshot
}

func (s *snapshot) Pkg() string {
	return s.pkg
}

func (s *snapshot) Name() string {
	return s.name
}

func (s *snapshot) HasHistogram() bool {
	return s.HistSnapshot != nil
}
