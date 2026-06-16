package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

func spawnFakeProcesses(ctx context.Context, cfg *Config, tel *Telemetry) {
	selfExe, err := os.Readlink("/proc/self/exe")
	if err != nil {
		log.Printf("cannot resolve self exe: %v", err)
		return
	}
	for _, p := range cfg.FakeProcs {
		cmd := exec.CommandContext(ctx, selfExe, "--child", strconv.Itoa(p.SleepEvery))
		cmd.Args[0] = p.Argv0 // имя процесса в ps/top задаётся значением argv[0]
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			log.Printf("spawn %s failed: %v", p.Argv0, err)
			continue
		}
		tel.emit("edr", "fake_proc_started", map[string]any{
			"name": p.Name,
			"pid":  cmd.Process.Pid,
		})
	}
}

func runChildSleeper(every int) {
	if every <= 0 {
		every = 30
	}
	ticker := time.NewTicker(time.Duration(every) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		appendHeartbeat("/var/log/kaspersky/trace.log")
	}
}
