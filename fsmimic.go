package main

import (
	"fmt"
	"os"
	"time"
)

func ensureKasperskyLikeDirs(cfg *Config) error {
	dirs := cfg.FakeDirs
	if len(dirs) == 0 {
		dirs = []string{
			"/var/log/kaspersky",
			"/var/opt/kaspersky/state",
			"/var/opt/kaspersky/conf",
		}
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	stubs := map[string]string{
		"/var/log/kaspersky/trace.log":    sampleTraceLog(),
		"/var/log/kaspersky/events.db":    sampleEventsDB(),
		"/var/opt/kaspersky/state/active": "1\n",
	}
	for path, content := range stubs {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
	}
	return nil
}

func sampleTraceLog() string {
	ts := time.Now().UTC().Format("2006-01-02 15:04:05")
	return ts + " I core      product started\n" +
		ts + " I update    bases are up to date\n" +
		ts + " I scan      on-access scan active\n"
}

func sampleEventsDB() string {
	return "SQLite format 3\x00events\n"
}

func appendHeartbeat(path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05") + " I core      heartbeat\n")
}
