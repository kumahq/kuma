package v1alpha1

type PolicyDefault struct {
	AppendRules []Rule
}

func (x *To) GetDefault() interface{} {
	return PolicyDefault{
		AppendRules: x.Rules,
	}
}
