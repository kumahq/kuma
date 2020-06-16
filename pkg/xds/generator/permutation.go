package generator

func Permutation(input []string) [][]string {
	rv := [][]string{}
	prev := [][]string{}
	for i := 1; i <= len(input); i++ {
		prev = permutationN(i, input, prev)
		rv = append(rv, prev...)
	}
	return rv
}

func permutationN(n int, input []string, prev [][]string) [][]string {
	var rv [][]string
	if n == 1 {
		for _, s := range input {
			rv = append(rv, []string{s})
		}
		return rv
	}

	for _, p := range prev {
		lastElem := p[len(p)-1]
		idx := indexOf(input, lastElem)
		rest := input[idx+1:]
		for _, r := range rest {
			nstr := make([]string, len(p)+1)
			copy(nstr, append(p, r))
			rv = append(rv, nstr)
		}
	}
	return rv
}

func indexOf(array []string, elem string) int {
	for i, e := range array {
		if e == elem {
			return i
		}
	}
	return -1
}
