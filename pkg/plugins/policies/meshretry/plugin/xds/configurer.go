package xds

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	envoy_host_meta "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/host/omit_host_metadata/v3"
	envoy_prev_hosts "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/host/previous_hosts/v3"
	envoy_prev_priority "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/priority/previous_priorities/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_meta "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

const (
	HttpRetryOnDefault = "gateway-error,connect-failure," +
		"refused-stream"
	HttpRetryOnRetriableStatusCodes = "retriable-status-codes"
	GrpcRetryOnDefault              = "cancelled,connect-failure," +
		"gateway-error,refused-stream,reset,resource-exhausted,unavailable"
)

// DeprecatedConfigurer should be only used for configuring old MeshService outbounds.
// It should be removed after we stop using kuma.io/service tag, and move fully to new MeshService
// Deprecated
type DeprecatedConfigurer struct {
	Element  core_rules.Element
	Rules    core_rules.Rules
	Protocol core_mesh.Protocol
}

func (c *DeprecatedConfigurer) ConfigureListener(ln *envoy_listener.Listener) error {
	conf := c.getConf(c.Element)

	for _, fc := range ln.FilterChains {
		if c.Protocol == core_mesh.ProtocolTCP && conf != nil && conf.TCP != nil && conf.TCP.MaxConnectAttempt != nil {
			return v3.UpdateTCPProxy(fc, func(proxy *envoy_tcp.TcpProxy) error {
				proxy.MaxConnectAttempts = util_proto.UInt32(*conf.TCP.MaxConnectAttempt)
				return nil
			})
		} else {
			return v3.UpdateHTTPConnectionManager(fc, func(manager *envoy_hcm.HttpConnectionManager) error {
				return c.ConfigureRoute(manager.GetRouteConfig())
			})
		}
	}

	return nil
}

func (c *DeprecatedConfigurer) ConfigureRoute(route *envoy_route.RouteConfiguration) error {
	if route == nil {
		return nil
	}

	defaultPolicy, err := getRouteRetryConfig(c.getConf(c.Element), c.Protocol)
	if err != nil {
		return err
	}
	for _, virtualHost := range route.VirtualHosts {
		for _, route := range virtualHost.Routes {
			routeConfig := c.getConf(c.Element.WithKeyValue(core_rules.RuleMatchesHashTag, route.Name))
			policy, err := getRouteRetryConfig(routeConfig, c.Protocol)
			if err != nil {
				return err
			}
			if policy == nil {
				if defaultPolicy == nil {
					continue
				}
				policy = defaultPolicy
			}
			switch a := route.GetAction().(type) {
			case *envoy_route.Route_Route:
				a.Route.RetryPolicy = policy
			}
		}
	}

	return nil
}

func getRouteRetryConfig(conf *api.Conf, protocol core_mesh.Protocol) (*envoy_route.RetryPolicy, error) {
	if conf == nil {
		return nil, nil
	}
	switch protocol {
	case "http":
		return genHttpRetryPolicy(conf.HTTP)
	case "grpc":
		return genGrpcRetryPolicy(conf.GRPC)
	default:
		return nil, nil
	}
}

func GrpcRetryOn(conf *[]api.GRPCRetryOn) string {
	if conf == nil || len(*conf) == 0 {
		return ""
	}
	var retryOn []string

	for _, item := range *conf {
		// As `retryOn` is an enum value, and we use Kubernetes PascalCase convention but Envoy is using hyphens,
		// so we need to convert it
		retryOn = append(retryOn, api.GrpcRetryOnEnumToEnvoyValue[item])
	}

	return strings.Join(retryOn, ",")
}

