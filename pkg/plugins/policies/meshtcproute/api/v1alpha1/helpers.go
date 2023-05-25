package v1alpha1

type PolicyDefault struct {
	Rules []Rule
}

func (x *To) GetDefault() interface{} {
	return PolicyDefault{
		Rules: x.Rules,
	}
}
