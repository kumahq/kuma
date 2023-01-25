package pointer

// Deref returns the value the pointer points to. If ptr is nil the function returns zero value
func Deref[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func DerefOr[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}

// To returns pointer to the passed value
func To[T any](t T) *T {
	return &t
}
