package iterables

// IsNilOrEmpty checks if a slice or array is nil or has zero elements
func IsNilOrEmpty[T any](arr *[]T) bool {
    return arr == nil || len(*arr) == 0
}
