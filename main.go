package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type dockerStats struct {
	CPU struct {
		CPUUsage struct {
			Total  uint64   `json:"total_usage"`
			PerCPU []uint64 `json:"percpu_usage"`
		} `json:"cpu_usage"`
		System uint64 `json:"system_cpu_usage"`
		Online int    `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPU struct {
		CPUUsage struct {
			Total uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		System uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	Memory struct {
		Usage uint64            `json:"usage"`
		Limit uint64            `json:"limit"`
		Stats map[string]uint64 `json:"stats"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		Incoming uint64 `json:"rx_bytes"`
		Outgoing uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

func statsValues(s dockerStats) (*float64, float64, float64) {
	var cpu *float64
	if s.CPU.System > s.PreCPU.System && s.CPU.CPUUsage.Total >= s.PreCPU.CPUUsage.Total {
		cores := s.CPU.Online
		if cores == 0 {
			cores = len(s.CPU.CPUUsage.PerCPU)
		}
		if cores > 0 {
			v := float64(s.CPU.CPUUsage.Total-s.PreCPU.CPUUsage.Total) / float64(s.CPU.System-s.PreCPU.System) * float64(cores) * 100
			cpu = &v
		}
	}
	used := s.Memory.Usage
	cache, ok := s.Memory.Stats["inactive_file"]
	if !ok {
		cache = s.Memory.Stats["cache"]
	}
	if cache < used {
		used -= cache
	}
	return cpu, float64(used) / (1024 * 1024), float64(s.Memory.Limit) / (1024 * 1024)
}
func decodeLogs(raw []byte) string {
	if len(raw) < 8 || raw[0] > 2 || raw[1] != 0 || raw[2] != 0 || raw[3] != 0 {
		return string(raw)
	}
	out := []byte{}
	for len(raw) >= 8 {
		size := int(binary.BigEndian.Uint32(raw[4:8]))
		if size > len(raw)-8 {
			break
		}
		out = append(out, raw[8:8+size]...)
		raw = raw[8+size:]
	}
	return string(out)
}
func get(ctx context.Context, client *http.Client, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://docker"+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Docker API returned HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024+1))
	if len(raw) > 1024*1024 {
		return nil, errors.New("Docker response exceeds limit")
	}
	return raw, err
}
func collect(ctx context.Context, client *http.Client, logContainer string) (map[string]any, error) {
	raw, err := get(ctx, client, "/version")
	if err != nil {
		return nil, err
	}
	var version struct{ Version string }
	if json.Unmarshal(raw, &version) != nil {
		return nil, errors.New("Invalid Docker version")
	}
	raw, err = get(ctx, client, "/containers/json?all=1")
	if err != nil {
		return nil, err
	}
	var containers []struct {
		ID     string `json:"Id"`
		Names  []string
		Image  string
		State  string
		Status string
	}
	if json.Unmarshal(raw, &containers) != nil {
		return nil, errors.New("Invalid Docker container list")
	}
	truncated := len(containers) > 100
	if truncated {
		containers = containers[:100]
	}
	rows := make([]map[string]any, len(containers))
	semaphore := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, c := range containers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()
			name := c.ID[:min(12, len(c.ID))]
			if len(c.Names) > 0 {
				name = strings.TrimPrefix(c.Names[0], "/")
			}
			row := map[string]any{"id": c.ID, "name": name, "image": c.Image, "state": c.State, "status": c.Status, "cpu": nil, "memory": nil, "memoryLimit": nil, "networkIn": nil, "networkOut": nil}
			if c.State == "running" {
				if raw, err := get(ctx, client, "/containers/"+c.ID+"/stats?stream=false&one-shot=true"); err == nil {
					var stats dockerStats
					if json.Unmarshal(raw, &stats) == nil {
						cpu, memory, limit := statsValues(stats)
						row["cpu"] = cpu
						row["memory"] = memory
						row["memoryLimit"] = limit
						var in, out uint64
						for _, network := range stats.Networks {
							in += network.Incoming
							out += network.Outgoing
						}
						row["networkIn"] = float64(in) / (1024 * 1024)
						row["networkOut"] = float64(out) / (1024 * 1024)
					}
				}
			}
			rows[i] = row
		}()
	}
	wg.Wait()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	running := 0
	cpu := 0.
	memory := 0.
	knownCPU := false
	for _, r := range rows {
		if r["state"] == "running" {
			running++
		}
		if v, ok := r["cpu"].(*float64); ok && v != nil {
			cpu += *v
			knownCPU = true
		}
		if v, ok := r["memory"].(float64); ok {
			memory += v
		}
	}
	logs := []string{}
	if logContainer != "" {
		valid := false
		for _, c := range containers {
			if c.ID == logContainer || strings.HasPrefix(c.ID, logContainer) && len(logContainer) >= 12 {
				logContainer = c.ID
				valid = true
			}
		}
		if !valid {
			return nil, errors.New("Log container is not in the container list")
		}
		raw, err := get(ctx, client, "/containers/"+logContainer+"/logs?stdout=1&stderr=1&tail=50&timestamps=1")
		if err != nil {
			logs = []string{"Docker logs unavailable for this container"}
		} else {
			logs = strings.Split(strings.TrimSpace(decodeLogs(raw)), "\n")
		}
	}
	var totalCPU any
	if knownCPU {
		totalCPU = cpu
	}
	return map[string]any{"engine": map[string]any{"version": version.Version, "containers": len(rows), "running": running, "truncated": truncated}, "containers": rows, "usage": map[string]any{"cpu": totalCPU, "memory": memory}, "logs": logs}, nil
}
func run() error {
	var input struct {
		Protocol int `json:"protocol"`
		Config   struct {
			Socket       string `json:"socket"`
			LogContainer string `json:"logContainer"`
		} `json:"config"`
	}
	if err := json.NewDecoder(io.LimitReader(os.Stdin, 16384)).Decode(&input); err != nil {
		return err
	}
	if input.Protocol != 1 {
		return errors.New("Unsupported collector protocol")
	}
	client, cleanup, err := dockerClient(input.Config.Socket)
	if err != nil {
		return err
	}
	defer cleanup()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	result, err := collect(ctx, client, input.Config.LogContainer)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
