package lookup

import (
	"net"
)

type LookupIPFunc func(string) ([]net.IP, error)
