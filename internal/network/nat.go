//go:build linux

package network

import (
	"fmt"
	"os"
	"os/exec"
)

func SetupNAT() error {
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0644); err != nil {
		return fmt.Errorf("enable ip_forward: %w", err)
	}

	subnet := BridgeSubnet().String()

	out, err := exec.Command("iptables", "-t", "nat", "-C", "POSTROUTING", "-s", subnet, "-j", "MASQUERADE").CombinedOutput()
	if err == nil {
		return nil
	}
	_ = out

	if out, err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", subnet, "-j", "MASQUERADE").CombinedOutput(); err != nil {
		return fmt.Errorf("add NAT rule: %s: %w", string(out), err)
	}

	return nil
}
