package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/kumahq/kuma/v3/api/openapi/types"
	kumactl_cmd "github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	kuma_version "github.com/kumahq/kuma/v3/pkg/version"
)

func TestCheckCompatibility(t *testing.T) {
	origBuild := kuma_version.Build
	origProduct := kuma_version.Product
	defer func() {
		kuma_version.Build = origBuild
		kuma_version.Product = origProduct
	}()

	serverVersion := func(version string) func() (*types.IndexResponse, error) {
		return func() (*types.IndexResponse, error) {
			return &types.IndexResponse{Product: "Kuma", Version: version}, nil
		}
	}

	clientVersion := func(version, product string) {
		kuma_version.Build = kuma_version.BuildInfo{Version: version, Product: product}
		kuma_version.Product = product
	}

	tests := []struct {
		name     string
		client   string
		server   func() (*types.IndexResponse, error)
		expected []string
	}{
		{
			name:   "preview server with higher major version warns",
			client: "2.14.5",
			server: serverVersion("3.0.0-preview1-abc123"),
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.0.0-preview1-abc123",
				"The server runs a newer major version than kumactl; fields unknown to this client are silently dropped when reading or applying resources",
			},
		},
		{
			name:     "preview server with same major version stays silent",
			client:   "3.0.0",
			server:   serverVersion("3.0.0-preview.7"),
			expected: []string{},
		},
		{
			name:     "preview server with unparseable dev client stays silent",
			client:   "unknown",
			server:   serverVersion("3.0.0-preview1-abc123"),
			expected: []string{},
		},
		{
			name:   "preview tag that is not semver still warns on higher major",
			client: "2.14.5",
			server: serverVersion("3.preview"),
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.preview",
				"The server runs a newer major version than kumactl",
			},
		},
		{
			name:   "non-preview server with higher major version warns about dropped fields",
			client: "2.14.5",
			server: serverVersion("3.0.0"),
			expected: []string{
				"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kuma for 3.0.0",
				"The server runs a newer major version than kumactl; fields unknown to this client are silently dropped when reading or applying resources",
			},
		},
		{
			name:     "same version and product stays silent",
			client:   "2.14.5",
			server:   serverVersion("2.14.5"),
			expected: []string{},
		},
		{
			name:   "same version but different product warns about mismatch",
			client: "2.14.5",
			server: func() (*types.IndexResponse, error) {
				return &types.IndexResponse{Product: "Kong Mesh", Version: "2.14.5"}, nil
			},
			expected: []string{"You are using kumactl version 2.14.5 for Kuma, but the server returned version: Kong Mesh for 2.14.5"},
		},
		{
			name:   "dev build client against a release server does not warn about dropped fields",
			client: "0.0.0-preview.vabc123",
			server: serverVersion("3.0.0"),
			expected: []string{
				"You are using kumactl version 0.0.0-preview.vabc123 for Kuma, but the server returned version: Kuma for 3.0.0",
			},
		},
		{
			name:   "failed version fetch warns and returns nil",
			client: "2.14.5",
			server: func() (*types.IndexResponse, error) { return nil, errors.New("boom") },
			expected: []string{
				"Failed to retrieve server version, can't check compatibility: boom",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientVersion(test.client, "Kuma")
			out := &bytes.Buffer{}

			result := kumactl_cmd.CheckCompatibility(test.server, out)

			for _, expected := range test.expected {
				if !strings.Contains(out.String(), expected) {
					t.Fatalf("expected output to contain %q, got %q", expected, out.String())
				}
			}
			switch test.name {
			case "failed version fetch warns and returns nil":
				if result != nil {
					t.Fatalf("expected nil result on fetch failure, got %v", result)
				}
			default:
				if result == nil {
					t.Fatal("expected non-nil result on successful fetch")
				}
			}
			if count := strings.Count(out.String(), "WARNING:"); count != len(test.expected) {
				fmt.Printf("got output: %q\n", out.String())
				t.Fatalf("expected %d warnings, got %d", len(test.expected), count)
			}
		})
	}
}
