package stat

import (
	"context"
	"sync"
	"time"

	"github.com/rs/xstats"
)

// CountAggregatorConfig is the configuration for a CountAggregator.
type CountAggregatorConfig struct {
	FlushInterval time.Duration `description:"Frequency of when to send aggregated metrics."`
	StatConfig    *DatadogConfig
}

// Name of the configuration as it might appear in config files.
func (*CountAggregatorConfig) Name() string {
	return "statcountaggregator"
}

// CountAggregatorComponent implements the settings.Component interface for
// a countaggregator.
type CountAggregatorComponent struct {
	StatComponent *DatadogComponent
}

// Settings generates a config with default values applied.
func (c *CountAggregatorComponent) Settings() *CountAggregatorConfig {
	return &CountAggregatorConfig{
		FlushInterval: 10 * time.Second,
		StatConfig:    c.StatComponent.Settings(),
	}
}

// New creates a configured countaggregator that wraps around a stats client.
func (c *CountAggregatorComponent) New(ctx context.Context, conf *CountAggregatorConfig) (xstats.XStater, error) {

	stater, err := c.StatComponent.New(ctx, conf.StatConfig)
	if err != nil {
		return nil, err
	}
	countAggregator := CountAggregator{
		Stater:        stater,
		Bucket:        make(map[StatTagKey]float64),
		lock:          &sync.Mutex{},
		flushJob:      make(chan int, 1),
		FlushInterval: conf.FlushInterval,
	}

	return &countAggregator, nil
}
