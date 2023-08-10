package etcd

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/etcd"
	common_etcd "github.com/kumahq/kuma/pkg/plugins/common/etcd"
	util_files "github.com/kumahq/kuma/pkg/util/files"
)

type EtcdContainer struct {
	container testcontainers.Container
}

func (v *EtcdContainer) Start() error {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: resourceDir(),
		},

		ExposedPorts: []string{"2379/tcp"},
		WaitingFor:   wait.ForListeningPort("2379"),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}
	v.container = c
	return nil
}

func resourceDir() string {
	cwd, _ := os.Getwd()
	_, callerFile, _, _ := runtime.Caller(0)

	dir := findProjectRoot(cwd, callerFile)
	rDir := path.Join(dir, "tools/etcd")
	return rDir
}

func findProjectRoot(cwd, callerFile string) string {
	projectRootParent := util_files.GetProjectRootParent(cwd)
	fileRelativeToProjectRootParent := util_files.RelativeToProjectRootParent(callerFile)
	projectRoot := util_files.GetProjectRoot(cwd)
	fileRelativeToProjectRoot := util_files.RelativeToProjectRoot(callerFile)

	file := path.Join(projectRootParent, fileRelativeToProjectRootParent)
	if !util_files.FileExists(file) {
		file = path.Join(projectRoot, fileRelativeToProjectRoot)
	}
	if !util_files.FileExists(file) {
		file = path.Join(util_files.GetGopath(), "pkg", "mod", util_files.RelativeToPkgMod(callerFile))
	}

	return util_files.GetProjectRoot(file)
}

func (v *EtcdContainer) Stop() error {
	if v.container != nil {
		return v.container.Terminate(context.Background())
	}
	return nil
}

func (v *EtcdContainer) Config() (*etcd.EtcdConfig, error) {
	cfg := etcd.DefaultEtcdConfig()
	ctx := context.Background()
	ip, err := v.container.Host(ctx)
	if err != nil {
		return nil, err
	}
	port, err := v.container.MappedPort(ctx, "2379")
	if err != nil {
		return nil, err
	}
	cfg.Endpoints = []string{fmt.Sprintf("http://%s:%d", ip, port.Int())}
	if err := config.Load("", cfg); err != nil {
		return nil, err
	}
	var etcdClient *clientv3.Client
	Eventually(func() error {
		var etcdError error
		etcdClient, etcdError = common_etcd.NewClient(cfg)
		return etcdError
	}, "10s", "100ms").Should(Succeed())

	_, err = etcdClient.Put(ctx, "test", "test")
	Expect(err).ToNot(HaveOccurred())
	response, err := etcdClient.Get(ctx, "test")
	Expect(err).ToNot(HaveOccurred())
	Eventually(response.Kvs[0].Value).Should(Equal([]byte("test")))
	defer etcdClient.Close()

	return cfg, nil
}
