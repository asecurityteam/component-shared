package log

import (
	"context"

	"github.com/asecurityteam/logevent"
)

// Logger is the project logging client interface. It is
// currently an alias to the logevent project.
type Logger = logevent.Logger

// LogFn is the type that should be accepted by components that
// intend to log content using the context logger.
type LogFn func(context.Context) Logger

// LoggerFromContext is the concrete implementation of LogFn
// that should be used at runtime.
var LoggerFromContext = logevent.FromContext
