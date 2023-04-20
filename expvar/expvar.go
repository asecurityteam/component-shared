package expvar

import (
	"context"
	"runtime"
	"sync"
	"time"

	stat "github.com/asecurityteam/component-shared/stat"
	"github.com/asecurityteam/settings"
)

const (
	statMemstatsAlloc        = "go_expvar.memstats.alloc"
	statMemstatsFrees        = "go_expvar.memstats.frees"
	statMemstatsHeapAlloc    = "go_expvar.memstats.heap_alloc"
	statMemstatsHeapIdle     = "go_expvar.memstats.heap_idle"
	statMemstatsHeapInuse    = "go_expvar.memstats.heap_inuse"
	statMemstatsHeapObjects  = "go_expvar.memstats.heap_objects"
	statMemstatsHeapReleased = "go_expvar.memstats.heap_released"
	statMemstatsHeapSys      = "go_expvar.memstats.heap_sys"
	statMemstatsLookups      = "go_expvar.memstats.lookups"
	statMemstatsMallocs      = "go_expvar.memstats.mallocs"
	statMemstatsNumGC        = "go_expvar.memstats.num_gc"
	statMemstatsPauseNS      = "go_expvar.memstats.pause_ns"
	statMemstatsPauseTotalNS = "go_expvar.memstats.pause_total_ns"
	statMemstatsTotalAlloc   = "go_expvar.memstats.total_alloc"
	statGoroutinesExists     = "go_expvar.goroutines.exists"
	expvarInterval           = 5 * time.Second
)

// Expvar tracks the memory usage of the Go runtime and collects metrics instrumented from Go’s expvar package
type Expvar struct {
	Stat                     stat.Stat
	MemstatsAllocName        string
	MemstatsFreesName        string
	MemstatsHeapAllocName    string
	MemstatsHeapIdleName     string
	MemstatsHeapInuseName    string
	MemstatsHeapObjectsName  string
	MemstatsHeapReleasedName string
	MemstatsHeapSysName      string
	MemstatsLookupsName      string
	MemstatsMallocsName      string
	MemstatsNumGCName        string
	MemstatsPauseNSName      string
	MemstatsPauseTotalNSName string
	MemstatsTotalAllocName   string
	GoroutinesExistsName     string
	Interval                 time.Duration
	reportMut                *sync.Mutex
	stop                     chan interface{}
	lastNumGC                uint32
	readMemStats             func(*runtime.MemStats)
	numGoroutine             func() int
}

// Report loops on a time interval and pushes a set of gauge metrics.
func (e *Expvar) Report() {
	ticker := time.NewTicker(e.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.report()
		case <-e.stop:
			return
		}
	}
}

// Close the reporting loop.
func (e *Expvar) Close() error {
	close(e.stop)
	return nil
}

func (e *Expvar) report() {
	memstats := new(runtime.MemStats)
	e.readMemStats(memstats)
	numGoroutines := e.numGoroutine()

	e.reportMut.Lock()
	defer e.reportMut.Unlock()

	e.Stat.Gauge(e.MemstatsAllocName, float64(memstats.Alloc))
	e.Stat.Gauge(e.MemstatsFreesName, float64(memstats.Frees))
	e.Stat.Gauge(e.MemstatsHeapAllocName, float64(memstats.HeapAlloc))
	e.Stat.Gauge(e.MemstatsHeapIdleName, float64(memstats.HeapIdle))
	e.Stat.Gauge(e.MemstatsHeapInuseName, float64(memstats.HeapInuse))
	e.Stat.Gauge(e.MemstatsHeapObjectsName, float64(memstats.HeapObjects))
	e.Stat.Gauge(e.MemstatsHeapReleasedName, float64(memstats.HeapReleased))
	e.Stat.Gauge(e.MemstatsHeapSysName, float64(memstats.HeapSys))
	e.Stat.Gauge(e.MemstatsLookupsName, float64(memstats.Lookups))
	e.Stat.Gauge(e.MemstatsMallocsName, float64(memstats.Mallocs))
	e.Stat.Gauge(e.MemstatsNumGCName, float64(memstats.NumGC))
	e.Stat.Gauge(e.MemstatsPauseTotalNSName, float64(memstats.PauseTotalNs))
	e.Stat.Gauge(e.MemstatsTotalAllocName, float64(memstats.TotalAlloc))
	e.Stat.Gauge(e.GoroutinesExistsName, float64(numGoroutines))

	// PauseNs is a circular buffer of recent GC stop-the-world
	// pause times in nanoseconds.
	//
	// The most recent pause is at PauseNs[(NumGC+255)%256]. In
	// general, PauseNs[N%256] records the time paused in the most
	// recent N%256th GC cycle. There may be multiple pauses per
	// GC cycle; this is the sum of all pauses during a cycle.
	pauseNS := memstats.PauseNs
	if e.lastNumGC == memstats.NumGC { // no GC cycles have run
		return
	}
	start := e.lastNumGC % 256
	end := (memstats.NumGC+255)%256 + 1
	var values []uint64
	if start < end {
		values = pauseNS[start:end]
	} else {
		values = append(pauseNS[start:], pauseNS[:end]...)
	}

	e.lastNumGC = memstats.NumGC

	for _, val := range values {
		e.Stat.Histogram(e.MemstatsPauseNSName, float64(val))
	}
}

