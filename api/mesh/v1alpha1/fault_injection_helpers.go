package v1alpha1

func (m *FaultInjection) SourceTags() (setList []SingleValueTagSet) {
	for _, selector := range m.GetSources() {
		setList = append(setList, selector.Match)
	}
	return
}
