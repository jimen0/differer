package differer

import "net/http"

// clienter allows to perform an HTTP request.
type clienter interface {
	Do(*http.Request) (*http.Response, error)
}
