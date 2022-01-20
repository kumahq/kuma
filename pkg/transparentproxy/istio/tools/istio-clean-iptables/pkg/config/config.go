// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"fmt"
)

// Command line options
// nolint: maligned
type Config struct {
	DryRun                    bool     `json:"DRY_RUN"`
	ProxyUID                  string   `json:"PROXY_UID"`
	ProxyGID                  string   `json:"PROXY_GID"`
	RedirectDNS               bool     `json:"REDIRECT_DNS"`
	RedirectAllDNSTraffic     bool     `json:"REDIRECT_ALL_DNS_TRAFFIC"`
	DNSServersV4              []string `json:"DNS_SERVERS_V4"`
	DNSServersV6              []string `json:"DNS_SERVERS_V6"`
	AgentDNSListenerPort      string   `json:"AGENT_DNS_LISTENER_PORT"`
	DNSUpstreamTargetChain    string   `json:"DNS_UPSTREAM_TARGET_CHAIN"`
	SkipDNSConntrackZoneSplit bool     `json:"SKIP_DNS_CONNTRACK_ZONE_SPLIT"`
}

func (c *Config) String() string {
	output, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		fmt.Printf("Unable to marshal config object: %v", err)
		panic(err)
	}
	return string(output)
}

func (c *Config) Print() {
	fmt.Println("Variables:")
	fmt.Println("----------")
	fmt.Printf("PROXY_UID=%s\n", c.ProxyUID)
	fmt.Printf("PROXY_GID=%s\n", c.ProxyGID)
	fmt.Printf("DNS_CAPTURE=%t\n", c.RedirectDNS)
	fmt.Printf("DNS_SERVERS=%s,%s\n", c.DNSServersV4, c.DNSServersV6)
	fmt.Printf("AGENT_DNS_LISTENER_PORT=%s\n", c.AgentDNSListenerPort)
	fmt.Printf("DNS_UPSTREAM_TARGET_CHAIN=%s\n", c.DNSUpstreamTargetChain)
	fmt.Printf("SKIP_DNS_CONNTRACK_ZONE_SPLIT=%t\n", c.SkipDNSConntrackZoneSplit)
	fmt.Println("")
}
