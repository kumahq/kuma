package v1alpha1

import "time"

func (x *Timeout_Conf) GetConnectTimeoutOrDefault(defaultConnectTimeout time.Duration) time.Duration {
	if x == nil {
		return defaultConnectTimeout
	}
	connectTimeout := x.GetConnectTimeout()
	if connectTimeout == nil {
		return defaultConnectTimeout
	}
	return connectTimeout.AsDuration()
}

func (x *Timeout_Conf) GetTCPIdleTimeout() time.Duration {
	if x == nil {
		return 0
	}
	tcp := x.GetTcp()
	if tcp == nil {
		return 0
	}
	return tcp.IdleTimeout.AsDuration()
}

func (x *Timeout_Conf) GetHTTPIdleTimeout() time.Duration {
	if x == nil {
		return 0
	}
	http := x.GetHttp()
	if http == nil {
		return 0
	}
	return http.IdleTimeout.AsDuration()
}

func (x *Timeout_Conf) GetHTTPRequestTimeout() time.Duration {
	if x == nil {
		return 0
	}
	http := x.GetHttp()
	if http == nil {
		return 0
	}
	return http.RequestTimeout.AsDuration()
}

func (x *Timeout_Conf) GetGRPCStreamIdleTimeout() time.Duration {
	if x == nil {
		return 0
	}
	grpc := x.GetGrpc()
	if grpc == nil {
		return 0
	}
	return grpc.StreamIdleTimeout.AsDuration()
}

func (x *Timeout_Conf) GetGRPCMaxStreamDuration() *time.Duration {
	if x == nil {
		return nil
	}
	grpc := x.GetGrpc()
	if grpc == nil {
		return nil
	}
	maxStreamDuration := grpc.MaxStreamDuration.AsDuration()
	return &maxStreamDuration
}
