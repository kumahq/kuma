//go:build !linux

package netns

import (
	"github.com/pkg/errors"
	"github.com/vishvananda/netns"
)

func netNsDeleteNamed(string) error {
	return errors.New("Only supported on linux")
}

func netNsNewNamed(string) (netns.NsHandle, error) {
	return 0, errors.New("Only supported on linux")
}

const NUD_REACHABLE = 0x00
