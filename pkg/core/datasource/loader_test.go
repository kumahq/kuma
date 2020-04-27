package datasource_test

import (
	"context"
	"os"

	"io/ioutil"

	"github.com/golang/protobuf/ptypes/wrappers"

	system_proto "github.com/Kong/kuma/api/system/v1alpha1"
	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/secrets/cipher"
	"github.com/Kong/kuma/pkg/core/secrets/manager"
	secret_store "github.com/Kong/kuma/pkg/core/secrets/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DataSource Loader", func() {

	var secretManager manager.SecretManager
	var dataSourceLoader datasource.Loader

	BeforeEach(func() {
		secretManager = manager.NewSecretManager(secret_store.NewSecretStore(memory.NewStore()), cipher.None())
		dataSourceLoader = datasource.NewDataSourceLoader(secretManager)
	})

	Context("Secret", func() {
		It("should load secret", func() {
			// given
			secretResource := system.SecretResource{
				Spec: system_proto.Secret{
					Data: &wrappers.BytesValue{
						Value: []byte("abc"),
					},
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
			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(file.Name(), []byte("abc"), os.ModeAppend)
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
					Inline: &wrappers.BytesValue{
						Value: []byte("abc"),
					},
				},
			})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal([]byte("abc")))
		})
	})
})
