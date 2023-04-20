package signals

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/asecurityteam/settings"
)

func fanIn(sigs []Signal) Signal {
	c := make(chan error, len(sigs))
	for _, sig := range sigs {
		go func(sig Signal) {
			c <- <-sig
		}(sig)
	}
	return c
}

func osSignal(sigs []os.Signal) Signal {
	c := make(chan os.Signal, len(sigs))
	exit := make(chan error, 1)
	signal.Reset(sigs...)
	signal.Notify(c, sigs...)
	go func() {
		<-c
		exit <- nil
	}()
	return exit
}

// OSConfig contains configuration for creating an OSSignal listener.
type OSConfig struct {
	Signals []int `description:"Which signals to listen for."`
}

// Name of the configuration as it might appear in a file.
func (*OSConfig) Name() string {
	return "os"
}

// Description of the configuration for help output.
func (*OSConfig) Description() string {
	return "OS signal handlers for system shutdown."
}

// OSComponent enables creation of an OS signal handler.
type OSComponent struct{}

// Settings generates a default configuration.
func (*OSComponent) Settings() *OSConfig {
	return &OSConfig{
		Signals: []int{int(syscall.SIGTERM), int(syscall.SIGINT)},
	}
}

// New generates a new Signal using OS signals.
func (*OSComponent) New(_ context.Context, conf *OSConfig) (Signal, error) {
	sigs := make([]os.Signal, 0, len(conf.Signals))
	for _, sig := range conf.Signals {
		sigs = append(sigs, syscall.Signal(sig))
	}
	return osSignal(sigs), nil
}

// Config contains all configuration for enabling various shut down signals.
type Config struct {
	Installed []string `description:"Which signal handlers are installed. Choices are OS."`
	OS        *OSConfig
}

// Name of the configuration as it appears in a file.
func (*Config) Name() string {
	return "signals"
}

// Description of the configuration for help output.
func (*Config) Description() string {
	return "Shutdown signal configuration."
}

// Component enables creation of signal handlers to shut down a system.
type Component struct {
	OS *OSComponent
}

// NewComponent populates any defaults.
func NewComponent() *Component {
	return &Component{
		OS: &OSComponent{},
	}
}

// Settings generates a default configuration.
func (c *Component) Settings() *Config {
	return &Config{
		Installed: []string{"OS"},
		OS:        c.OS.Settings(),
	}
}

// New generates a new Signal from all installed signals.
func (c *Component) New(ctx context.Context, conf *Config) (Signal, error) {
	sigs := make([]Signal, 0, len(conf.Installed))
	for _, installed := range conf.Installed {
		switch {
		case strings.EqualFold(installed, "OS"):
			sig, err := c.OS.New(ctx, conf.OS)
			if err != nil {
				return nil, err
			}
			sigs = append(sigs, sig)
		default:
			return nil, fmt.Errorf("unknown installed signal type %s", installed)
		}
	}
	return fanIn(sigs), nil
}

// Load is a convenience method for binding the source to the component.
func Load(ctx context.Context, source settings.Source, c *Component) (Signal, error) {
	dst := new(Signal)
	err := settings.NewComponent(ctx, source, c, dst)
	if err != nil {
		return nil, err
	}
	return *dst, nil
}

// New is the top-level entry point for creating a new shutdown signal.
func New(ctx context.Context, source settings.Source) (Signal, error) {
	return Load(ctx, source, NewComponent())
}
