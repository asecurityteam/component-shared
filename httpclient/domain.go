package httpclient

import (
	"net/http"
)

// RoundTripper is the interface that handles all HTTP operations. It is almost
// exclusively used with an http.Client wrapped around it. This is included here
// for documentation purposes only.
type RoundTripper = http.RoundTripper
