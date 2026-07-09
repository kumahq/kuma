package kuma_cp_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
)

var _ = Describe("MeshServiceConfig Validate", func() {
	type testCase struct {
		cfg     kuma_cp.MeshServiceLabelPropagation
		wantErr string
	}

	DescribeTable("LabelPropagation validation",
		func(given testCase) {
			cfg := kuma_cp.MeshServiceConfig{
				LabelPropagation: given.cfg,
			}
			err := cfg.Validate()
			if given.wantErr != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(given.wantErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		},
		Entry("disabled with empty allow-list", testCase{
			cfg: kuma_cp.MeshServiceLabelPropagation{
				Enabled:          false,
				AllowedLabelKeys: []string{},
			},
		}),
		Entry("enabled with non-reserved keys", testCase{
			cfg: kuma_cp.MeshServiceLabelPropagation{
				Enabled:          true,
				AllowedLabelKeys: []string{"app", "version", "team"},
			},
		}),
		Entry("rejects kuma.io/ reserved key", testCase{
			cfg: kuma_cp.MeshServiceLabelPropagation{
				Enabled:          true,
				AllowedLabelKeys: []string{"app", "kuma.io/service"},
			},
			wantErr: `reserved key "kuma.io/service"`,
		}),
		Entry("rejects k8s.kuma.io/ reserved key", testCase{
			cfg: kuma_cp.MeshServiceLabelPropagation{
				Enabled:          false,
				AllowedLabelKeys: []string{"k8s.kuma.io/namespace"},
			},
			wantErr: `reserved key "k8s.kuma.io/namespace"`,
		}),
		Entry("enabled with nil allow-list", testCase{
			cfg: kuma_cp.MeshServiceLabelPropagation{
				Enabled:          true,
				AllowedLabelKeys: nil,
			},
		}),
	)
})
