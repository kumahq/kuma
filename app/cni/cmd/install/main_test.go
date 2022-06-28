package main

import (
	"github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestTransformJsonConfig(t *testing.T) {
	// given
	g := gomega.NewWithT(t)
	kumaCniConfig := `{
      "type": "kuma-cni",
      "log_level": "info",
      "kubernetes": {
        "kubeconfig": "/etc/cni/net.d/ZZZ-kuma-cni-kubeconfig",
        "cni_bin_dir": "/opt/cni/bin",
        "exclude_namespaces": [
          "kuma-system"
        ]
      }
    }`
	calicoConfig, _ := ioutil.ReadFile("data/given/10-calico.conflist")
	expectedConfig, _ := ioutil.ReadFile("data/expected/10-calico.conflist")

	// when
	result, _ := transformJsonConfig(kumaCniConfig, calicoConfig)

	// then
	g.Expect(result).To(gomega.MatchJSON(expectedConfig))
}
