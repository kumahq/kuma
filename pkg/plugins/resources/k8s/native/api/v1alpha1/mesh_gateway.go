package v1alpha1

func RegisterK8sGatewayTypes() {
	SchemeBuilder.Register(&MeshGatewayInstance{}, &MeshGatewayInstanceList{})
}

func RegisterK8sGatewayAPITypes() {
	SchemeBuilder.Register(&MeshGatewayConfig{}, &MeshGatewayConfigList{})
}
