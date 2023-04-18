package stat

import (
	"context"

	"github.com/rs/xstats"
)

// Stat is the project metrics client interface. it is currently
// an alias for xstats.XStater.
type Stat = xstats.XStater

// StatFn is the type that should be accepted by code that
// intend to emit custom metrics using the context stat client.
type StatFn func(context.Context) Stat

// StatFromContext is the concrete implementation of StatFn that
// should be used at runtime.
var StatFromContext = xstats.FromContext
