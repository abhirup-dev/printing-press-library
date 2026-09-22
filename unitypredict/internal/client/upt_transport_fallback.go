// Copyright 2026 abhirup and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel extension (printing-press preserved file). DO NOT EDIT generated siblings.

package client

// The Chrome-uTLS transport speaks h2 over TLS only (http2.Transport with a
// custom DialTLS). Plain-http base URLs — printing-press verify's spec mock,
// local proxies, on-prem endpoints — die with "http2: unencrypted HTTP/2 not
// enabled". Dispatch on scheme instead: https keeps the Chrome fingerprint,
// http uses a standard transport.

import (
	"crypto/tls"
	"net/http"
	"time"
)

type uptSchemeTransport struct {
	secure http.RoundTripper // Chrome-uTLS h2
	plain  http.RoundTripper // standard http/1.1 (+ALPN) for http:// URLs
}

func (t uptSchemeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL != nil && req.URL.Scheme == "https" {
		return t.secure.RoundTrip(req)
	}
	return t.plain.RoundTrip(req)
}

// uptPlainTransport builds the fallback transport for plain-http base URLs.
func uptPlainTransport(skipTLSVerify bool) http.RoundTripper {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.IdleConnTimeout = 90 * time.Second
	if skipTLSVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} /* #nosec G402 -- opt-in via --insecure, parity with the generated chrome transport */
	}
	return tr
}
