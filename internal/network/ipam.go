//go:build linux

package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

const ipamFile = "/var/lib/minic/network/ipam.json"

type ipamState struct {
	NextIP byte `json:"next_ip"`
}

var ipamMu sync.Mutex

func AllocateIP() (net.IP, error) {
	ipamMu.Lock()
	defer ipamMu.Unlock()

	state := loadIPAM()

	if state.NextIP == 0 {
		state.NextIP = 2
	}

	if state.NextIP > 254 {
		return nil, fmt.Errorf("no IPs available")
	}

	ip := net.IPv4(10, 0, 42, state.NextIP)
	state.NextIP++

	if err := saveIPAM(state); err != nil {
		return nil, err
	}

	return ip, nil
}

func loadIPAM() ipamState {
	data, err := os.ReadFile(ipamFile)
	if err != nil {
		return ipamState{}
	}
	var s ipamState
	json.Unmarshal(data, &s)
	return s
}

func saveIPAM(s ipamState) error {
	os.MkdirAll("/var/lib/minic/network", 0755)
	data, _ := json.Marshal(s)
	return os.WriteFile(ipamFile, data, 0644)
}
