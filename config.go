package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	AgentID      string         `json:"agent_id"`
	LogDir       string         `json:"log_dir"`
	CollectorURL string         `json:"collector_url,omitempty"`
	FakeDirs     []string       `json:"fake_dirs"`
	FakeProcs    []FakeProcSpec `json:"fake_procs"`
	Decoys       []DecoySpec    `json:"decoys"`
	Heartbeat    HeartbeatSpec  `json:"heartbeat"`
}

type FakeProcSpec struct {
	Name       string `json:"name"`
	Argv0      string `json:"argv0"`
	SleepEvery int    `json:"sleep_every_sec"`
}

type DecoySpec struct {
	Port     int               `json:"port"`
	Proto    string            `json:"proto"`
	Behavior string            `json:"behavior"`
	Params   map[string]string `json:"params,omitempty"`
}

type HeartbeatSpec struct {
	IntervalSec int `json:"interval_sec"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.AgentID == "" {
		cfg.AgentID = "honeyagent"
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "/var/log/honeyagent"
	}
	return &cfg, nil
}
