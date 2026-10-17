package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
)

var fullContainerID = regexp.MustCompile(`^[a-f0-9]{64}$`)

type uncertainDockerError struct{ message string }

func (e *uncertainDockerError) Error() string { return e.message }

func decodeSelectedLogs(raw []byte) ([]string, error) {
	var text string
	if len(raw) >= 4 && (raw[0] == 1 || raw[0] == 2) && raw[1] == 0 && raw[2] == 0 && raw[3] == 0 {
		var output strings.Builder
		for len(raw) > 0 {
			if len(raw) < 8 || (raw[0] != 1 && raw[0] != 2) || raw[1] != 0 || raw[2] != 0 || raw[3] != 0 {
				return nil, errors.New("Invalid multiplexed log frame")
			}
			size := uint64(binary.BigEndian.Uint32(raw[4:8]))
			if size > uint64(len(raw)-8) {
				return nil, errors.New("Incomplete multiplexed log frame")
			}
			output.Write(raw[8 : 8+int(size)])
			raw = raw[8+int(size):]
		}
		text = output.String()
	} else {
		text = string(raw)
	}
	text = strings.TrimRight(text, "\r\n")
	if text == "" {
		return []string{}, nil
	}
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	if len(lines) > 200 {
		lines = lines[len(lines)-200:]
	}
	return lines, nil
}

func handleDockerOperation(ctx context.Context, client *http.Client, action, id string, grants []string) (map[string]any, error) {
	if !fullContainerID.MatchString(id) {
		return nil, errors.New("A full 64-character container ID is required")
	}
	result := map[string]any{"containerId": id, "action": action, "at": time.Now().UTC().Format(time.RFC3339Nano)}
	if action == "logs" {
		raw, err := get(ctx, client, "/containers/"+id+"/logs?stdout=1&stderr=1&tail=200&timestamps=1")
		if err != nil {
			return nil, fmt.Errorf("Container logs unavailable: %w", err)
		}
		lines, err := decodeSelectedLogs(raw)
		if err != nil {
			return nil, err
		}
		result["logs"] = lines
		return result, nil
	}
	if !slices.Contains([]string{"start", "restart", "stop"}, action) {
		return nil, errors.New("Unsupported container action")
	}
	if !slices.Contains(grants, "actions.execute") {
		return nil, errors.New("Container management is not granted")
	}
	endpoint := "http://docker/containers/" + id + "/" + action
	if action == "stop" || action == "restart" {
		endpoint += "?t=5"
	}
	request, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, &uncertainDockerError{"Docker control request failed or timed out; verify the observed container state"}
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode == 204 {
		result["outcome"] = "completed"
		return result, nil
	}
	if response.StatusCode == 304 && (action == "start" || action == "stop") {
		result["outcome"] = "already-in-state"
		return result, nil
	}
	if response.StatusCode >= 500 {
		return nil, &uncertainDockerError{fmt.Sprintf("Docker control returned HTTP %d; verify the observed container state", response.StatusCode)}
	}
	return nil, fmt.Errorf("Docker control returned HTTP %d", response.StatusCode)
}
