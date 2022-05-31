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
