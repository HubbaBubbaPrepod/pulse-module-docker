package main

import (
	"encoding/binary"
	"testing"
)

func TestStatsAccountForCPUAndCache(t *testing.T) {
	s := dockerStats{}
	s.CPU.CPUUsage.Total = 300
	s.PreCPU.CPUUsage.Total = 100
	s.CPU.System = 1000
	s.PreCPU.System = 500
	s.CPU.Online = 4
	s.Memory.Usage = 100 * 1024 * 1024
	s.Memory.Limit = 200 * 1024 * 1024
	s.Memory.Stats = map[string]uint64{"inactive_file": 20 * 1024 * 1024}
	cpu, memory, limit := statsValues(s)
	if cpu == nil || *cpu != 160 || memory != 80 || limit != 200 {
		t.Fatal(cpu, memory, limit)
	}
	s.CPU.System = 0
	if cpu, _, _ := statsValues(s); cpu != nil {
		t.Fatal("unsupported CPU must remain unknown")
	}
}
func TestMultiplexedDockerLogs(t *testing.T) {
	raw := make([]byte, 8)
	raw[0] = 1
	binary.BigEndian.PutUint32(raw[4:], 5)
	raw = append(raw, []byte("hello")...)
	if got := decodeLogs(raw); got != "hello" {
		t.Fatal(got)
	}
	if got := decodeLogs([]byte("plain tty")); got != "plain tty" {
		t.Fatal(got)
	}
}
