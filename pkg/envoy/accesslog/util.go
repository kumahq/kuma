package accesslog

// stringList represents a list of strings.
type stringList []string

func (s stringList) Contains(given string) bool {
	for _, value := range s {
		if value == given {
			return true
		}
	}
	return false
}

func (s stringList) Filter(predicate func(string) bool) (filtered stringList) {
	for _, value := range s {
		if predicate(value) {
			filtered = append(filtered, value)
		}
	}
	return
}

func (s stringList) AppendToSet(dest []string) []string {
	for _, value := range s {
		if !stringList(dest).Contains(value) {
			dest = append(dest, value)
		}
	}
	return dest
}
