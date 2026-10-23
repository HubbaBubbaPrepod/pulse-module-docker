//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Microsoft/go-winio"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func pipeEngineFixture(t *testing.T, handler http.Handler) (*http.Client, func()) {
	t.Helper()
	name := fmt.Sprintf(`\\.\pipe\pulse-docker-qa-%d-%d`, os.Getpid(), time.Now().UnixNano())
	listener, err := winio.ListenPipe(name, nil)
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	client, cleanup, err := dockerClient(name)
	if err != nil {
		listener.Close()
		t.Fatal(err)
	}
	return client, func() { cleanup(); server.Close(); listener.Close() }
}
func TestWindowsDockerPipeSupportsFullDuplexHTTP(t *testing.T) {
	client, cleanup := pipeEngineFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Version":"fixture"}`))
	}))
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for i := 0; i < 2; i++ {
		raw, err := get(ctx, client, "/version")
		if err != nil || string(raw) != `{"Version":"fixture"}` {
			t.Fatal("named pipe cannot perform ordinary HTTP requests", err)
		}
	}
}
func TestWindowsDockerPipeAllowsBoundedGraceAndCancellation(t *testing.T) {
	client, cleanup := pipeEngineFixture(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/json") {
			fmt.Fprintf(w, `{"Id":%q,"State":{"Running":false}}`, strings.Repeat("a", 64))
			return
		}
		select {
		case <-time.After(3250 * time.Millisecond):
			w.WriteHeader(204)
		case <-r.Context().Done():
		}
	}))
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	if _, err := handleDockerOperation(ctx, client, "stop", strings.Repeat("a", 64), []string{"actions.execute"}); err != nil {
		t.Fatal("client aborted before the declared stop grace period", err)
	}
	cancelled, cancelRequest := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelRequest()
	start := time.Now()
	_, err := get(cancelled, client, "/wait")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatal("named pipe request did not cancel promptly", err)
	}
}
