package datasource_test

import (
	"context"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/datasource"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/v3/pkg/test/resources/model"
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
					Data: system_proto.Bytes([]byte("abc")),
				},
			},
		}
		dataSourceLoader = datasource.NewStaticLoader(secrets)
	})

	Context("Secret", func() {
		It("should load secret", func() {
			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Secret: pointer.To("test-secret"),
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})

		It("should throw an error when secret is not found", func() {
			// when
			_, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Secret: pointer.To("test-secret-2"),
			})

			// then
			Expect(err).To(MatchError(`could not load data: resource not found: type="Secret" name="test-secret-2" mesh="default"`))
		})
	})

	Context("Inline", func() {
		It("should load from inline", func() {
			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Inline: system_proto.Bytes([]byte("abc")),
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
				InlineString: pointer.To("abc"),
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})
	})
})
