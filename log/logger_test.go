package log

import (
	"context"
	"testing"

	"github.com/asecurityteam/settings"
	"github.com/stretchr/testify/require"
)

func TestComponent(t *testing.T) {
	src := settings.NewMapSource(map[string]interface{}{
		"logger": map[string]interface{}{
			"output": "NULL",
		},
	})
	tr, err := New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)

	src = settings.NewMapSource(map[string]interface{}{
		"logger": map[string]interface{}{
			"output": "STDOUT",
		},
	})
	tr, err = New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, tr)

	src = settings.NewMapSource(map[string]interface{}{
		"logger": map[string]interface{}{
			"output": "MISSING",
		},
	})
	_, err = New(context.Background(), src)
	require.NotNil(t, err)
}
