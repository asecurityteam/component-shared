package signals

import (
	"context"
	"syscall"
	"testing"
	"time"

	// "github.com/asecurityteam/settings"
	"github.com/asecurityteam/settings"
	"github.com/stretchr/testify/require"
)

func TestOSComponent(t *testing.T) {
	cmp := &OSComponent{}
	conf := cmp.Settings()
	sig, err := cmp.New(context.Background(), conf)
	require.Nil(t, err)
	require.NotNil(t, sig)
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	select {
	case <-sig:
	case <-time.After(time.Second):
		require.Fail(t, "did not receive signal")
	}
}

func TestComponent(t *testing.T) {
	src := settings.NewMapSource(map[string]interface{}{
		"signals": map[string]interface{}{
			"installed": "",
		},
	})
	sig, err := New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, sig)

	src = settings.NewMapSource(map[string]interface{}{
		"signals": map[string]interface{}{
			"installed": "OS",
			"os": map[string]interface{}{
				"signals": []string{"15", "2"},
			},
		},
	})
	sig, err = New(context.Background(), src)
	require.Nil(t, err)
	require.NotNil(t, sig)
	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	select {
	case <-sig:
	case <-time.After(time.Second):
		require.Fail(t, "did not receive signal")
	}
}
