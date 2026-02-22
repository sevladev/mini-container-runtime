//go:build linux

package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

const (
	BridgeName = "minic0"
	BridgeCIDR = "10.0.42.1/24"
)

func SetupBridge() (*netlink.Bridge, error) {
	br, _ := netlink.LinkByName(BridgeName)
	if br != nil {
		return br.(*netlink.Bridge), nil
	}

	bridge := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name: BridgeName,
		},
	}

	if err := netlink.LinkAdd(bridge); err != nil {
		return nil, fmt.Errorf("create bridge: %w", err)
	}

	addr, err := netlink.ParseAddr(BridgeCIDR)
	if err != nil {
		return nil, fmt.Errorf("parse bridge addr: %w", err)
	}

	if err := netlink.AddrAdd(bridge, addr); err != nil {
		return nil, fmt.Errorf("add bridge addr: %w", err)
	}

	if err := netlink.LinkSetUp(bridge); err != nil {
		return nil, fmt.Errorf("bridge up: %w", err)
	}

	return bridge, nil
}

func BridgeSubnet() *net.IPNet {
	_, subnet, _ := net.ParseCIDR(BridgeCIDR)
	return subnet
}
