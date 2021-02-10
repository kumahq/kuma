package mesh

import "time"

func (t *TimeoutResource) GetConnectTimeoutOrDefault(defaultConnectTimeout time.Duration) time.Duration {
	if t == nil {
		return defaultConnectTimeout
	}
	connectTimeout := t.Spec.GetConf().GetConnectTimeout()
	if connectTimeout == nil {
		return defaultConnectTimeout
	}
	return connectTimeout.AsDuration()
}

func (t *TimeoutResource) GetHTTPIdleTimeout() *time.Duration {
	if t == nil {
		return nil
	}
	http := t.Spec.GetConf().GetHttp()
	if http == nil {
		return nil
	}
	idle := http.IdleTimeout.AsDuration()
	return &idle
}
