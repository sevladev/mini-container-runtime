//go:build linux

package network

import (
	"fmt"
	"net"
	"runtime"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func CreateVethPair(containerID string, pid int, bridge *netlink.Bridge, ip net.IP) error {
	hostVeth := "veth-" + containerID[:8]
	peerVeth := "peer-" + containerID[:8]

	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostVeth,
		},
		PeerName: peerVeth,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("create veth pair: %w", err)
	}

	hostLink, err := netlink.LinkByName(hostVeth)
	if err != nil {
		return fmt.Errorf("get host veth: %w", err)
	}

	if err := netlink.LinkSetMaster(hostLink, bridge); err != nil {
		return fmt.Errorf("attach to bridge: %w", err)
	}

	if err := netlink.LinkSetUp(hostLink); err != nil {
		return fmt.Errorf("host veth up: %w", err)
	}

	peerLink, err := netlink.LinkByName(peerVeth)
	if err != nil {
		return fmt.Errorf("get peer veth: %w", err)
	}

	if err := netlink.LinkSetNsPid(peerLink, pid); err != nil {
		return fmt.Errorf("move veth to netns: %w", err)
	}

	if err := configureContainerNet(pid, peerVeth, ip); err != nil {
		return fmt.Errorf("configure container net: %w", err)
	}

	return nil
}

func configureContainerNet(pid int, peerName string, ip net.IP) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origNs, err := netns.Get()
	if err != nil {
		return fmt.Errorf("get current netns: %w", err)
	}
	defer origNs.Close()

	containerNs, err := netns.GetFromPid(pid)
	if err != nil {
		return fmt.Errorf("get container netns: %w", err)
	}
	defer containerNs.Close()

	if err := netns.Set(containerNs); err != nil {
		return fmt.Errorf("enter container netns: %w", err)
	}
	defer netns.Set(origNs)

	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("get lo: %w", err)
	}
	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("lo up: %w", err)
	}

	peer, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("get peer: %w", err)
	}
	if err := netlink.LinkSetName(peer, "eth0"); err != nil {
		return fmt.Errorf("rename to eth0: %w", err)
	}

	eth0, err := netlink.LinkByName("eth0")
	if err != nil {
		return fmt.Errorf("get eth0: %w", err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: BridgeSubnet().Mask,
		},
	}
	if err := netlink.AddrAdd(eth0, addr); err != nil {
		return fmt.Errorf("add addr to eth0: %w", err)
	}

	if err := netlink.LinkSetUp(eth0); err != nil {
		return fmt.Errorf("eth0 up: %w", err)
	}

	gateway := net.IPv4(10, 0, 42, 1)
	defaultRoute := &netlink.Route{
		Gw: gateway,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("add default route: %w", err)
	}

	return nil
}

func RemoveVeth(containerID string) {
	hostVeth := "veth-" + containerID[:8]
	if link, err := netlink.LinkByName(hostVeth); err == nil {
		netlink.LinkDel(link)
	}
}
