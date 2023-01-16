package v1alpha1

func (m *FaultInjection) SourceTags() []SingleValueTagSet {
	var setList []SingleValueTagSet
	for _, selector := range m.GetSources() {
		setList = append(setList, selector.Match)
	}
	return setList
}
