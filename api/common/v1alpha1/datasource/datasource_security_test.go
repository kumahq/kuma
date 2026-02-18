package datasource_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/api/common/v1alpha1/datasource"
)

func TestDataSourceSecurity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DataSource Security Suite")
}

var _ = Describe("SecureDataSource File Path Validation", func() {
	var (
		tmpDir string
		ctx    context.Context
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "datasource-test-*")
		Expect(err).ToNot(HaveOccurred())
		ctx = context.Background()
	})

	AfterEach(func() {
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Context("ReadByControlPlane with File datasource", func() {
		It("should reject relative paths", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "relative/path/to/file.txt",
				},
			}

			// when
			_, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file path must be absolute"))
		})

		It("should reject paths with directory traversal", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "/etc/../../../etc/passwd",
				},
			}

			// when
			_, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("directory traversal"))
		})

		It("should reject paths with .. sequences", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "/some/path/../../../etc/passwd",
				},
			}

			// when
			_, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("directory traversal"))
		})

		It("should reject empty paths", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "",
				},
			}

			// when
			_, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file path cannot be empty"))
		})

		It("should accept valid absolute paths", func() {
			// given
			testFile := filepath.Join(tmpDir, "test.txt")
			testContent := []byte("test content")
			err := os.WriteFile(testFile, testContent, 0o600)
			Expect(err).ToNot(HaveOccurred())

			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: testFile,
				},
			}

			// when
			content, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal(testContent))
		})

		It("should handle non-existent files gracefully", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "/absolutely/nonexistent/path/file.txt",
				},
			}

			// when
			_, err := sds.ReadByControlPlane(ctx, nil, "default")

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file"))
		})
	})

	Context("ValidateSecureDataSource with File datasource", func() {
		It("should reject relative paths during validation", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "relative/path.txt",
				},
			}

			// when
			verr := sds.ValidateSecureDataSource(nil)

			// then
			Expect(verr.HasViolations()).To(BeTrue())
			Expect(verr.Error()).To(ContainSubstring("file path must be absolute"))
		})

		It("should reject directory traversal during validation", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "/etc/../../../etc/passwd",
				},
			}

			// when
			verr := sds.ValidateSecureDataSource(nil)

			// then
			Expect(verr.HasViolations()).To(BeTrue())
			Expect(verr.Error()).To(ContainSubstring("directory traversal"))
		})

		It("should accept valid absolute paths during validation", func() {
			// given
			sds := &datasource.SecureDataSource{
				Type: datasource.SecureDataSourceFile,
				File: &datasource.File{
					Path: "/etc/kuma/certs/ca.crt",
				},
			}

			// when
			verr := sds.ValidateSecureDataSource(nil)

			// then
			Expect(verr.HasViolations()).To(BeFalse())
		})
	})
})
