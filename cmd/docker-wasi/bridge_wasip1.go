package main

import (
	"encoding/json"
	"errors"
	"runtime"
	"unsafe"
)

//go:wasmimport pulse_host request
func hostRequest(request, length, response, capacity uint32) uint32

func requestHost(method string, params map[string]any) (hostReply, error) {
	var reply hostReply
	raw, err := json.Marshal(map[string]any{"method": method, "params": params})
	if err != nil || len(raw) == 0 || len(raw) > 64*1024 {
		return reply, errors.New("Invalid broker request")
	}
	output := make([]byte, 512*1024)
	size := hostRequest(uint32(uintptr(unsafe.Pointer(&raw[0]))), uint32(len(raw)), uint32(uintptr(unsafe.Pointer(&output[0]))), uint32(len(output)))
	runtime.KeepAlive(raw)
	runtime.KeepAlive(output)
	if size == 0 || size > uint32(len(output)) || json.Unmarshal(output[:size], &reply) != nil {
		return reply, errors.New("Invalid broker response")
	}
	return reply, nil
}
