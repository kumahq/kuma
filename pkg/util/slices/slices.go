package slices

func FlatMap[A any, B any](input []A, f func(A) []B) []B {
	var result []B
	for _, a := range input {
		result = append(result, f(a)...)
	}
	return result
}

func FilterMap[A any, B any](input []A, f func(A) (B, bool)) []B {
	var output []B
	for _, a := range input {
		if b, ok := f(a); ok {
			output = append(output, b)
		}
	}
	return output
}
