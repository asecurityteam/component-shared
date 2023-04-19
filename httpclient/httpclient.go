package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/asecurityteam/transport"

	"github.com/asecurityteam/settings"
	transportd "github.com/asecurityteam/transportd/pkg"
	componentsd "github.com/asecurityteam/transportd/pkg/components"
)

const (
	// TypeDefault is used to select the default Go HTTP client.
	TypeDefault = "DEFAULT"
	// TypeSmart is used to select the transportd HTTP client.
	TypeSmart = "SMART"
)

// DefaultConfig contains all settings for the default Go HTTP client.
type DefaultConfig struct {
	ContentType string
}

// DefaultComponent is a component for creating the default Go HTTP client.
type DefaultComponent struct{}

// Settings returns the default configuration.
func (*DefaultComponent) Settings() *DefaultConfig {
	return &DefaultConfig{
		ContentType: "application/json",
	}
}

// New constructs a client from the given configuration
func (*DefaultComponent) New(ctx context.Context, conf *DefaultConfig) (http.RoundTripper, error) {
	return transport.NewHeader(
		func(*http.Request) (string, string) {
			return "Content-Type", conf.ContentType
		},
	)(http.DefaultTransport), nil
}

// SmartConfig contains all settings for the transportd HTTP client.
type SmartConfig struct {
	OpenAPI string `description:"The full OpenAPI specification with transportd extensions."`
}

// Name of the configuration tree.
func (*SmartConfig) Name() string {
	return "smart"
}

// SmartComponent is a component for creating a transportd HTTP client.
type SmartComponent struct {
	Plugins []transportd.NewComponent
}

// Settings returns the default configuration.
func (*SmartComponent) Settings() *SmartConfig {
	return &SmartConfig{}
}

// New constructs a client from the given configuration.
func (c *SmartComponent) New(ctx context.Context, conf *SmartConfig) (http.RoundTripper, error) {
	return transportd.NewTransport(ctx, []byte(conf.OpenAPI), c.Plugins...)
}

// Config wraps all HTTP related settings.
type Config struct {
	Type    string `description:"The type of HTTP client. Choices are SMART and DEFAULT."`
	Default *DefaultConfig
	Smart   *SmartConfig
}

// Name of the config.
func (*Config) Name() string {
	return "httpclient"
}

// Component is the top level HTTP client component.
type Component struct {
	Default *DefaultComponent
	Smart   *SmartComponent
}

// NewComponent populates an HTTPComponent with defaults.
func NewComponent() *Component {
	return &Component{
		Default: &DefaultComponent{},
		Smart: &SmartComponent{
			Plugins: componentsd.Defaults,
		},
	}
}

// Settings returns the default configuration.
func (c *Component) Settings() *Config {
	return &Config{
		Type:    "DEFAULT",
		Default: c.Default.Settings(),
		Smart:   c.Smart.Settings(),
	}
}

// New constructs a client from the given configuration.
func (c *Component) New(ctx context.Context, conf *Config) (http.RoundTripper, error) {
	switch {
	case strings.EqualFold(conf.Type, TypeDefault):
		return c.Default.New(ctx, conf.Default)
	case strings.EqualFold(conf.Type, TypeSmart):
		return c.Smart.New(ctx, conf.Smart)
	default:
		return nil, fmt.Errorf("unknown HTTP client type %s", conf.Type)
	}
}

// Load is a convenience method for binding the source to the component.
func Load(ctx context.Context, source settings.Source, c *Component) (http.RoundTripper, error) {
	dst := new(http.RoundTripper)
	err := settings.NewComponent(ctx, source, c, dst)
	if err != nil {
		return nil, err
	}
	return *dst, nil
}

// New is the top-level entry point for creating a new HTTP client.
// The default set of plugins will be installed for the smart client. Use the
// LoadHTTP() method if a custom set of plugins are required.
func New(ctx context.Context, source settings.Source) (http.RoundTripper, error) {
	return Load(ctx, source, NewComponent())
}
