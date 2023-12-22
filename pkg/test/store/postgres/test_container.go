package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kumahq/kuma/pkg/config"
	pg_config "github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/plugins/common/postgres"
	util_files "github.com/kumahq/kuma/pkg/util/files"
)

type PostgresContainer struct {
	container testcontainers.Container
	WithTLS   bool
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

func resourceDir() string {
	cwd, _ := os.Getwd()
	_, callerFile, _, _ := runtime.Caller(0)

	dir := findProjectRoot(cwd, callerFile)
	rDir := path.Join(dir, "tools/postgres")
	err := os.Chmod(path.Join(rDir, "certs/postgres.client.key"), 0o600)
	if err != nil {
		panic(err)
	}
	return rDir
}

func (v *PostgresContainer) Start() error {
	ctx := context.Background()

	buildArgs := map[string]*string{}

	mode := "standard"
	if v.WithTLS {
		mode = "tls"
	}
	buildArgs["MODE"] = &mode

	// Add a uniqueId to make each image different, this was causing flakiness as test-container
	// deletes the image that it builds unconditionally during teardown and fails if the image doesn't exist.
	// In the case of parallel tests it's possible that the same image was used in multiple tests and the second teardown would fail.
	uniqueId := strconv.Itoa(GinkgoParallelProcess())
	buildArgs["UNIQUEID"] = &uniqueId

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:   resourceDir(),
			BuildArgs: buildArgs,
		},
		Env: map[string]string{
			"POSTGRES_USER":     "kuma",
			"POSTGRES_PASSWORD": "kuma",
			"POSTGRES_DB":       "kuma",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432"),
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

func (v *PostgresContainer) Stop() error {
	if v.container != nil {
		return v.container.Terminate(context.Background())
	}
	return nil
}

type DbOption string

func (v *PostgresContainer) Config() (pg_config.PostgresStoreConfig, error) {
	cfg := pg_config.DefaultPostgresStoreConfig()
	ip, err := v.container.Host(context.Background())
	if err != nil {
		return pg_config.PostgresStoreConfig{}, err
	}
	port, err := v.container.MappedPort(context.Background(), "5432")
	if err != nil {
		return pg_config.PostgresStoreConfig{}, err
	}
	cfg.Host = ip
	cfg.Port = port.Int()
	rDir := resourceDir()
	if v.WithTLS {
		cfg.TLS.Mode = pg_config.VerifyCa
		cfg.TLS.KeyPath = path.Join(rDir, "certs/postgres.client.key")
		cfg.TLS.CertPath = path.Join(rDir, "certs/postgres.client.crt")
		cfg.TLS.CAPath = path.Join(rDir, "certs/rootCA.crt")
	}
	if err := config.Load("", cfg); err != nil {
		return pg_config.PostgresStoreConfig{}, err
	}
	// make sure the server is reachable
	Eventually(func() error {
		db, err := postgres.ConnectToDb(*cfg)
		if err != nil {
			return err
		}
		defer db.Close()
		cfg.DbName = fmt.Sprintf("kuma_%s", strings.ReplaceAll(core.NewUUID(), "-", ""))
		GinkgoLogr.Info(fmt.Sprintf("Connecting and creating database to container id: %s, "+
			"port 5432 mapped to host port: %d; db name %s", v.container.GetContainerID(), cfg.Port, cfg.DbName))
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DbName))
		return err
	}, "10s", "100ms").Should(Succeed())

	// make sure the database instance is reachable
	Eventually(func() error {
		db, err := postgres.ConnectToDb(*cfg)
		if err != nil {
			return err
		}
		defer db.Close()
		return nil
	}, "10s", "100ms").Should(Succeed())
	GinkgoLogr.Info("Database successfully created and started", "name", cfg.DbName)
	return *cfg, nil
}

func (v *PostgresContainer) PrintDebugInfo(expectedDbName string, expectedDbPort int) {
	container := v.container
	logEntries := []string{
		"db container details:",
		fmt.Sprintf("trying to connect db %s on port %d, latest container: %s",
			expectedDbName, expectedDbPort, container.GetContainerID()),
	}

	if state, _ := container.State(context.TODO()); state != nil {
		logEntries = append(logEntries, fmt.Sprintf("container state: %s", state.Status))
	}

	if logReader, _ := container.Logs(context.Background()); logReader != nil {
		if logs, _ := io.ReadAll(logReader); len(logs) > 0 {
			logEntries = append(logEntries, fmt.Sprintf("container logs: %s", logs))
		}
	}

	rDir := resourceDir()
	clientKey, _ := os.ReadFile(path.Join(rDir, "certs/postgres.client.key"))
	clientCert, _ := os.ReadFile(path.Join(rDir, "certs/postgres.client.crt"))
	_ = container.CopyToContainer(context.Background(), clientKey, "/tmp/postgres.client.key", 0o600)
	_ = container.CopyToContainer(context.Background(), clientCert, "/tmp/postgres.client.crt", 0o600)
	listCmd := []string{"sh", "-c", "export PGSSLMODE=verify-full; export PGSSLROOTCERT=/var/lib/postgresql/rootCA.crt; " +
		"export PGSSLCERT=/tmp/postgres.client.crt; export PGSSLKEY=/tmp/postgres.client.key; psql -h localhost --no-password -l -U kuma -P pager=off"}
	if exitCode, resultReader, err := container.Exec(context.Background(), listCmd); err == nil {
		output, _ := io.ReadAll(resultReader)
		logEntries = append(logEntries, fmt.Sprintf("psql list database(exitCode: %d): %s", exitCode, output))
		if strings.Contains(string(output), expectedDbName) {
			logEntries = append(logEntries, fmt.Sprintf("database %s found in this container %s", expectedDbName, container.GetContainerID()))
		} else {
			logEntries = append(logEntries, fmt.Sprintf("database %s does not exist in this container %s", expectedDbName, container.GetContainerID()))
		}
	} else {
		logEntries = append(logEntries, fmt.Sprintf("error executing psql list database: %v", err))
	}

	GinkgoLogr.Info(strings.Join(logEntries, "\n"))
}
