package datasource_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("DataSource Loader", func() {
	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secrets := []*system.SecretResource{
			{
				Meta: &model.ResourceMeta{
					Mesh: "default",
					Name: "test-secret",
				},
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("abc")),
				},
			},
		}
		dataSourceLoader = datasource.NewStaticLoader(secrets)
	})

	Context("Secret", func() {
		It("should load secret", func() {
			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_Secret{
					Secret: "test-secret",
				},
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})

		It("should throw an error when secret is not found", func() {
			// when
			_, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_Secret{
					Secret: "test-secret-2",
				},
			})

			// then
			Expect(err).To(MatchError(`could not load data: resource not found: type="Secret" name="test-secret-2" mesh="default"`))
		})
	})

	Context("Inline", func() {
		It("should load from inline", func() {
			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_Inline{
					Inline: util_proto.Bytes([]byte("abc")),
				},
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})
	})

	Context("Inline string", func() {
		It("should load from inline string", func() {
			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_InlineString{
					InlineString: "abc",
				},
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})
	})
})
