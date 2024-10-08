package http

import (
	"net/http"

	"github.com/picatz/simnet"
)

// Transport is an http.RoundTripper that simulates network conditions.
type Transport struct {
	Underlying *http.Transport // Underlying transport (optional)
	Dialer     *simnet.Dialer  // Simulated Dialer
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Underlying
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		transport = transport.Clone()
	}

	transport.DialContext = t.Dialer.DialContext

	return transport.RoundTrip(req)
}
