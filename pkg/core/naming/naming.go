package naming

// allow either a bool or a predicate func
type boolOrFunc interface {
	bool | func() bool
}

func eval[T boolOrFunc](p T) bool {
	switch v := any(p).(type) {
	case bool:
		return v
	case func() bool:
		return v != nil && v()
	default:
		return false
	}
}

func GetNameOrFallback[T boolOrFunc](pred T, name, fallback string) string {
	if eval(pred) {
		return name
	}
	return fallback
}

func GetNameOrFallbackFunc[T boolOrFunc](pred T) func(value, fallback string) string {
	return func(name, fallback string) string {
		return GetNameOrFallback(pred, name, fallback)
	}
}
