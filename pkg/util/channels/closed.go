package channels

// IsClosed checks if channel is closed by reading the value. It is useful for checking
func IsClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
