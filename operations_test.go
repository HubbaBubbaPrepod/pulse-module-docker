package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestContainerLogsUseSelectedFullID(t *testing.T) {
	first, second := strings.Repeat("a", 64), strings.Repeat("b", 64)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Query().Get("tail") != "200" || r.URL.Query().Get("timestamps") != "1" {
			t.Error("invalid log query", r.URL.String())
		}
		switch r.URL.Path {
		case "/containers/" + first + "/logs":
			w.Write([]byte(" first container\n"))
		case "/containers/" + second + "/logs":
			w.Write([]byte("second container\n"))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
	for _, tc := range []struct{ id, line string }{{first, " first container"}, {second, "second container"}} {
		result, err := handleDockerOperation(context.Background(), client, "logs", tc.id, nil)
		if err != nil || result["containerId"] != tc.id {
			t.Fatal(result, err)
		}
		lines := result["logs"].([]string)
		if len(lines) != 1 || lines[0] != tc.line {
			t.Fatal(lines)
		}
	}
}
func TestContainerControlRequiresGrantAndExactTarget(t *testing.T) {
	id := strings.Repeat("c", 64)
	var calls atomic.Int32
	lastAction := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/json") {
			started := "2026-10-23T10:00:00Z"
			if lastAction == "restart" {
				started = "2026-10-23T10:00:01Z"
			}
			fmt.Fprintf(w, `{"Id":%q,"State":{"Running":%t,"StartedAt":%q}}`, id, lastAction != "stop", started)
			return
		}
		calls.Add(1)
		lastAction = strings.TrimPrefix(r.URL.Path, "/containers/"+id+"/")
		if r.Method != "POST" || !strings.HasPrefix(r.URL.Path, "/containers/"+id+"/") {
			t.Error("wrong target/method", r.Method, r.URL.Path)
		}
		if strings.HasSuffix(r.URL.Path, "/stop") || strings.HasSuffix(r.URL.Path, "/restart") {
			if r.URL.Query().Get("t") != "5" {
				t.Error("grace period missing")
			}
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
	if _, err := handleDockerOperation(context.Background(), client, "stop", id, nil); err == nil {
		t.Fatal("control without grant accepted")
	}
	for _, invalid := range []string{id[:12], "../escape", "name"} {
		if _, err := handleDockerOperation(context.Background(), client, "start", invalid, []string{"actions.execute"}); err == nil {
			t.Fatal("non-full ID accepted", invalid)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("invalid requests reached Docker")
	}
	for _, action := range []string{"start", "restart", "stop"} {
		result, err := handleDockerOperation(context.Background(), client, action, id, []string{"actions.execute"})
		if err != nil || result["containerId"] != id {
			t.Fatal(result, err)
		}
	}
	if calls.Load() != 3 {
		t.Fatal(calls.Load())
	}
}
func TestContainerControlHandlesEngineResponses(t *testing.T) {
	for _, tc := range []struct {
		name, action string
		status       int
		ok           bool
	}{{"already running", "start", 304, true}, {"already stopped", "stop", 304, true}, {"restart conflict", "restart", 409, false}, {"deleted", "stop", 404, false}, {"engine failure", "start", 500, false}} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					fmt.Fprintf(w, `{"Id":%q,"State":{"Running":%t,"StartedAt":"2026-10-23T10:00:00Z"}}`, strings.Repeat("d", 64), tc.action != "stop")
					return
				}
				w.WriteHeader(tc.status)
				w.Write([]byte("private engine diagnostic"))
			}))
			defer srv.Close()
			client := srv.Client()
			client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
			_, err := handleDockerOperation(context.Background(), client, tc.action, strings.Repeat("d", 64), []string{"actions.execute"})
			if (err == nil) != tc.ok {
				t.Fatal(err)
			}
			if err != nil && strings.Contains(err.Error(), "private engine diagnostic") {
				t.Fatal("raw engine body exposed")
			}
		})
	}
}
func TestSelectedLogsDecodeFramesAndRejectTruncation(t *testing.T) {
	raw := make([]byte, 8)
	raw[0] = 1
	binary.BigEndian.PutUint32(raw[4:], 7)
	raw = append(raw, []byte(" hello\n")...)
	lines, err := decodeSelectedLogs(raw)
	if err != nil || len(lines) != 1 || lines[0] != " hello" {
		t.Fatal(lines, err)
	}
	if _, err := decodeSelectedLogs(raw[:len(raw)-1]); err == nil {
		t.Fatal("truncated multiplexed log silently accepted")
	}
	lines, err = decodeSelectedLogs(nil)
	if err != nil || len(lines) != 0 {
		t.Fatal(lines, err)
	}
}

func TestControlLostResponseHasUnknownOutcome(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintf(w, `{"Id":%q,"State":{"Running":true,"StartedAt":"2026-10-23T10:00:00Z"}}`, strings.Repeat("a", 64))
			return
		}
		connection, _, err := w.(http.Hijacker).Hijack()
		if err == nil {
			connection.Close()
		}
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &rewriteDockerTransport{base: srv.URL, transport: client.Transport}
	_, err := handleDockerOperation(context.Background(), client, "restart", strings.Repeat("a", 64), []string{"actions.execute"})
	var uncertain *uncertainDockerError
	if !errors.As(err, &uncertain) {
		t.Fatal("lost response was reported as definite failure", err)
	}
}
