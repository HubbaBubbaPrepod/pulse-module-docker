package main

import (
	"encoding/json"
	"io"
	"os"
)

type invocation struct {
	Protocol int            `json:"protocol"`
	Trigger  string         `json:"trigger"`
	Action   string         `json:"action"`
	Input    map[string]any `json:"input"`
}
type hostReply struct {
	OK     bool            `json:"ok"`
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
type result struct {
	Protocol int             `json:"protocol"`
	Status   string          `json:"status"`
	Data     json.RawMessage `json:"data"`
	Error    string          `json:"error,omitempty"`
}

func invoke(input invocation, call func(string, map[string]any) (hostReply, error)) result {
	out := result{Protocol: 2, Status: "error", Data: json.RawMessage(`{}`), Error: "Invalid Docker invocation"}
	if input.Protocol != 2 {
		return out
	}
	method := "docker.snapshot"
	params := map[string]any{}
	if input.Trigger == "manual" {
		switch input.Action {
		case "logs", "start", "restart", "stop":
		default:
			return out
		}
		id, ok := input.Input["containerId"].(string)
		if !ok || id == "" {
			return out
		}
		method = "docker." + input.Action
		params["containerId"] = id
	} else if input.Action != "" {
		return out
	}
	reply, err := call(method, params)
	if err != nil {
		out.Error = "Host broker unavailable"
		return out
	}
	if reply.Status != "ok" && reply.Status != "error" && reply.Status != "unknown" {
		out.Error = "Invalid host response"
		return out
	}
	if reply.Status == "ok" && !reply.OK {
		out.Error = "Host broker denied operation"
		return out
	}
	if len(reply.Data) > 0 {
		var data map[string]json.RawMessage
		if json.Unmarshal(reply.Data, &data) != nil || data == nil {
			return out
		}
		out.Data = reply.Data
	}
	out.Status = reply.Status
	out.Error = reply.Error
	if len(out.Error) > 256 {
		out.Error = "Host operation failed"
	}
	return out
}

func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 2*1024*1024+1))
	var input invocation
	if err != nil || len(raw) > 2*1024*1024 || json.Unmarshal(raw, &input) != nil {
		os.Exit(1)
	}
	if json.NewEncoder(os.Stdout).Encode(invoke(input, requestHost)) != nil {
		os.Exit(1)
	}
}