func genGrpcRetryPolicy(conf *api.GRPC) (*envoy_route.RetryPolicy, error) {
	if conf == nil {
		return nil, nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn: GrpcRetryOnDefault,
	}

	if conf.PerTryTimeout != nil {
		policy.PerTryTimeout = util_proto.Duration(conf.PerTryTimeout.Duration)
	}

	if conf.NumRetries != nil {
		if *conf.NumRetries == 0 { // If numRetries is 0 just don't configure retries
			return nil, nil
		}
		policy.NumRetries = util_proto.UInt32(*conf.NumRetries)
	}

	retryOn := GrpcRetryOn(conf.RetryOn)
	if retryOn != "" {
		policy.RetryOn = retryOn
	}

	if conf.BackOff != nil {
		policy.RetryBackOff = &envoy_route.RetryPolicy_RetryBackOff{}

		if conf.BackOff.BaseInterval != nil {
			policy.RetryBackOff.BaseInterval = util_proto.Duration(conf.BackOff.BaseInterval.Duration)
		}
		if conf.BackOff.MaxInterval != nil {
			policy.RetryBackOff.MaxInterval = util_proto.Duration(conf.BackOff.MaxInterval.Duration)
		}
	}

	if conf.RateLimitedBackOff != nil {
		policy.RateLimitedRetryBackOff = configureRateLimitedRetryBackOff(conf.RateLimitedBackOff)
	}

	return &policy, nil
}

func genHttpRetryPolicy(conf *api.HTTP) (*envoy_route.RetryPolicy, error) {
	if conf == nil {
		return nil, nil
	}

	policy := envoy_route.RetryPolicy{
		RetryOn: HttpRetryOnDefault,
	}

	if conf.PerTryTimeout != nil {
		policy.PerTryTimeout = util_proto.Duration(conf.PerTryTimeout.Duration)
	}

	if conf.NumRetries != nil {
		if *conf.NumRetries == 0 { // If numRetries is 0 just don't configure retries
			return nil, nil
		}
		policy.NumRetries = util_proto.UInt32(*conf.NumRetries)
	}

	if conf.BackOff != nil {
		retryBackOff := &envoy_route.RetryPolicy_RetryBackOff{}

		if conf.BackOff.BaseInterval != nil {
			retryBackOff.BaseInterval = util_proto.Duration(conf.BackOff.BaseInterval.Duration)
		}

		if conf.BackOff.MaxInterval != nil {
			retryBackOff.MaxInterval = util_proto.Duration(conf.BackOff.MaxInterval.Duration)
		}

		policy.RetryBackOff = retryBackOff
	}

	if conf.RateLimitedBackOff != nil {
		policy.RateLimitedRetryBackOff = configureRateLimitedRetryBackOff(conf.RateLimitedBackOff)
	}

	retryOn, retriableStatusCodes, retriableMethods := splitRetryOn(conf.RetryOn)
	if retryOn != "" {
		policy.RetryOn = retryOn
	}
	if len(retriableStatusCodes) != 0 {
		policy.RetriableStatusCodes = retriableStatusCodes
		policy.RetryOn = ensureRetriableStatusCodes(policy.RetryOn)
	}

	for _, method := range retriableMethods {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: method,
			},
		}
		policy.RetriableRequestHeaders = append(policy.RetriableRequestHeaders,
			&envoy_route.HeaderMatcher{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
				InvertMatch: false,
			})
	}

	if conf.RetriableRequestHeaders != nil {
		for _, requestHeader := range *conf.RetriableRequestHeaders {
			policy.RetriableRequestHeaders = append(policy.RetriableRequestHeaders, headerMatcher(requestHeader))
		}
	}

	if conf.RetriableResponseHeaders != nil {
		for _, responseHeader := range *conf.RetriableResponseHeaders {
			policy.RetriableHeaders = append(policy.RetriableHeaders, headerMatcher(responseHeader))
		}
	}

	if conf.HostSelection != nil {
		if err := configureHostSelectionPredicates(conf.HostSelection, &policy); err != nil {
			return nil, err
		}
	}

	if conf.HostSelectionMaxAttempts != nil {
		policy.HostSelectionRetryMaxAttempts = *conf.HostSelectionMaxAttempts
	}

	return &policy, nil
}

