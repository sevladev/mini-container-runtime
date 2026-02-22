//go:build linux

package network

import (
	"fmt"
	"net"
)

type ContainerNetwork struct {
	ContainerID string
	IP          net.IP
}

func Setup(containerID string, pid int) (*ContainerNetwork, error) {
	bridge, err := SetupBridge()
	if err != nil {
		return nil, fmt.Errorf("setup bridge: %w", err)
	}

	if err := SetupNAT(); err != nil {
		return nil, fmt.Errorf("setup NAT: %w", err)
	}

	ip, err := AllocateIP()
	if err != nil {
		return nil, fmt.Errorf("allocate IP: %w", err)
	}

	if err := CreateVethPair(containerID, pid, bridge, ip); err != nil {
		return nil, fmt.Errorf("create veth: %w", err)
	}

	return &ContainerNetwork{
		ContainerID: containerID,
		IP:          ip,
	}, nil
}

func Cleanup(containerID string) {
	RemoveVeth(containerID)
}
