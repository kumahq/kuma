package common

import (
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	ControllerName = gatewayapi.GatewayController("gateways.kuma.io/controller")
)
