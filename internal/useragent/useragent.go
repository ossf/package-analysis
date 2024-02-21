package useragent

import (
	"fmt"
	"net/http"
)

const defaultUserAgentFmt = "package-analysis (github.com/ossf/package-analysis%s)"

type uaRoundTripper struct {
	parent    http.RoundTripper
	userAgent string
}

// RoundTrip implements the http.RoundTripper interface.
func (rt *uaRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", rt.userAgent)
	return rt.parent.RoundTrip(req)
}

// RoundTripper wraps parent with a RoundTripper that add a user-agent header
// with the contents of ua.
func RoundTripper(ua string, parent http.RoundTripper) http.RoundTripper {
	return &uaRoundTripper{
		parent:    parent,
		userAgent: ua,
	}
}

// DefaultRoundTripper wraps parent with a RoundTripper that adds a default
// Package Analysis user-agent header.
//
// If supplied, extra information can be added to the user-agent, allowing the
// user-agent to be customized for production environments.
func DefaultRoundTripper(parent http.RoundTripper, extra string) http.RoundTripper {
	if extra != "" {
		extra = ", " + extra
	}
	return RoundTripper(fmt.Sprintf(defaultUserAgentFmt, extra), parent)
}
