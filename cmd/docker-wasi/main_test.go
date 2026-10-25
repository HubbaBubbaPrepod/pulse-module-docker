package main

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestGuestRequestsOnlyBoundDockerMethods(t *testing.T) {
	for _, action := range []string{"", "logs", "start", "restart", "stop"} {
		input := invocation{Protocol: 2}
		expected := "docker.snapshot"
		if action != "" {
			input.Trigger = "manual"
			input.Action = action
			input.Input = map[string]any{"containerId": "selected"}
			expected = "docker." + action
		}
		calls := 0
		result := invoke(input, func(method string, params map[string]any) (hostReply, error) {
			calls++
			if method != expected {
				t.Fatal("wrong broker operation", method)
			}
			if action == "" && len(params) != 0 {
				t.Fatal("snapshot must not select a host endpoint")
			}
			if action != "" && params["containerId"] != "selected" {
				t.Fatal("target lost")
			}
			return hostReply{OK: true, Status: "ok", Data: json.RawMessage(`{"actual":true}`)}, nil
		})
		if calls != 1 || result.Status != "ok" || string(result.Data) != `{"actual":true}` {
			t.Fatal("broker result lost", result)
		}
	}
}
func TestGuestPreservesDeniedAndUnknownOutcomesWithoutRetry(t *testing.T) {
	for _, status := range []string{"error", "unknown"} {
		calls := 0
		result := invoke(invocation{Protocol: 2, Trigger: "manual", Action: "restart", Input: map[string]any{"containerId": "selected"}}, func(string, map[string]any) (hostReply, error) {
			calls++
			return hostReply{Status: status, Error: "Host denied operation", Data: json.RawMessage(`{}`)}, nil
		})
		if calls != 1 || result.Status != status {
			t.Fatal("outcome changed or replayed", result)
		}
	}
	result := invoke(invocation{Protocol: 2}, func(string, map[string]any) (hostReply, error) {
		return hostReply{}, errors.New("transport unavailable")
	})
	if result.Status != "error" {
		t.Fatal("failed bridge claimed success")
	}
}
func TestGuestRejectsUnknownProtocolAndActionBeforeHostCall(t *testing.T) {
	for _, input := range []invocation{{Protocol: 1}, {Protocol: 2, Trigger: "manual", Action: "exec"}, {Protocol: 2, Trigger: "manual", Action: "logs"}} {
		result := invoke(input, func(string, map[string]any) (hostReply, error) {
			t.Fatal("invalid invocation dispatched")
			return hostReply{}, nil
		})
		if result.Status != "error" {
			t.Fatal("invalid invocation accepted")
		}
	}
}