func configureRateLimitedRetryBackOff(rateLimitedBackOff *api.RateLimitedBackOff) *envoy_route.RetryPolicy_RateLimitedRetryBackOff {
	rateLimitedRetryBackoff := &envoy_route.RetryPolicy_RateLimitedRetryBackOff{
		ResetHeaders: []*envoy_route.RetryPolicy_ResetHeader{},
	}
	for _, resetHeader := range pointer.Deref(rateLimitedBackOff.ResetHeaders) {
		rateLimitedRetryBackoff.ResetHeaders = append(rateLimitedRetryBackoff.ResetHeaders, &envoy_route.RetryPolicy_ResetHeader{
			Name:   string(resetHeader.Name),
			Format: api.RateLimitFormatEnumToEnvoyValue[resetHeader.Format],
		})
	}

	if rateLimitedBackOff.MaxInterval != nil {
		rateLimitedRetryBackoff.MaxInterval = util_proto.Duration(rateLimitedBackOff.MaxInterval.Duration)
	}

	return rateLimitedRetryBackoff
}

func configureHostSelectionPredicates(hostSelection *[]api.Predicate, policy *envoy_route.RetryPolicy) error {
	var retryHostPredicates []*envoy_route.RetryPolicy_RetryHostPredicate
	for _, hostSelect := range pointer.Deref(hostSelection) {
		switch hostSelect.PredicateType {
		case api.OmitPreviousHosts:
			prevHosts, err := util_proto.MarshalAnyDeterministic(
				&envoy_prev_hosts.PreviousHostsPredicate{},
			)
			if err != nil {
				return err
			}
			retryHostPredicates = append(retryHostPredicates,
				&envoy_route.RetryPolicy_RetryHostPredicate{
					Name: "envoy.retry_host_predicates.previous_hosts",
					ConfigType: &envoy_route.RetryPolicy_RetryHostPredicate_TypedConfig{
						TypedConfig: prevHosts,
					},
				},
			)
		case api.OmitHostsWithTags:
			taggedHosts, err := util_proto.MarshalAnyDeterministic(
				&envoy_host_meta.OmitHostMetadataConfig{
					MetadataMatch: envoy_meta.LbMetadata(hostSelect.Tags),
				},
			)
			if err != nil {
				return err
			}
			retryHostPredicates = append(retryHostPredicates,
				&envoy_route.RetryPolicy_RetryHostPredicate{
					Name: "envoy.retry_host_predicates.omit_host_metadata",
					ConfigType: &envoy_route.RetryPolicy_RetryHostPredicate_TypedConfig{
						TypedConfig: taggedHosts,
					},
				},
			)
		case api.OmitPreviousPriorities:
			prevPriorities, err := util_proto.MarshalAnyDeterministic(
				&envoy_prev_priority.PreviousPrioritiesConfig{
					UpdateFrequency: hostSelect.UpdateFrequency,
				},
			)
			if err != nil {
				return err
			}
			policy.RetryPriority = &envoy_route.RetryPolicy_RetryPriority{
				Name: "envoy.retry_priorities.previous_priorities",
				ConfigType: &envoy_route.RetryPolicy_RetryPriority_TypedConfig{
					TypedConfig: prevPriorities,
				},
			}
		}
	}
	policy.RetryHostPredicate = retryHostPredicates
	return nil
}

