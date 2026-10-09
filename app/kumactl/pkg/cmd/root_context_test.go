package cmd_test

import (
	"bytes"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/api/openapi/types"
	kumactl_cmd "github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	kuma_version "github.com/kumahq/kuma/v3/pkg/version"
)

type checkCompatibilityCase struct {
	clientVersion string
	serverProduct string
	serverVersion string
	fetchErr      error
	expected      []string
	expectNil     bool
}

var _ = Describe("CheckCompatibility", func() {
	var backupBuild kuma_version.BuildInfo
	var backupProduct string
	var out *bytes.Buffer

	BeforeEach(func() {
		backupBuild = kuma_version.Build
		backupProduct = kuma_version.Product
		out = &bytes.Buffer{}
	})

	AfterEach(func() {
		kuma_version.Build = backupBuild
		kuma_version.Product = backupProduct
	})

	check := func(c checkCompatibilityCase) {
		kuma_version.Build = kuma_version.BuildInfo{Version: c.clientVersion, Product: "Kuma"}
		kuma_version.Product = "Kuma"

		var result *types.IndexResponse
		if c.fetchErr != nil {
			result = kumactl_cmd.CheckCompatibility(func() (*types.IndexResponse, error) {
				return nil, c.fetchErr
			}, out)
		} else {
			result = kumactl_cmd.CheckCompatibility(func() (*types.IndexResponse, error) {
				return &types.IndexResponse{Product: c.serverProduct, Version: c.serverVersion}, nil
			}, out)
		}

		if c.expectNil {
			Expect(result).To(BeNil())
		} else {
			Expect(result).NotTo(BeNil())
		}
		Expect(strings.Count(out.String(), "WARNING:")).To(Equal(len(c.expected)))
		for _, expected := range c.expected {
			Expect(out.String()).To(ContainSubstring(expected))
		}
	}

	DescribeTable("version compatibility warning",
		check,
		Entry("preview server with higher major version warns", checkCompatibilityCase{
			clientVersion: "2.14.5",
			serverProduct: "Kuma",
			serverVersion: "3.0.0-preview1-abc123",
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.0.0-preview1-abc123",
				"The server runs a newer major version than kumactl; fields unknown to this client are silently dropped when reading or applying resources",
			},
		}),
		Entry("preview server with same major version stays silent", checkCompatibilityCase{
			clientVersion: "3.0.0",
			serverProduct: "Kuma",
			serverVersion: "3.0.0-preview.7",
		}),
		Entry("preview server with unparseable dev client stays silent", checkCompatibilityCase{
			clientVersion: "unknown",
			serverProduct: "Kuma",
			serverVersion: "3.0.0-preview1-abc123",
		}),
		Entry("preview tag that is not semver still warns on higher major", checkCompatibilityCase{
			clientVersion: "2.14.5",
			serverProduct: "Kuma",
			serverVersion: "3.preview",
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.preview",
				"The server runs a newer major version than kumactl; fields unknown to this client are silently dropped when reading or applying resources",
			},
		}),
		Entry("non-preview server with higher major version warns about dropped fields", checkCompatibilityCase{
			clientVersion: "2.14.5",
			serverProduct: "Kuma",
			serverVersion: "3.0.0",
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.0.0",
				"The server runs a newer major version than kumactl; fields unknown to this client are silently dropped when reading or applying resources",
			},
		}),
		Entry("same version and product stays silent", checkCompatibilityCase{
			clientVersion: "2.14.5",
			serverProduct: "Kuma",
			serverVersion: "2.14.5",
		}),
		Entry("same version but different product warns about mismatch", checkCompatibilityCase{
			clientVersion: "2.14.5",
			serverProduct: "Kong Mesh",
			serverVersion: "2.14.5",
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kong Mesh for 2.14.5",
			},
		}),
		Entry("dev build client against a release server does not warn about dropped fields", checkCompatibilityCase{
			clientVersion: "0.0.0-preview.vabc123",
			serverProduct: "Kuma",
			serverVersion: "3.0.0",
			expected: []string{
				"You are using kumactl version 0.0.0-preview.vabc123 for Kuma, but the server returned version: Kuma for 3.0.0",
			},
		}),
		Entry("failed version fetch warns and returns nil", checkCompatibilityCase{
			clientVersion: "2.14.5",
			fetchErr:      errors.New("boom"),
			expected: []string{
				"Failed to retrieve server version, can't check compatibility: boom",
			},
			expectNil: true,
		}),
	)
})
