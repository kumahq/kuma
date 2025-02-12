//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	commonv1alpha1 "github.com/kumahq/kuma/api/common/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AffinityTag) DeepCopyInto(out *AffinityTag) {
	*out = *in
	if in.Weight != nil {
		in, out := &in.Weight, &out.Weight
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AffinityTag.
func (in *AffinityTag) DeepCopy() *AffinityTag {
	if in == nil {
		return nil
	}
	out := new(AffinityTag)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Conf) DeepCopyInto(out *Conf) {
	*out = *in
	if in.LocalityAwareness != nil {
		in, out := &in.LocalityAwareness, &out.LocalityAwareness
		*out = new(LocalityAwareness)
		(*in).DeepCopyInto(*out)
	}
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(LoadBalancer)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Conf.
func (in *Conf) DeepCopy() *Conf {
	if in == nil {
		return nil
	}
	out := new(Conf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Connection) DeepCopyInto(out *Connection) {
	*out = *in
	if in.SourceIP != nil {
		in, out := &in.SourceIP, &out.SourceIP
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Connection.
func (in *Connection) DeepCopy() *Connection {
	if in == nil {
		return nil
	}
	out := new(Connection)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Cookie) DeepCopyInto(out *Cookie) {
	*out = *in
	if in.TTL != nil {
		in, out := &in.TTL, &out.TTL
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Cookie.
func (in *Cookie) DeepCopy() *Cookie {
	if in == nil {
		return nil
	}
	out := new(Cookie)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CrossZone) DeepCopyInto(out *CrossZone) {
	*out = *in
	if in.Failover != nil {
		in, out := &in.Failover, &out.Failover
		*out = make([]Failover, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.FailoverThreshold != nil {
		in, out := &in.FailoverThreshold, &out.FailoverThreshold
		*out = new(FailoverThreshold)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CrossZone.
func (in *CrossZone) DeepCopy() *CrossZone {
	if in == nil {
		return nil
	}
	out := new(CrossZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Failover) DeepCopyInto(out *Failover) {
	*out = *in
	if in.From != nil {
		in, out := &in.From, &out.From
		*out = new(FromZone)
		(*in).DeepCopyInto(*out)
	}
	in.To.DeepCopyInto(&out.To)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Failover.
func (in *Failover) DeepCopy() *Failover {
	if in == nil {
		return nil
	}
	out := new(Failover)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FailoverThreshold) DeepCopyInto(out *FailoverThreshold) {
	*out = *in
	out.Percentage = in.Percentage
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FailoverThreshold.
func (in *FailoverThreshold) DeepCopy() *FailoverThreshold {
	if in == nil {
		return nil
	}
	out := new(FailoverThreshold)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FilterState) DeepCopyInto(out *FilterState) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FilterState.
func (in *FilterState) DeepCopy() *FilterState {
	if in == nil {
		return nil
	}
	out := new(FilterState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FromZone) DeepCopyInto(out *FromZone) {
	*out = *in
	if in.Zones != nil {
		in, out := &in.Zones, &out.Zones
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FromZone.
func (in *FromZone) DeepCopy() *FromZone {
	if in == nil {
		return nil
	}
	out := new(FromZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HashPolicy) DeepCopyInto(out *HashPolicy) {
	*out = *in
	if in.Terminal != nil {
		in, out := &in.Terminal, &out.Terminal
		*out = new(bool)
		**out = **in
	}
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = new(Header)
		**out = **in
	}
	if in.Cookie != nil {
		in, out := &in.Cookie, &out.Cookie
		*out = new(Cookie)
		(*in).DeepCopyInto(*out)
	}
	if in.Connection != nil {
		in, out := &in.Connection, &out.Connection
		*out = new(Connection)
		(*in).DeepCopyInto(*out)
	}
	if in.QueryParameter != nil {
		in, out := &in.QueryParameter, &out.QueryParameter
		*out = new(QueryParameter)
		**out = **in
	}
	if in.FilterState != nil {
		in, out := &in.FilterState, &out.FilterState
		*out = new(FilterState)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HashPolicy.
func (in *HashPolicy) DeepCopy() *HashPolicy {
	if in == nil {
		return nil
	}
	out := new(HashPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Header) DeepCopyInto(out *Header) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Header.
func (in *Header) DeepCopy() *Header {
	if in == nil {
		return nil
	}
	out := new(Header)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LeastRequest) DeepCopyInto(out *LeastRequest) {
	*out = *in
	if in.ChoiceCount != nil {
		in, out := &in.ChoiceCount, &out.ChoiceCount
		*out = new(uint32)
		**out = **in
	}
	if in.ActiveRequestBias != nil {
		in, out := &in.ActiveRequestBias, &out.ActiveRequestBias
		*out = new(intstr.IntOrString)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LeastRequest.
func (in *LeastRequest) DeepCopy() *LeastRequest {
	if in == nil {
		return nil
	}
	out := new(LeastRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancer) DeepCopyInto(out *LoadBalancer) {
	*out = *in
	if in.RoundRobin != nil {
		in, out := &in.RoundRobin, &out.RoundRobin
		*out = new(RoundRobin)
		**out = **in
	}
	if in.LeastRequest != nil {
		in, out := &in.LeastRequest, &out.LeastRequest
		*out = new(LeastRequest)
		(*in).DeepCopyInto(*out)
	}
	if in.RingHash != nil {
		in, out := &in.RingHash, &out.RingHash
		*out = new(RingHash)
		(*in).DeepCopyInto(*out)
	}
	if in.Random != nil {
		in, out := &in.Random, &out.Random
		*out = new(Random)
		**out = **in
	}
	if in.Maglev != nil {
		in, out := &in.Maglev, &out.Maglev
		*out = new(Maglev)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancer.
func (in *LoadBalancer) DeepCopy() *LoadBalancer {
	if in == nil {
		return nil
	}
	out := new(LoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalZone) DeepCopyInto(out *LocalZone) {
	*out = *in
	if in.AffinityTags != nil {
		in, out := &in.AffinityTags, &out.AffinityTags
		*out = new([]AffinityTag)
		if **in != nil {
			in, out := *in, *out
			*out = make([]AffinityTag, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalZone.
func (in *LocalZone) DeepCopy() *LocalZone {
	if in == nil {
		return nil
	}
	out := new(LocalZone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalityAwareness) DeepCopyInto(out *LocalityAwareness) {
	*out = *in
	if in.Disabled != nil {
		in, out := &in.Disabled, &out.Disabled
		*out = new(bool)
		**out = **in
	}
	if in.LocalZone != nil {
		in, out := &in.LocalZone, &out.LocalZone
		*out = new(LocalZone)
		(*in).DeepCopyInto(*out)
	}
	if in.CrossZone != nil {
		in, out := &in.CrossZone, &out.CrossZone
		*out = new(CrossZone)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalityAwareness.
func (in *LocalityAwareness) DeepCopy() *LocalityAwareness {
	if in == nil {
		return nil
	}
	out := new(LocalityAwareness)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Maglev) DeepCopyInto(out *Maglev) {
	*out = *in
	if in.TableSize != nil {
		in, out := &in.TableSize, &out.TableSize
		*out = new(uint32)
		**out = **in
	}
	if in.HashPolicies != nil {
		in, out := &in.HashPolicies, &out.HashPolicies
		*out = new([]HashPolicy)
		if **in != nil {
			in, out := *in, *out
			*out = make([]HashPolicy, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Maglev.
func (in *Maglev) DeepCopy() *Maglev {
	if in == nil {
		return nil
	}
	out := new(Maglev)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshLoadBalancingStrategy) DeepCopyInto(out *MeshLoadBalancingStrategy) {
	*out = *in
	if in.TargetRef != nil {
		in, out := &in.TargetRef, &out.TargetRef
		*out = new(commonv1alpha1.TargetRef)
		(*in).DeepCopyInto(*out)
	}
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = new([]To)
		if **in != nil {
			in, out := *in, *out
			*out = make([]To, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshLoadBalancingStrategy.
func (in *MeshLoadBalancingStrategy) DeepCopy() *MeshLoadBalancingStrategy {
	if in == nil {
		return nil
	}
	out := new(MeshLoadBalancingStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *QueryParameter) DeepCopyInto(out *QueryParameter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new QueryParameter.
func (in *QueryParameter) DeepCopy() *QueryParameter {
	if in == nil {
		return nil
	}
	out := new(QueryParameter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Random) DeepCopyInto(out *Random) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Random.
func (in *Random) DeepCopy() *Random {
	if in == nil {
		return nil
	}
	out := new(Random)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RingHash) DeepCopyInto(out *RingHash) {
	*out = *in
	if in.HashFunction != nil {
		in, out := &in.HashFunction, &out.HashFunction
		*out = new(HashFunctionType)
		**out = **in
	}
	if in.MinRingSize != nil {
		in, out := &in.MinRingSize, &out.MinRingSize
		*out = new(uint32)
		**out = **in
	}
	if in.MaxRingSize != nil {
		in, out := &in.MaxRingSize, &out.MaxRingSize
		*out = new(uint32)
		**out = **in
	}
	if in.HashPolicies != nil {
		in, out := &in.HashPolicies, &out.HashPolicies
		*out = new([]HashPolicy)
		if **in != nil {
			in, out := *in, *out
			*out = make([]HashPolicy, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RingHash.
func (in *RingHash) DeepCopy() *RingHash {
	if in == nil {
		return nil
	}
	out := new(RingHash)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoundRobin) DeepCopyInto(out *RoundRobin) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoundRobin.
func (in *RoundRobin) DeepCopy() *RoundRobin {
	if in == nil {
		return nil
	}
	out := new(RoundRobin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *To) DeepCopyInto(out *To) {
	*out = *in
	in.TargetRef.DeepCopyInto(&out.TargetRef)
	in.Default.DeepCopyInto(&out.Default)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new To.
func (in *To) DeepCopy() *To {
	if in == nil {
		return nil
	}
	out := new(To)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ToZone) DeepCopyInto(out *ToZone) {
	*out = *in
	if in.Zones != nil {
		in, out := &in.Zones, &out.Zones
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ToZone.
func (in *ToZone) DeepCopy() *ToZone {
	if in == nil {
		return nil
	}
	out := new(ToZone)
	in.DeepCopyInto(out)
	return out
}
