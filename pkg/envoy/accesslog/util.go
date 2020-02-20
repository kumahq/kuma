package accesslog

// stringSet represents a set of strings.
type stringSet []string

func (s stringSet) Contains(given string) bool {
	for _, value := range s {
		if value == given {
			return true
		}
	}
	return false
}