// Config is a container for internal expvar metrics settings.
type Config struct {
	Alloc            string        `description:"Name of the metric tracking allocated bytes"`
	Frees            string        `description:"Name of the metric tracking number of frees"`
	HeapAlloc        string        `description:"Name of the metric tracking allocated bytes"`
	HeapIdle         string        `description:"Name of the metric tracking bytes in unused spans"`
	HeapInuse        string        `description:"Name of the metric tracking bytes in in-use spans"`
	HeapObjects      string        `description:"Name of the metric tracking total number of object allocated"`
	HeapReleased     string        `description:"Name of the metric tracking bytes realeased to the OS"`
	HeapSys          string        `description:"Name of the metric tracking bytes obtained from the system"`
	Lookups          string        `description:"Name of the metric tracking number of pointer lookups"`
	Mallocs          string        `description:"Name of the metric tracking number of mallocs"`
	NumGC            string        `description:"Name of the metric tracking number of garbage collections"`
	PauseNS          string        `description:"Name of the metric tracking duration of GC pauses"`
	PauseTotalNS     string        `description:"Name of the metric tracking total GC pause duration over lifetime process"`
	TotalAlloc       string        `description:"Name of the metric tracking allocated bytes (even if freed)"`
	GoroutinesExists string        `description:"Name of the metric tracking number of active go routines"`
	ReportInterval   time.Duration `description:"Interval on which metrics are reported."`
}

// Name of the configuration root.
func (*Config) Name() string {
	return "expvar"
}

// Description returns the help information for the configuration root.
func (*Config) Description() string {
	return "Expvar metric names"
}

// Component implements the settings.Component interface for expvar metrics.
type Component struct {
	Stat stat.Stat
}

// NewComponent generates a blank component instance.
func NewComponent() *Component {
	return &Component{}
}

// WithStat returns a copy of the component bound to the given Stat instance.
func (c *Component) WithStat(s stat.Stat) *Component {
	return &Component{Stat: s}
}

// Settings returns a configuration with all defaults set.
func (*Component) Settings() *Config {
	return &Config{
		Alloc:            statMemstatsAlloc,
		Frees:            statMemstatsFrees,
		HeapAlloc:        statMemstatsHeapAlloc,
		HeapIdle:         statMemstatsHeapIdle,
		HeapInuse:        statMemstatsHeapInuse,
		HeapObjects:      statMemstatsHeapObjects,
		HeapReleased:     statMemstatsHeapReleased,
		HeapSys:          statMemstatsHeapSys,
		Lookups:          statMemstatsLookups,
		Mallocs:          statMemstatsMallocs,
		NumGC:            statMemstatsNumGC,
		PauseNS:          statMemstatsPauseNS,
		PauseTotalNS:     statMemstatsPauseTotalNS,
		TotalAlloc:       statMemstatsTotalAlloc,
		GoroutinesExists: statGoroutinesExists,
		ReportInterval:   expvarInterval,
	}
}

// New produces a ServerFn bound to the given configuration.
func (c *Component) New(_ context.Context, conf *Config) (*Expvar, error) {
	return &Expvar{
		Stat:                     c.Stat,
		MemstatsAllocName:        conf.Alloc,
		MemstatsFreesName:        conf.Frees,
		MemstatsHeapAllocName:    conf.HeapAlloc,
		MemstatsHeapIdleName:     conf.HeapIdle,
		MemstatsHeapInuseName:    conf.HeapInuse,
		MemstatsHeapObjectsName:  conf.HeapObjects,
		MemstatsHeapReleasedName: conf.HeapReleased,
		MemstatsHeapSysName:      conf.HeapSys,
		MemstatsLookupsName:      conf.Lookups,
		MemstatsMallocsName:      conf.Mallocs,
		MemstatsNumGCName:        conf.NumGC,
		MemstatsPauseNSName:      conf.PauseNS,
		MemstatsPauseTotalNSName: conf.PauseTotalNS,
		MemstatsTotalAllocName:   conf.TotalAlloc,
		GoroutinesExistsName:     conf.GoroutinesExists,
		Interval:                 conf.ReportInterval,
		reportMut:                &sync.Mutex{},
		stop:                     make(chan interface{}),
		readMemStats:             runtime.ReadMemStats,
		numGoroutine:             runtime.NumGoroutine,
	}, nil

}

// Load is a convenience method for binding the source to the component.
func Load(ctx context.Context, source settings.Source, c *Component) (*Expvar, error) {
	dst := new(Expvar)
	err := settings.NewComponent(ctx, source, c, dst)
	if err != nil {
		return nil, err
	}
	return dst, nil
}
