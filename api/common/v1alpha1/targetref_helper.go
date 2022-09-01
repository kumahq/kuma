package v1alpha1

func (x *TargetRef) GetKindEnum() TargetRef_Kind {
	return TargetRef_Kind(TargetRef_Kind_value[x.Kind])
}

func (x *TargetRef) SetKindEnum(e TargetRef_Kind) {
	x.Kind = TargetRef_Kind_name[int32(e)]
}
