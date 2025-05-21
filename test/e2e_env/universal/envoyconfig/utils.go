package envoyconfig

import (
	"encoding/json"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"regexp"
	"slices"
	"strings"

	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func waitMeshServiceReady(mesh, name string) {
	Eventually(func(g Gomega) {
		spec, status, err := GetMeshServiceStatus(universal.Cluster, name, mesh)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(spec.Identities).To(Equal([]meshservice_api.MeshServiceIdentity{
			{
				Type:  meshservice_api.MeshServiceIdentityServiceTagType,
				Value: name,
			},
		}))
		g.Expect(status.TLS.Status).To(Equal(meshservice_api.TLSReady))
	}, "30s", "1s").Should(Succeed())
}

func getConfig(mesh, dpp string) string {
	output, err := universal.Cluster.GetKumactlOptions().
		RunKumactlAndGetOutput("inspect", "dataplane", dpp, "--type", "config", "--mesh", mesh, "--shadow", "--include=diff")
	Expect(err).ToNot(HaveOccurred())
	redacted := redactDnsLookupFamily(
		redactStatPrefixes(
			redactIPs(
				redactKumaDynamicConfig(output),
			),
		),
	)

	response := types.GetDataplaneXDSConfigResponse{}
	Expect(json.Unmarshal([]byte(redacted), &response)).To(Succeed())
	Expect(response.Diff).ToNot(BeNil())
	response.Diff = pointer.To(slices.DeleteFunc(*response.Diff, func(item api_common.JsonPatchItem) bool {
		return item.Op == api_common.Test
	}))
	slices.SortStableFunc(*response.Diff, func(a, b api_common.JsonPatchItem) int {
		return strings.Compare(a.Path, b.Path)
	})

	result, err := json.MarshalIndent(response, "", "  ")
	Expect(err).ToNot(HaveOccurred())
	return string(result)
}

var ipv6Regex = `\[?` + // Optional opening square bracket for IPv6 in URLs (e.g., [2001:db8::1])
	`(` +
	// Full IPv6 address with 8 segments (e.g., 2001:0db8:85a3:0000:0000:8a2e:0370:7334)
	`([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|` +
	// IPv6 with leading compression (e.g., ::1, ::8a2e:0370:7334)
	`([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|` +
	// IPv6 with trailing compression (e.g., 2001:db8::, 2001:db8::1:2)
	`([0-9a-fA-F]{1,4}:){1,7}:|` +
	// IPv6 with mixed compression (e.g., 2001:db8:0:0::1:2)
	`([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|` +
	`([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|` +
	`([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|` +
	`([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|` +
	// IPv6 with only one segment and compression (e.g., 2001::1:2:3:4:5:6)
	`[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|` +
	// Fully compressed IPv6 (::) or with trailing segments (::1, ::8a2e:0370:7334)
	`:((:[0-9a-fA-F]{1,4}){1,7}|:)|` +
	// Link-local IPv6 with zone identifiers (e.g., fe80::1%eth0)
	`fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|` +
	// IPv6 with embedded IPv4 (e.g., ::ffff:192.168.1.1)
	`::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}` +
	`(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|` +
	// Mixed IPv6 and IPv4 (e.g., 2001:db8::192.168.1.1)
	`([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}` +
	`(25[0-5]|(2[0-4]|1?[0-9])?[0-9])` +
	`)` +
	`]?` // Optional closing square bracket for IPv6 in URLs (e.g., [2001:db8::1])

var ipv4Regex = `\b` + // Word boundary to ensure we match standalone IPv4 addresses
	`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}` + // Matches IPv4 format (e.g., 192.168.0.1)
	`\b`

var ipRegex = regexp.MustCompile(ipv4Regex + "|" + ipv6Regex)

func redactIPs(jsonStr string) string {
	return ipRegex.ReplaceAllString(jsonStr, "IP_REDACTED")
}

// TODO this should be removed after fixing: https://github.com/kumahq/kuma/issues/12733
var statsPrefixRegex = regexp.MustCompile(`"statPrefix":[[:space:]]*"[^"]*"`)

func redactStatPrefixes(jsonStr string) string {
	return statsPrefixRegex.ReplaceAllString(jsonStr, "\"statPrefix\": \"STAT_PREFIX_REDACTED\"")
}

var dnsLookupRegex = regexp.MustCompile(`,[[:space:]]*"dnsLookupFamily":[[:space:]]*"[^"]*"`)

// This needs to be removed as we run tests on ipv4 and ipv6. In ipv4 dnslookupFamily is set to V4_ONLY,
// and in the case of ipv6 this field is default, so it is missing in the config.
func redactDnsLookupFamily(jsonStr string) string {
	return dnsLookupRegex.ReplaceAllString(jsonStr, "")
}

var dynamicConfigJsonPatch = []byte(`[{ "op": "remove", "path": "/xds/type.googleapis.com~1envoy.config.listener.v3.Listener/_kuma:dynamicconfig" }]`)

// We can remove dynamic config as this contains dns config which changes in multiple places making it hard to mask
func redactKumaDynamicConfig(jsonStr string) string {
	patch, err := jsonpatch.DecodePatch(dynamicConfigJsonPatch)
	if err != nil {
		panic(err)
	}

	options := jsonpatch.NewApplyOptions()
	options.AllowMissingPathOnRemove = true
	modified, err := patch.ApplyWithOptions([]byte(jsonStr), options)
	if err != nil {
		panic(err)
	}

	return string(modified)
}

func cleanupAfterTest(mesh string, policies ...core_model.ResourceTypeDescriptor) func() {
	return func() {
		Expect(DeleteMeshResources(universal.Cluster, mesh, policies...)).To(Succeed())
	}
}
