package main

import (
	"fmt"
	"io"
	"mirrorselect/internal/llog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Mirror struct {
	URL     *url.URL
	Country string
	Latency int64
	Size    int64
	Time    float64
	Ports   bool
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

type ByLatency []Mirror

func (m ByLatency) Len() int           { return len(m) }
func (m ByLatency) Less(i, j int) bool { return m[i].Latency < m[j].Latency }
func (m ByLatency) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type ByTransferSpeed []Mirror

func (m ByTransferSpeed) Len() int           { return len(m) }
func (m ByTransferSpeed) Less(i, j int) bool { return m[i].bps() > m[j].bps() }
func (m ByTransferSpeed) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func TopNByLatency(mirrors []Mirror, n int) []Mirror {
	sort.Sort(ByLatency(mirrors))
	if n > len(mirrors) {
		n = len(mirrors)
	}
	return mirrors[:n]
}

func (m Mirror) bps() float64 {
	return float64(m.Size) / m.Time
}

// TestLatency tests the latency of a mirror by sending a HEAD request to the
// mirror's /dists/<dist>/Release file. It sets the mirror's Latency field to
// the time taken to receive a response in milliseconds.
// It also validates that the mirror has the correct release (noble, focal, etc.)
func (m *Mirror) TestLatency(timeout int, dist string) {
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

func (m *Mirror) TestDownload(dist string) {
	client := http.Client{Timeout: 2 * time.Second}
	// the Release file is guaranteed to be there regardless of release, or arch
	target := fmt.Sprintf("%sdists/%s/Release", m.URL.String(), dist)
	start := time.Now()
	resp, err := client.Get(target)
	if err != nil {
		llog.Errorf("%s failed: %s", target, err)
		m.Valid = false
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		llog.Errorf("failed to read response body: %s", err)
		m.Valid = false
		return
	}
	m.Time = time.Since(start).Seconds()
	m.Size = int64(len(body))
	m.Valid = true
	llog.Debugf("%s %s", humanizeTransferSpeed(m.Size, m.Time), m.URL.Hostname())
}

func filterInvalidMirrors(mirrors []Mirror) []Mirror {
	var valid []Mirror
	for _, mirror := range mirrors {
		if mirror.Valid {
			valid = append(valid, mirror)
		}
	}
	return valid
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

func getMirrors(country string, proto string, arch string) ([]Mirror, error) {
	initial, err := crawlLaunchpad(country)
	if err != nil {
		return initial, err
	}
	filtered := filterMirrors(initial, proto)
	return filtered, nil
}
