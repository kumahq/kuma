package v1alpha1

type PolicyDefault struct {
	Rules []Rule `policyMerge:"mergeValuesByKey"`
}

func (x *To) GetDefault() interface{} {
	return PolicyDefault{
		Rules: x.Rules,
	}
}
