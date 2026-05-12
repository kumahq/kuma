package v1alpha1

func RegisterK8sGatewayTypes() {
	knownTypes = append(knownTypes, &MeshGatewayInstance{}, &MeshGatewayInstanceList{})
}

func RegisterK8sGatewayAPITypes() {
	knownTypes = append(knownTypes, &MeshGatewayConfig{}, &MeshGatewayConfigList{})
}
