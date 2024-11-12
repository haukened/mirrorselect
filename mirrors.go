package main

import (
	"fmt"
	"io"
	"mirrorselect/internal/llog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const BASE_URL = "http://mirrors.ubuntu.com/%s.txt"

type Mirror struct {
	URL     *url.URL
	Latency int64
	Size    float64
	Time    int64
	Valid   bool
}

func NewMirror(target string) (Mirror, bool) {
	parsed, err := url.Parse(target)
	if err != nil {
		llog.Errorf("failed to parse Mirror URL: %s", err)
		return Mirror{}, false
	}
	return Mirror{URL: parsed}, true
}

// TestLatency tests the latency of a mirror by sending a HEAD request to the
// mirror's /dists/<dist>/Release file. It sets the mirror's Latency field to
// the time taken to receive a response in milliseconds.
// It also validates that the mirror has the correct release (noble, focal, etc.)
func (m *Mirror) TestLatency(timeout int64, dist string) {
	if m.URL.String() == "" {
		m.Valid = false
		llog.Errorf("empty URL for mirror %v", m)
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	target := fmt.Sprintf("%sdists/%s/Release", m.URL.String(), dist)
	start := time.Now()
	resp, err := client.Head(target)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			llog.Debugf("%s timed out", target)
		} else {
			llog.Errorf("%s failed: %s", target, err)
		}
		m.Valid = false
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	m.Latency = elapsed.Milliseconds()
	m.Valid = true
	llog.Debugf("%3d ms %s", m.Latency, m.URL.Hostname())
}

func (m *Mirror) TestDownload() {
	start := time.Now()
	resp, err := http.Get(m.URL.String())
	if err != nil {
		llog.Errorf("%s failed: %s", m.URL.String(), err)
		m.Valid = false
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)
	m.Size = float64(resp.ContentLength) / 1024 / 1024
	m.Time = elapsed.Milliseconds()
	m.Valid = true
	llog.Debugf("%3d ms %s", m.Time, m.URL.Hostname())
}

func filterMirrors(mirrors []Mirror, proto string) []Mirror {
	if proto == "any" {
		return mirrors
	}
	var filtered []Mirror
	for _, mirror := range mirrors {
		if mirror.URL.Scheme == proto {
			filtered = append(filtered, mirror)
		}
	}
	return filtered
}

func getMirrors(country string, proto string) ([]Mirror, error) {
	llog.Debugf("Getting mirrors for country %s with protocol %s", country, proto)

	llog.Debugf("Fetching mirrors from %s", fmt.Sprintf(BASE_URL, country))
	resp, err := http.Get(fmt.Sprintf(BASE_URL, country))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	llog.Debugf("Response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch mirrors: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mirrors []Mirror
	urls := strings.Split(string(body), "\n")
	for _, url := range urls {
		if url == "" {
			continue
		}
		mirror, ok := NewMirror(url)
		if !ok {
			continue
		}
		mirrors = append(mirrors, mirror)
	}

	return filterMirrors(mirrors, strings.ToLower(proto)), nil
}
