package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "--child" {
		every, _ := strconv.Atoi(os.Args[2])
		runChildSleeper(every)
		return
	}

	configPath := flag.String("config", "/etc/honeyagent/config.json", "path to config")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	tel := newTelemetry(cfg)
	defer tel.Close()
	tel.emit("system", "agent_started", map[string]any{"agent_id": cfg.AgentID})

	if err := ensureKasperskyLikeDirs(cfg); err != nil {
		log.Printf("warning: fs mimic init failed: %v", err)
	}

	spawnFakeProcesses(ctx, cfg, tel)
	go heartbeatLoop(ctx, cfg, tel)

	for _, decoy := range cfg.Decoys {
		switch decoy.Proto {
		case "tcp":
			go startTCP(ctx, decoy, tel)
		case "udp":
			go startUDP(ctx, decoy, tel)
		}
	}

	<-sigCh
	log.Println("shutdown signal received")
	tel.emit("system", "agent_stopped", map[string]any{"agent_id": cfg.AgentID})
}

func heartbeatLoop(ctx context.Context, cfg *Config, tel *Telemetry) {
	interval := cfg.Heartbeat.IntervalSec
	if interval <= 0 {
		interval = 60
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tel.emit("edr", "heartbeat", map[string]any{"agent_id": cfg.AgentID})
		}
	}
}
