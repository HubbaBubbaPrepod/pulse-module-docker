package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

func dockerClient(socket string) (*http.Client, func(), error) {
	if socket == "" {
		socket = "/var/run/docker.sock"
	}
	if !filepath.IsAbs(socket) {
		return nil, nil, errors.New("Docker socket must be an absolute local path")
	}
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	return &http.Client{Transport: transport, Timeout: 20 * time.Second}, transport.CloseIdleConnections, nil
}
