package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
)

type pipeAddress string

func (p pipeAddress) Network() string { return "pipe" }
func (p pipeAddress) String() string  { return string(p) }

type pipeConnection struct {
	*os.File
	name string
}

func (p *pipeConnection) LocalAddr() net.Addr                { return pipeAddress(p.name) }
func (p *pipeConnection) RemoteAddr() net.Addr               { return pipeAddress(p.name) }
func (p *pipeConnection) SetDeadline(t time.Time) error      { return p.File.SetDeadline(t) }
func (p *pipeConnection) SetReadDeadline(t time.Time) error  { return p.File.SetReadDeadline(t) }
func (p *pipeConnection) SetWriteDeadline(t time.Time) error { return p.File.SetWriteDeadline(t) }
func dockerClient(socket string) (*http.Client, func(), error) {
	if socket == "" {
		socket = `\\.\pipe\docker_engine`
	}
	if !strings.HasPrefix(socket, `\\.\pipe\`) {
		return nil, nil, errors.New("Docker endpoint must be a local named pipe")
	}
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		name, err := syscall.UTF16PtrFromString(socket)
		if err != nil {
			return nil, err
		}
		handle, err := syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_OVERLAPPED, 0)
		for errors.Is(err, syscall.Errno(231)) { // ERROR_PIPE_BUSY: retry within the request deadline.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(20 * time.Millisecond):
			}
			handle, err = syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_OVERLAPPED, 0)
		}
		if err != nil {
			return nil, err
		}
		file := os.NewFile(uintptr(handle), socket)
		if file == nil {
			_ = syscall.CloseHandle(handle)
			return nil, errors.New("cannot initialize Docker named pipe")
		}
		return &pipeConnection{File: file, name: socket}, nil
	}}
	return &http.Client{Transport: transport, Timeout: 3 * time.Second}, transport.CloseIdleConnections, nil
}
