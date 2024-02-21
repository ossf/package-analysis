package useragent_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ossf/package-analysis/internal/useragent"
)

func TestRoundTripper(t *testing.T) {
	want := "test user agent string"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("user-agent")
		if got != want {
			t.Errorf("User Agent = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := http.Client{
		Transport: useragent.RoundTripper(want, http.DefaultTransport),
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get() = %v; want no error", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get() status = %v; want 200", resp.StatusCode)
	}
}

func TestDefaultRoundTripper(t *testing.T) {
	want := "package-analysis (github.com/ossf/package-analysis, extra)"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("user-agent")
		if got != want {
			t.Errorf("User Agent = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := http.Client{
		Transport: useragent.DefaultRoundTripper(http.DefaultTransport, "extra"),
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get() = %v; want no error", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get() status = %v; want 200", resp.StatusCode)
	}
}

func TestDefaultRoundTripper_NoExtra(t *testing.T) {
	want := "package-analysis (github.com/ossf/package-analysis)"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("user-agent")
		if got != want {
			t.Errorf("User Agent = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := http.Client{
		Transport: useragent.DefaultRoundTripper(http.DefaultTransport, ""),
	}
	resp, err := c.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get() = %v; want no error", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Get() status = %v; want 200", resp.StatusCode)
	}
}
