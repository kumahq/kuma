package v1alpha1

func (x *MeshTrafficPermission_Conf) GetActionEnum() MeshTrafficPermission_Conf_Action {
	return MeshTrafficPermission_Conf_Action(MeshTrafficPermission_Conf_Action_value[x.Action])
}

func (x *MeshTrafficPermission_Conf) SetActionEnum(e MeshTrafficPermission_Conf_Action) {
	x.Action = MeshTrafficPermission_Conf_Action_name[int32(e)]
}
