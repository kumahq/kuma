package v1alpha1

func (rl *RateLimit) SourceTags() []SingleValueTagSet {
	var setList []SingleValueTagSet
	for _, selector := range rl.GetSources() {
		setList = append(setList, selector.Match)
	}
	return setList
}