func headerMatcher(header common_api.HeaderMatch) *envoy_route.HeaderMatcher {
	matcher := &envoy_route.HeaderMatcher{
		Name:        string(header.Name),
		InvertMatch: false,
	}
	t := common_api.HeaderMatchExact
	if header.Type != nil {
		t = *header.Type
	}

	switch t {
	case common_api.HeaderMatchRegularExpression:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_StringMatch{
			StringMatch: &envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_SafeRegex{
					SafeRegex: &envoy_type_matcher.RegexMatcher{
						Regex: string(header.Value),
					},
				},
			},
		}
	case common_api.HeaderMatchPrefix:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_StringMatch{
			StringMatch: &envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
					Prefix: string(header.Value),
				},
			},
		}
	case common_api.HeaderMatchExact:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_StringMatch{
			StringMatch: &envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
					Exact: string(header.Value),
				},
			},
		}
	case common_api.HeaderMatchPresent:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_PresentMatch{
			PresentMatch: true,
		}
	case common_api.HeaderMatchAbsent:
		matcher.HeaderMatchSpecifier = &envoy_route.HeaderMatcher_PresentMatch{
			PresentMatch: false,
		}
	}

	return matcher
}

func splitRetryOn(conf *[]api.HTTPRetryOn) (string, []uint32, []string) {
	if conf == nil {
		return "", []uint32{}, nil
	}
	var retryOn []string
	var retriableStatusCodes []uint32
	var retriableMethods []string

	for _, item := range *conf {
		key := string(item)
		statusCode, err := strconv.ParseUint(key, 10, 32)
		switch {
		case err == nil && statusCode <= math.MaxInt && http.StatusText(int(statusCode)) != "":
			retriableStatusCodes = append(retriableStatusCodes, uint32(statusCode))
		case strings.HasPrefix(key, string(api.HttpMethodPrefix)):
			retriableMethods = append(retriableMethods, api.HttpRetryOnEnumToEnvoyValue[item])
		default:
			// As `retryOn` is an enum value, and we use Kubernetes PascalCase convention but Envoy is using hyphens,
			// so we need to convert it
			retryOn = append(retryOn, api.HttpRetryOnEnumToEnvoyValue[item])
		}
	}
	return strings.Join(retryOn, ","), retriableStatusCodes, retriableMethods
}

func ensureRetriableStatusCodes(policyRetryOn string) string {
	policyRetrySplit := strings.Split(policyRetryOn, ",")
	seenRetriable := false
	for _, r := range policyRetrySplit {
		if r == HttpRetryOnRetriableStatusCodes {
			seenRetriable = true
			break
		}
	}
	if !seenRetriable {
		policyRetrySplit = append(policyRetrySplit, HttpRetryOnRetriableStatusCodes)
	}
	policyRetryOn = strings.Join(policyRetrySplit, ",")
	return policyRetryOn
}

func (c *DeprecatedConfigurer) getConf(element core_rules.Element) *api.Conf {
	if c.Rules == nil {
		return nil
	}
	return core_rules.ComputeConf[api.Conf](c.Rules, element)
}

type Configurer struct {
	Conf     api.Conf
	Protocol core_mesh.Protocol
}

func (c *Configurer) ConfigureListener(listener *envoy_listener.Listener) error {
	for _, fc := range listener.FilterChains {
		if c.Protocol == core_mesh.ProtocolTCP && c.Conf.TCP != nil && c.Conf.TCP.MaxConnectAttempt != nil {
			return v3.UpdateTCPProxy(fc, func(proxy *envoy_tcp.TcpProxy) error {
				proxy.MaxConnectAttempts = util_proto.UInt32(*c.Conf.TCP.MaxConnectAttempt)
				return nil
			})
		} else {
			return v3.UpdateHTTPConnectionManager(fc, func(manager *envoy_hcm.HttpConnectionManager) error {
				return c.ConfigureRoute(manager.GetRouteConfig())
			})
		}
	}

	return nil
}

func (c *Configurer) ConfigureRoute(route *envoy_route.RouteConfiguration) error {
	if route == nil {
		return nil
	}

	policy, err := getRouteRetryConfig(&c.Conf, c.Protocol)
	if err != nil {
		return err
	}
	if policy == nil {
		return nil
	}
	for _, virtualHost := range route.VirtualHosts {
		for _, route := range virtualHost.Routes {
			switch a := route.GetAction().(type) {
			case *envoy_route.Route_Route:
				a.Route.RetryPolicy = policy
			}
		}
	}

	return nil
}
