package parameters

// NoFlush will generate arguments for the "-n, --noflush" flag
// Don't flush the previous contents of the table. If not specified, both
// commands flush (delete) all previous contents of the respective table.
//
// ref. iptables-restore(8) > DESCRIPTION
func NoFlush() *Parameter {
	return &Parameter{
		long:       "--noflush",
		short:      "-n",
		parameters: nil, // flag doesn't expect parameters
		negate:     nil, // no negation allowed
	}
}
