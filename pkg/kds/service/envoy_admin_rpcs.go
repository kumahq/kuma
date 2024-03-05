package service

import (
	util_grpc "github.com/kumahq/kuma/pkg/util/grpc"
)

const (
	ConfigDumpRPC = "XDS Config Dump"
	StatsRPC      = "Stats"
	ClustersRPC   = "Clusters"
)

type EnvoyAdminRPCs struct {
	XDSConfigDump util_grpc.ReverseUnaryRPCs
	Stats         util_grpc.ReverseUnaryRPCs
	Clusters      util_grpc.ReverseUnaryRPCs
}

func NewEnvoyAdminRPCs() EnvoyAdminRPCs {
	return EnvoyAdminRPCs{
		XDSConfigDump: util_grpc.NewReverseUnaryRPCs(),
		Stats:         util_grpc.NewReverseUnaryRPCs(),
		Clusters:      util_grpc.NewReverseUnaryRPCs(),
	}
}
