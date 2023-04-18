package stat

import (
	"context"
	"testing"
	"time"

	"github.com/asecurityteam/settings"
	"github.com/stretchr/testify/require"
)

func TestNullComponent(t *testing.T) {
	cmp := &NullComponent{}
	conf := cmp.Settings()
	tr, err := cmp.New(context.Background(), conf)
	require.Nil(t, err)
	require.NotNil(t, tr)
}

func TestDatadogComponentBadConfig(t *testing.T) {
	cmp := &DatadogComponent{}
	conf := cmp.Settings()
	conf.Address = ""
	_, err := cmp.New(context.Background(), conf)
	require.NotNil(t, err)
}

func TestDatadogComponent(t *testing.T) {
	cmp := &DatadogComponent{}
	conf := cmp.Settings()
	tr, err := cmp.New(context.Background(), conf)
	require.Nil(t, err)
	require.NotNil(t, tr)
}

func TestComponent(t *testing.T) {
	src := settings.NewMapSource(map[string]interface{}{
		"stats": map[string]interface{}{
			"output": "NULL",
		},
	})
	tr, err := New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)

	src = settings.NewMapSource(map[string]interface{}{
		"stats": map[string]interface{}{
			"output": "DATADOG",
			"datadog": map[string]interface{}{
				"flushinterval": 10 * time.Second,
			},
		},
	})
	tr, err = New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)

	src = settings.NewMapSource(map[string]interface{}{
		"stats": map[string]interface{}{
			"output": "MISSING",
		},
	})
	_, err = New(context.Background(), src)
	require.NotNil(t, err)
}
