package v1alpha1

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/generic"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ generic.Insight = &DataplaneInsight{}

func NewSubscriptionStatus() *DiscoverySubscriptionStatus {
	return &DiscoverySubscriptionStatus{
		Total: &DiscoveryServiceStats{},
		Cds:   &DiscoveryServiceStats{},
		Eds:   &DiscoveryServiceStats{},
		Lds:   &DiscoveryServiceStats{},
		Rds:   &DiscoveryServiceStats{},
	}
}

func NewVersion() *Version {
	return &Version{
		KumaDp: &KumaDpVersion{
			Version:   "",
			GitTag:    "",
			GitCommit: "",
			BuildDate: "",
		},
		Envoy: &EnvoyVersion{
			Version: "",
			Build:   "",
		},
		Dependencies: map[string]string{},
	}
}

func (x *EnvoyVersion) ParseVersion() (version, label string) {
	parts := strings.SplitN(x.Version, "-", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	} else {
		return x.Version, ""
	}
}

func (x *DataplaneInsight) IsOnline() bool {
	for _, s := range x.GetSubscriptions() {
		if s.GetConnectTime() != nil && s.GetDisconnectTime() == nil {
			return true
		}
	}
	return false
}

func (x *DataplaneInsight) GetSubscription(id string) (int, *DiscoverySubscription) {
	for i, s := range x.GetSubscriptions() {
		if s.Id == id {
			return i, s
		}
	}
	return -1, nil
}

func (x *DataplaneInsight) UpdateCert(generation time.Time, expiration time.Time, issuedBackend string, supportedBackends []string) error {
	if x.MTLS == nil {
		x.MTLS = &DataplaneInsight_MTLS{}
	}
	ts := util_proto.MustTimestampProto(expiration)
	if err := ts.CheckValid(); err != nil {
		return err
	}
	x.MTLS.CertificateExpirationTime = ts
	x.MTLS.CertificateRegenerations++
	ts = util_proto.MustTimestampProto(generation)
	if err := ts.CheckValid(); err != nil {
		return err
	}
	x.MTLS.IssuedBackend = issuedBackend
	x.MTLS.SupportedBackends = supportedBackends
	x.MTLS.LastCertificateRegeneration = ts
	return nil
}

func (x *DataplaneInsight) UpdateSubscription(s generic.Subscription) error {
	if x == nil {
		return nil
	}
	discoverySubscription, ok := s.(*DiscoverySubscription)
	if !ok {
		return errors.Errorf("invalid type %T for DataplaneInsight", s)
	}
	i, old := x.GetSubscription(discoverySubscription.Id)
	if old != nil {
		x.Subscriptions[i] = discoverySubscription
	} else {
		x.finalizeSubscriptions()
		x.Subscriptions = append(x.Subscriptions, discoverySubscription)
	}
	return nil
}

// If Kuma CP was killed ungracefully then we can get a subscription without a DisconnectTime.
// Because of the way we process subscriptions the lack of DisconnectTime on old subscription
// will cause wrong status.
func (x *DataplaneInsight) finalizeSubscriptions() {
	now := util_proto.Now()
	for _, subscription := range x.GetSubscriptions() {
		if subscription.DisconnectTime == nil {
			subscription.DisconnectTime = now
		}
	}
}

func (x *DataplaneInsight) GetLastSubscription() generic.Subscription {
	if len(x.GetSubscriptions()) == 0 {
		return (*DiscoverySubscription)(nil)
	}
	return x.GetSubscriptions()[len(x.GetSubscriptions())-1]
}

func (x *DiscoverySubscription) SetDisconnectTime(t time.Time) {
	x.DisconnectTime = util_proto.MustTimestampProto(t)
}

func (x *DataplaneInsight) Sum(v func(*DiscoverySubscription) uint64) uint64 {
	var result uint64 = 0
	for _, s := range x.GetSubscriptions() {
		result += v(s)
	}
	return result
}

func (s *DiscoverySubscriptionStatus) StatsOf(typeUrl string) *DiscoveryServiceStats {
	if s == nil {
		return &DiscoveryServiceStats{}
	}
	// we rely on type URL suffix to get rid of the dependency on concrete V2 / V3 implementation
	switch {
	case strings.HasSuffix(typeUrl, "Cluster"):
		if s.Cds == nil {
			s.Cds = &DiscoveryServiceStats{}
		}
		return s.Cds
	case strings.HasSuffix(typeUrl, "ClusterLoadAssignment"):
		if s.Eds == nil {
			s.Eds = &DiscoveryServiceStats{}
		}
		return s.Eds
	case strings.HasSuffix(typeUrl, "Listener"):
		if s.Lds == nil {
			s.Lds = &DiscoveryServiceStats{}
		}
		return s.Lds
	case strings.HasSuffix(typeUrl, "RouteConfiguration"):
		if s.Rds == nil {
			s.Rds = &DiscoveryServiceStats{}
		}
		return s.Rds
	default:
		return &DiscoveryServiceStats{}
	}
}
