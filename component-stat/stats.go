package stat

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/asecurityteam/settings"
	"github.com/rs/xstats"
	"github.com/rs/xstats/dogstatsd"
)

const (
	defaultDDAddr       = "localhost:8125"
	defaultDDPacketSize = 1 << 15 // Matches what is used in the xstats library.
	defaultDDFlush      = 10 * time.Second
	// OutputNull is the selection for no metrics.
	OutputNull = "NULL"
	// OutputDatadog selects the datadog/extended-statsd driver.
	OutputDatadog = "DATADOG"
	defaultOutput = OutputNull
)

var (
	defaultDDTags = []string{}
)

// NullConfig is empty. There are no options for NULL.
type NullConfig struct{}

// Name of the configuration as it might appear in config files.
func (*NullConfig) Name() string {
	return "nullstat"
}

// NullComponent implements the settings.Component interface for
// a NOP stat client.
type NullComponent struct{}

// Settings generates a config with default values applied.
func (*NullComponent) Settings() *NullConfig {
	return &NullConfig{}
}

// New creates a configured stats client.
func (*NullComponent) New(_ context.Context, conf *NullConfig) (Stat, error) {
	return xstats.FromContext(context.Background()), nil
}

// DatadogConfig is for configuration a datadog client.
type DatadogConfig struct {
	Address       string        `description:"Listener address to use when sending metrics."`
	FlushInterval time.Duration `description:"Frequencing of sending metrics to listener."`
	Tags          []string      `description:"Any static tags for all metrics."`
	PacketSize    int           `description:"Max packet size to send."`
}

// Name of the configuration as it might appear in config files.
func (*DatadogConfig) Name() string {
	return "datadog"
}

// DatadogComponent implements the settings.Component interface for
// a datadog stats client.
type DatadogComponent struct{}

// Settings generates a config with default values applied.
func (*DatadogComponent) Settings() *DatadogConfig {
	return &DatadogConfig{
		Address:       defaultDDAddr,
		FlushInterval: defaultDDFlush,
		Tags:          defaultDDTags,
		PacketSize:    defaultDDPacketSize,
	}
}

// New creates a configured stats client.
func (*DatadogComponent) New(_ context.Context, conf *DatadogConfig) (Stat, error) {
	writer, err := net.Dial("udp", conf.Address)
	if err != nil {
		return nil, err
	}
	stater := xstats.New(dogstatsd.NewMaxPacket(writer, conf.FlushInterval, conf.PacketSize))
	if len(conf.Tags) > 0 {
		stater.AddTags(conf.Tags...)
	}
	return stater, nil
}

// Config contains all configuration values for creating
// a system stat client.
type Config struct {
	Output   string `description:"Destination stream of the stats. One of NULLSTAT, DATADOG."`
	NullStat *NullConfig
	Datadog  *DatadogConfig
}

// Name of the configuration as it might appear in config files.
func (*Config) Name() string {
	return "stats"
}

// Component enables creating configured loggers.
type Component struct {
	NullStat *NullComponent
	Datadog  *DatadogComponent
}

// NewComponent populates an StatComponent with defaults.
func NewComponent() *Component {
	return &Component{
		NullStat: &NullComponent{},
		Datadog:  &DatadogComponent{},
	}
}

// Settings generates a StatsConfig with default values applied.
func (c *Component) Settings() *Config {
	return &Config{
		Output:   defaultOutput,
		NullStat: c.NullStat.Settings(),
		Datadog:  c.Datadog.Settings(),
	}
}

// New creates a configured stats client.
func (c *Component) New(ctx context.Context, conf *Config) (Stat, error) {
	switch {
	case strings.EqualFold(conf.Output, OutputNull):
		return c.NullStat.New(ctx, conf.NullStat)
	case strings.EqualFold(conf.Output, OutputDatadog):
		return c.Datadog.New(ctx, conf.Datadog)
	default:
		return nil, fmt.Errorf("unknown stats output %s", conf.Output)
	}
}

// Load is a convenience method for binding the source to the component.
func Load(ctx context.Context, source settings.Source, c *Component) (Stat, error) {
	dst := new(Stat)
	err := settings.NewComponent(ctx, source, c, dst)
	if err != nil {
		return nil, err
	}
	return *dst, nil
}

// New is the top-level entry point for creating a new stat client.
func New(ctx context.Context, source settings.Source) (Stat, error) {
	return Load(ctx, source, NewComponent())
}
