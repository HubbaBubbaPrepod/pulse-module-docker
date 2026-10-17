//go:build integration

package main

import (
	"context"
	"encoding/json"
	"flag"
	"testing"
	"time"
)

var integrationEngineSocket = flag.String("engine-socket", "", "Explicit local Docker test socket")
var integrationContainer = flag.String("engine-container", "", "Explicit disposable test container ID")

func TestRealEngineVersionedLogs(t *testing.T) {
	if *integrationContainer == "" {
		t.Skip("Provide explicit disposable container ID")
	}
	if !fullContainerID.MatchString(*integrationContainer) {
		t.Fatal("Invalid test container ID")
	}
	client, close, err := dockerClient(*integrationEngineSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	raw, err := get(ctx, client, "/version")
	if err != nil {
		t.Fatal(err)
	}
	var version struct {
		APIVersion string `json:"ApiVersion"`
	}
	if err = json.Unmarshal(raw, &version); err != nil {
		t.Fatal(err)
	}
	raw, err = get(ctx, client, "/v"+version.APIVersion+"/containers/"+*integrationContainer+"/logs?stdout=1&stderr=1&tail=200&timestamps=1")
	if err != nil {
		t.Fatal(err)
	}
	lines, err := decodeSelectedLogs(raw)
	if err != nil || len(lines) == 0 {
		t.Fatal("Fixture logs missing", err)
	}
}
