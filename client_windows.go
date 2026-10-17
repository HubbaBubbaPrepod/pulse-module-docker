package main

import (
	"context"
	"errors"
	"github.com/Microsoft/go-winio"
	"net"
	"net/http"
	"strings"
	"time"
)

func dockerClient(socket string) (*http.Client, func(), error) {
	if socket == "" {
		socket = `\\.\pipe\docker_engine`
	}
	if !strings.HasPrefix(socket, `\\.\pipe\`) {
		return nil, nil, errors.New("Docker endpoint must be a local named pipe")
	}
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) { return winio.DialPipeContext(ctx, socket) }}
	return &http.Client{Transport: transport, Timeout: 20 * time.Second}, transport.CloseIdleConnections, nil
}
