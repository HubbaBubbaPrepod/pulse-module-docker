package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCollectCPUUsesTwoCurrentSamples(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version":
			w.Write([]byte(`{"Version":"test"}`))
		case "/containers/json":
			w.Write([]byte(`[{"Id":"actual","Names":["/actual"],"State":"running"}]`))
		default:
			s := dockerStats{}
			s.CPU.Online = 2
			s.CPU.CPUUsage.Total = 10
			s.CPU.System = 100
			if calls.Add(1) > 1 {
				s.CPU.CPUUsage.Total = 25
				s.CPU.System = 200
			}
			s.Memory.Usage = 64 * 1024 * 1024
			json.NewEncoder(w).Encode(s)
		}
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
	result, err := collect(context.Background(), client, "")
	if err != nil {
		t.Fatal(err)
	}
	usage := result["usage"].(map[string]any)
	cpu, ok := usage["cpu"].(float64)
	if !ok || cpu != 30 {
		t.Fatalf("current CPU should be 30%%, got %v", usage["cpu"])
	}
	if usage["memory"] != float64(64) {
		t.Fatal(usage)
	}
}

func TestCollectDoesNotPresentPartialMetricsAsTotals(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/version":
			w.Write([]byte(`{"Version":"test"}`))
		case "/containers/json":
			w.Write([]byte(`[{"Id":"known","State":"running"},{"Id":"missing","State":"running"}]`))
		case "/containers/known/stats":
			s := dockerStats{}
			s.CPU.System = 100
			s.CPU.Online = 2
			s.Memory.Usage = 64 * 1024 * 1024
			json.NewEncoder(w).Encode(s)
		default:
			http.Error(w, "Unavailable", 503)
		}
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
	result, err := collect(context.Background(), client, "")
	if err != nil {
		t.Fatal(err)
	}
	usage := result["usage"].(map[string]any)
	if usage["cpu"] != nil || usage["memory"] != nil {
		t.Fatalf("incomplete totals must be unknown, got %v", usage)
	}
}

type rewriteDockerTransport struct {
	base      string
	transport http.RoundTripper
}

func (t *rewriteDockerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	copy := r.Clone(r.Context())
	target := *r.URL
	base, _ := http.NewRequest("GET", t.base, nil)
	target.Scheme = base.URL.Scheme
	target.Host = base.URL.Host
	copy.URL = &target
	return t.transport.RoundTrip(copy)
}

func TestStatsAccountForCPUAndCache(t *testing.T) {
	s := dockerStats{}
	s.CPU.CPUUsage.Total = 300
	s.PreCPU.CPUUsage.Total = 100
	s.CPU.System = 1000
	s.PreCPU.System = 500
	s.CPU.Online = 4
	s.Memory.Usage = 100 * 1024 * 1024
	s.Memory.Limit = 200 * 1024 * 1024
	s.Memory.Stats = map[string]uint64{"inactive_file": 20 * 1024 * 1024}
	cpu, memory, limit := statsValues(s)
	if cpu == nil || *cpu != 160 || memory != 80 || limit != 200 {
		t.Fatal(cpu, memory, limit)
	}
	s.CPU.System = 0
	if cpu, _, _ := statsValues(s); cpu != nil {
		t.Fatal("unsupported CPU must remain unknown")
	}
}
func TestMultiplexedDockerLogs(t *testing.T) {
	raw := make([]byte, 8)
	raw[0] = 1
	binary.BigEndian.PutUint32(raw[4:], 5)
	raw = append(raw, []byte("hello")...)
	if got := decodeLogs(raw); got != "hello" {
		t.Fatal(got)
	}
	if got := decodeLogs([]byte("plain tty")); got != "plain tty" {
		t.Fatal(got)
	}
}
