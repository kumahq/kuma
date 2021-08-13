package os

import (
	"math"
)

func RaiseFileLimit() error {
	return nil
}

func CurrentFileLimit() (uint64, error) {
	return math.MaxUint64, nil
}
