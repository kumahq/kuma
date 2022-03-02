package datasource_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secret_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("DataSource Loader", func() {

	var secretManager manager.ResourceManager
	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager = secret_manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None(), nil)
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})

	Context("Secret", func() {
		It("should load secret", func() {
			// given
			secretResource := system.SecretResource{
				Spec: &system_proto.Secret{
					Data: util_proto.Bytes([]byte("abc")),
				},
			}
			err := secretManager.Create(context.Background(), &secretResource, store.CreateByKey("test-secret", "default"))
			Expect(err).ToNot(HaveOccurred())

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
					Secret: "test-secret",
				},
			})

			// then
			Expect(err).To(MatchError(`could not load data: Resource not found: type="Secret" name="test-secret" mesh="default"`))
		})
	})

	Context("File", func() {
		It("should load from file", func() {
			// given
			file, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(file.Name(), []byte("abc"), os.ModeAppend)
			Expect(err).ToNot(HaveOccurred())

			// when
			data, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: file.Name(),
				},
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})

		It("should throw an error on problems with loading from file", func() {
			// when
			_, err := dataSourceLoader.Load(context.Background(), "default", &system_proto.DataSource{
				Type: &system_proto.DataSource_File{
					File: "non-existent-file",
				},
			})

			// then
			Expect(err).To(MatchError("could not load data: open non-existent-file: no such file or directory"))
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
