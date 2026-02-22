package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const stateBaseDir = "/var/lib/minic/containers"

type Metadata struct {
	ID        string    `json:"id"`
	Image     string    `json:"image"`
	Command   []string  `json:"command"`
	Status    State     `json:"status"`
	PID       int       `json:"pid"`
	IP        string    `json:"ip"`
	Hostname  string    `json:"hostname"`
	CreatedAt time.Time `json:"created_at"`
	ExitCode  int       `json:"exit_code"`
}

func stateDir(id string) string {
	return filepath.Join(stateBaseDir, id)
}

func statePath(id string) string {
	return filepath.Join(stateDir(id), "config.json")
}

func logsDir(id string) string {
	return filepath.Join(stateDir(id), "logs")
}

func SaveState(m *Metadata) error {
	dir := stateDir(m.ID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	os.MkdirAll(logsDir(m.ID), 0755)

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(m.ID), data, 0644)
}

func LoadState(id string) (*Metadata, error) {
	data, err := os.ReadFile(statePath(id))
	if err != nil {
		return nil, fmt.Errorf("container %q not found", id)
	}
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	reconcileState(&m)
	return &m, nil
}

func reconcileState(m *Metadata) {
	if m.Status != StateRunning {
		return
	}
	if !isProcessAlive(m.PID) {
		m.Status = StateStopped
		SaveState(m)
	}
}

func isProcessAlive(pid int) bool {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

func ListStates() ([]*Metadata, error) {
	entries, err := os.ReadDir(stateBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var containers []*Metadata
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m, err := LoadState(e.Name())
		if err != nil {
			continue
		}
		containers = append(containers, m)
	}
	return containers, nil
}

func RemoveState(id string) error {
	return os.RemoveAll(stateDir(id))
}

func FindByPrefix(prefix string) (*Metadata, error) {
	entries, err := os.ReadDir(stateBaseDir)
	if err != nil {
		return nil, fmt.Errorf("no containers found")
	}

	var match *Metadata
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if len(e.Name()) >= len(prefix) && e.Name()[:len(prefix)] == prefix {
			if match != nil {
				return nil, fmt.Errorf("ambiguous prefix %q", prefix)
			}
			m, err := LoadState(e.Name())
			if err != nil {
				continue
			}
			match = m
		}
	}

	if match == nil {
		return nil, fmt.Errorf("container %q not found", prefix)
	}
	return match, nil
}

func StdoutLogPath(id string) string {
	return filepath.Join(logsDir(id), "stdout.log")
}

func StderrLogPath(id string) string {
	return filepath.Join(logsDir(id), "stderr.log")
}
