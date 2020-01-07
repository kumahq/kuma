package types

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	MinPort uint32 = 1
	MaxPort uint32 = 65535
)

// PortRange reprensents a closed interval of TCP ports constrained by the Lowest and the Highest limits.
//
// E.g.,
// PortRange{8080, 8080} is a range consisting of a single port 8080,
// PortRange{8080, 8081} is a range consisting of 2 ports: 8080 and 8081,
// PortRange{0, 0} is an empty range (which is as a convenient way to indicate that no TCP port is necessary).
type PortRange struct {
	// Lowest port in the range, one of [0, 65535].
	lowest uint32
	// Highest port in the range, one of [Lowest, 65535].
	highest uint32
}

func NewPortRange(lowest, highest uint32) (*PortRange, error) {
	r := PortRange{lowest, highest}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

func MustExactPort(port uint32) PortRange {
	return MustPortRange(port, port)
}

func MustPortRange(lowest, highest uint32) PortRange {
	r, err := NewPortRange(lowest, highest)
	if err != nil {
		panic(err)
	}
	return *r
}

func (r PortRange) Empty() bool {
	return r.lowest == 0 && r.highest == 0
}

func (r PortRange) Lowest() uint32 {
	return r.lowest
}

func (r PortRange) Highest() uint32 {
	return r.highest
}

// validate is made non-public since PortRange
// has been designed to always be valid.
func (r PortRange) validate() error {
	if r.Empty() {
		return nil
	}
	if r.lowest < MinPort || r.highest < MinPort ||
		MaxPort < r.lowest || MaxPort < r.highest ||
		r.highest < r.lowest {
		return errors.New(invalidPortRange(r.String()))
	}
	return nil
}

func (r PortRange) String() string {
	switch {
	case r.Empty():
		return ""
	case r.lowest == r.highest:
		return fmt.Sprintf("%d", r.lowest)
	default:
		return fmt.Sprintf("%d-%d", r.lowest, r.highest)
	}
}

func (r *PortRange) Set(value string) error {
	return r.UnmarshalText([]byte(value))
}

func (PortRange) Type() string {
	return "portOrRange"
}

func (r *PortRange) UnmarshalText(text []byte) error {
	value, err := ParsePortRange(string(text))
	if err != nil {
		return err
	}
	*r = *value
	return nil
}

// ParsePortRange parses a string representation of the PortRange.
//
// Valid values include:
// "8080"      - represents a PortRange{8080, 8080}
// "8080-8081" - represents a PortRange{8080, 8081}
// "8080-"     - represents a PortRange{8080, 65535}
// "-8080"     - represents a PortRange{1, 8080}
// ""          - represents an empty port range
// "-"         - represents an empty port range
func ParsePortRange(text string) (*PortRange, error) {
	// split into left and right bounds
	left, right := "", ""
	parts := strings.Split(text, "-")
	switch len(parts) {
	case 1:
		left, right = parts[0], parts[0]
	case 2:
		left, right = parts[0], parts[1]
	default:
		return nil, errors.New(invalidPortRange(text))
	}
	if left == "" && right == "" {
		// empty port range
		return NewPortRange(0, 0)
	}
	// parse input values
	lowest, highest := uint32(0), uint32(0)
	var err error
	lowest, err = parsePortOrDefault(left, MinPort)
	if err != nil {
		return nil, errors.Wrapf(err, invalidPortRange(text))
	}
	highest, err = parsePortOrDefault(right, MaxPort)
	if err != nil {
		return nil, errors.Wrapf(err, invalidPortRange(text))
	}
	// convert into a range
	r, err := NewPortRange(lowest, highest)
	if err != nil || r.Empty() {
		return nil, errors.New(invalidPortRange(text))
	}
	return r, nil
}

func parsePortOrDefault(text string, defaultValue uint32) (uint32, error) {
	if text == "" {
		return defaultValue, nil
	} else {
		if value, err := strconv.ParseUint(text, 10, 32); err != nil {
			return 0, err
		} else {
			return uint32(value), nil
		}
	}
}

func invalidPortRange(text string) string {
	return fmt.Sprintf(`invalid value %q. %s`, text, `Valid port range formats: "8080", "8080-8081", "8080-", "-8080", "" (empty range), "-" (empty range). Valid port values are 1-65535.`)
}
