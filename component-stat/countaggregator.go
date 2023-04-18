package stat

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/xstats"
)

// CountAggregator is a wrapper around xstats.XStater, that aggregates
// Count metrics on a time interval before sending them through Stater
type CountAggregator struct {
	Stater        xstats.XStater
	Bucket        map[StatTagKey]float64
	lock          *sync.Mutex
	flushJob      chan int
	FlushInterval time.Duration
}

// StatTagKey is a struct that represents a composite key
// for storing aggregated stats
type StatTagKey struct {
	StatKey string
	TagsKey string
}

// insertStat inserts an incoming stat to Bucket, the container of aggregated data.
func (ca *CountAggregator) insertStat(stat string, count float64, tags ...string) {
	// sort tags, as we use a concatenation of strings as a key
	sort.Strings(tags)
	tagsKey := strings.Join(tags[:], " ")
	compositeKey := StatTagKey{
		StatKey: stat,
		TagsKey: tagsKey,
	}
	// provide mutual exclusion around the Bucket
	ca.lock.Lock()
	defer ca.lock.Unlock()
	if currentCount, ok := ca.Bucket[compositeKey]; ok {
		count = currentCount + count
	}
	ca.Bucket[compositeKey] = count
}

// clearBucket wipes the Bucket of all aggregated data, to be used on flush
func (ca *CountAggregator) clearBucket() {
	for key := range ca.Bucket {
		delete(ca.Bucket, key)
	}
}

// retrieveAndConvertStat takes in a composite key, retrieves the associated aggregated stat, and
// converts it back to a consumable stat for xstats.XStater
func (ca *CountAggregator) retrieveAndConvertStat(compositeKey StatTagKey) (string, float64, []string) {
	aggregateCount := ca.Bucket[compositeKey]
	tags := strings.Split(compositeKey.TagsKey, " ")
	stat := compositeKey.StatKey
	return stat, aggregateCount, tags
}

// flush calls xstats.XStater.Count on all aggregated stats, and proceeds to wipe the Bucket
func (ca *CountAggregator) flush() {
	time.Sleep(ca.FlushInterval)
	ca.lock.Lock()
	defer ca.lock.Unlock()
	for compositeKey := range ca.Bucket {
		aggregatedStat, aggregatedCount, aggregatedTags := ca.retrieveAndConvertStat(compositeKey)
		ca.Stater.Count(aggregatedStat, aggregatedCount, aggregatedTags...)
	}
	ca.clearBucket()
	<-ca.flushJob
}

// Count implements XStater interface. This Count in particular
// inserts a stat, then proceeds to try to flush the Bucket. If there exists a
// flush in progress, we proceed to overflow on the channel and return by default
func (ca *CountAggregator) Count(stat string, count float64, tags ...string) {
	ca.insertStat(stat, count, tags...)

	select {
	case ca.flushJob <- 1:
		go ca.flush()
	default:
		return
	}
}

// Gauge implements XStater interface
func (ca *CountAggregator) Gauge(stat string, value float64, tags ...string) {
	ca.Stater.Gauge(stat, value, tags...)
}

// Histogram implements XStater interface
func (ca *CountAggregator) Histogram(stat string, value float64, tags ...string) {
	ca.Stater.Histogram(stat, value, tags...)
}

// Timing implements XStater interface
func (ca *CountAggregator) Timing(stat string, duration time.Duration, tags ...string) {
	ca.Stater.Timing(stat, duration, tags...)
}

// AddTags implements XStater interface
func (ca *CountAggregator) AddTags(tags ...string) {
	ca.Stater.AddTags(tags...)
}

// GetTags implements XStater interface
func (ca *CountAggregator) GetTags() []string {
	return ca.Stater.GetTags()
}
