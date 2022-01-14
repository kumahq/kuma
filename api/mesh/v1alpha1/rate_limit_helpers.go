package v1alpha1

func (rl *RateLimit) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range rl.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}
