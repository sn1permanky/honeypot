package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Telemetry struct {
	agentID      string
	file         *os.File
	collectorURL string
	httpClient   *http.Client
	mu           sync.Mutex
}

func newTelemetry(cfg *Config) *Telemetry {
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		log.Printf("telemetry: mkdir %s: %v", cfg.LogDir, err)
	}
	path := filepath.Join(cfg.LogDir, "telemetry.jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		log.Fatalf("telemetry: open %s: %v", path, err)
	}
	return &Telemetry{
		agentID:      cfg.AgentID,
		file:         f,
		collectorURL: cfg.CollectorURL,
		httpClient:   &http.Client{Timeout: 3 * time.Second},
	}
}

func (t *Telemetry) emit(category, eventType string, fields map[string]any) {
	event := map[string]any{
		"time":     time.Now().UTC().Format(time.RFC3339),
		"agent_id": t.agentID,
		"category": category,
		"type":     eventType,
		"fields":   fields,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.file.Write(payload)
	t.file.Write([]byte("\n"))
	if t.collectorURL != "" {
		go t.sendToCollector(payload)
	}
}

func (t *Telemetry) sendToCollector(payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", t.collectorURL, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (t *Telemetry) Close() error {
	return t.file.Close()
}
