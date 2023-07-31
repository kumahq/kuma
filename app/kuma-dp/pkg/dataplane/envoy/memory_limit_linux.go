package envoy

import (
	"os"
	"strconv"
	"strings"

	"github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/cgroups"
)

type UIntOrString struct {
	Type   string
	UInt   uint64
	String string
}

func DetectMaxMemory() uint64 {
	switch cgroups.Mode() {
	case cgroups.Legacy:
		res := maybeReadAsBytes("/sys/fs/cgroup/memory.limit_in_bytes")
		if res != nil && res.Type == "int" {
			return res.UInt
		}
	case cgroups.Hybrid, cgroups.Unified:
		res := maybeReadAsBytes("/sys/fs/cgroup/memory.max")
		if res != nil && res.Type == "int" {
			return res.UInt
		}
	}
	return 0
}

func maybeReadAsBytes(path string) *UIntOrString {
	byteContents, err := os.ReadFile(path)
	if err == nil {
		contents := strings.TrimSpace(string(byteContents))
		bytes, err := strconv.ParseUint(contents, 10, 64)
		if err != nil {
			return &UIntOrString{
				Type:   "string",
				String: contents,
			}
		}
		return &UIntOrString{
			Type: "int",
			UInt: bytes,
		}
	}
	return nil
}
