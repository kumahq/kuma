//go:build linux

package netns

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func netNsDeleteNamed(name string) error {
	return netns.DeleteNamed(name)
}

func netNsNewNamed(name string) (netns.NsHandle, error) {
	return netns.NewNamed(name)
}

const NUD_REACHABLE = netlink.NUD_REACHABLE
