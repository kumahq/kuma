package v1alpha1

import (
	"time"
)

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

func (x *Timeout_Conf) GetHTTPStreamIdleTimeout() time.Duration {
	if x == nil {
		return 0
	}
	if sit := x.GetHttp().GetStreamIdleTimeout(); sit != nil {
		return sit.AsDuration()
	}
	return x.GetGrpc().GetStreamIdleTimeout().AsDuration()
}

func (x *Timeout_Conf) GetHTTPMaxStreamDuration() time.Duration {
	if x == nil {
		return 0
	}
	if msd := x.GetHttp().GetMaxStreamDuration(); msd != nil {
		return msd.AsDuration()
	}
	return x.GetGrpc().GetMaxStreamDuration().AsDuration()
}
