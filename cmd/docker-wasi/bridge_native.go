//go:build !wasip1

package main

import "errors"

func requestHost(string, map[string]any) (hostReply, error) {
	return hostReply{}, errors.New("The isolated guest requires a Pulse WASI host")
}
