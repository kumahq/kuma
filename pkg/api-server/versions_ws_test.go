package api_server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/Masterminds/semver/v3"

	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

func buildConstraints(versions map[string]envoyVersion) ([]*semver.Constraints, error) {
	var errs []string
	var constraints []*semver.Constraints

	for c := range versions {
		constraint, err := semver.NewConstraint(c)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		constraints = append(constraints, constraint)
	}

	if len(errs) > 0 {
		return nil, errors.Errorf("couldn't build constraints:\n%s", strings.Join(errs, "\n"))
	}

	return constraints, nil
}

func getConstraint(constraints []*semver.Constraints, version string) (*semver.Constraints, error) {
	v, err := semver.NewVersion(version)
	if err != nil {
		return nil, err
	}

	var matchedConstrain []*semver.Constraints

	for _, constraint := range constraints {
		if constraint.Check(v) {
			matchedConstrain = append(matchedConstrain, constraint)
		}
	}

	if len(matchedConstrain) == 0 {
		return nil, errors.Errorf("no constraints for version: %s found", version)
	}

	if len(matchedConstrain) > 1 {
		var matched []string
		for _, c := range matchedConstrain {
			matched = append(matched, c.String())
		}

		return nil, errors.Errorf(
			"more than one constraint for version: %s\n%s",
			version,
			strings.Join(matched, "\n"),
		)
	}

	return matchedConstrain[0], nil
}

func validateConstrainForEnvoy(constrain string, version string) error {
	if constrain != version {
		return errors.Errorf("envoy version in constrain: %s doesn't equal expected one: %s", constrain, version)
	}

	return nil
}

type envoyVersion struct {
	Envoy string
}

type versions struct {
	KumaDp map[string]envoyVersion
}

var _ = Describe("Versions WS", func() {
	It("should return the supported versions", func() {
		// setup
		resourceStore := memory.NewStore()
		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())
		apiServer := createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()

		// wait for the server
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://%s/versions", apiServer.Address()))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://%s/versions", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		var data versions

		Expect(json.NewDecoder(resp.Body).Decode(&data)).ToNot(HaveOccurred())

		constraints, err := buildConstraints(data.KumaDp)
		Expect(err).ToNot(HaveOccurred())

		Expect(data).ToNot(BeNil())
		Expect(data.KumaDp).ToNot(BeNil())

		// 1.0.0
		constrain, err := getConstraint(constraints, "1.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.0")).To(Succeed())

		// 1.0.1
		constrain, err = getConstraint(constraints, "1.0.1")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.0")).To(Succeed())

		// 1.0.2
		constrain, err = getConstraint(constraints, "1.0.2")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.1")).To(Succeed())

		// 1.0.3
		constrain, err = getConstraint(constraints, "1.0.3")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.1")).To(Succeed())

		// 1.0.4
		constrain, err = getConstraint(constraints, "1.0.4")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.1")).To(Succeed())

		// 1.0.5
		constrain, err = getConstraint(constraints, "1.0.5")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.2")).To(Succeed())

		// 1.0.6
		constrain, err = getConstraint(constraints, "1.0.6")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.2")).To(Succeed())

		// 1.0.7
		constrain, err = getConstraint(constraints, "1.0.7")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.2")).To(Succeed())

		// 1.0.8
		constrain, err = getConstraint(constraints, "1.0.8")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "1.16.2")).To(Succeed())

		// ~1.1.0
		constrain, err = getConstraint(constraints, "1.1.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "~1.17.0")).To(Succeed())

		// ~1.2.0
		constrain, err = getConstraint(constraints, "1.2.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(validateConstrainForEnvoy(data.KumaDp[constrain.String()].Envoy, "~1.18.0")).To(Succeed())
	})
})
