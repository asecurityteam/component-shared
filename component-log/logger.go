package log

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/asecurityteam/logevent"
	"github.com/asecurityteam/settings"
)

const (
	// OutputStdout sends logs to stdout.
	OutputStdout = "STDOUT"
	// OutputNull sends logs to /dev/null.
	OutputNull    = "NULL"
	defaultLevel  = "INFO"
	defaultOutput = OutputStdout
)

// Config contains all configuration values for creating a system logger.
type Config struct {
	Level  string `description:"The minimum level of logs to emit. One of DEBUG, INFO, WARN, ERROR."`
	Output string `description:"Destination stream of the logs. One of STDOUT, NULL."`
}

// Name of the configuration as it might appear in config files.
func (*Config) Name() string {
	return "logger"
}

// Component enables creating configured loggers.
type Component struct{}

// NewComponent populates all the defaults.
func NewComponent() *Component {
	return &Component{}
}

// Settings generates a LoggerConfig with default values applied.
func (*Component) Settings() *Config {
	return &Config{
		Level:  defaultLevel,
		Output: defaultOutput,
	}
}

// New creates a configured logger instance.
func (*Component) New(_ context.Context, conf *Config) (Logger, error) {
	var output io.Writer
	switch {
	case strings.EqualFold(conf.Output, OutputStdout):
		output = os.Stdout
	case strings.EqualFold(conf.Output, OutputNull):
		output = ioutil.Discard
	default:
		return nil, fmt.Errorf("unknown logger output %s", conf.Output)
	}
	return logevent.New(logevent.Config{Level: conf.Level, Output: output}), nil
}

// Load is a convenience method for binding the source to the component.
func Load(ctx context.Context, source settings.Source, c *Component) (Logger, error) {
	dst := new(Logger)
	err := settings.NewComponent(ctx, source, c, dst)
	if err != nil {
		return nil, err
	}
	return *dst, nil
}

// New is the top-level entry point for creating a new log client.
func New(ctx context.Context, source settings.Source) (Logger, error) {
	return Load(ctx, source, NewComponent())
}
