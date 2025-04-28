package container

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerBuilder struct {
	image      string
	privileged bool
	cmd        []string
	postStart  [][]string
	waitingFor []wait.Strategy
	files      []testcontainers.ContainerFile
}

func NewContainerSetup() *ContainerBuilder {
	return &ContainerBuilder{
		cmd: []string{"sleep", "infinity"},
	}
}

func (b *ContainerBuilder) WithImage(image string) *ContainerBuilder {
	b.image = image
	return b
}

func (b *ContainerBuilder) WithPrivileged(privileged bool) *ContainerBuilder {
	b.privileged = privileged
	return b
}

func (b *ContainerBuilder) WithFiles(filesMap map[string]string) *ContainerBuilder {
	for filePath, content := range filesMap {
		b.files = append(b.files, testcontainers.ContainerFile{
			Reader:            strings.NewReader(content),
			ContainerFilePath: filePath,
			FileMode:          0o644,
		})
	}
	return b
}

func (b *ContainerBuilder) WithKumactlBinary(binary string) *ContainerBuilder {
	b.files = append(b.files, testcontainers.ContainerFile{
		HostFilePath:      binary,
		ContainerFilePath: "/usr/local/bin/kumactl",
		FileMode:          0o700,
	})

	b.waitingFor = append(
		b.waitingFor,
		wait.ForExec([]string{"kumactl", "version"}).
			WithStartupTimeout(10*time.Second).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}).
			WithResponseMatcher(func(body io.Reader) bool {
				data, _ := io.ReadAll(body)
				return bytes.Contains(data, []byte("Client: "))
			}),
	)

	return b
}

func (b *ContainerBuilder) WithCmd(cmd ...string) *ContainerBuilder {
	b.cmd = cmd
	return b
}

func (b *ContainerBuilder) WithPostStart(postStart [][]string) *ContainerBuilder {
	b.postStart = postStart
	return b
}

func (b *ContainerBuilder) WithWaitingForCmd(cmd ...string) *ContainerBuilder {
	b.waitingFor = append(b.waitingFor, wait.ForExec(cmd))
	return b
}

func (b *ContainerBuilder) Start(ctx context.Context) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:      b.image,
				Privileged: b.privileged,
				Files:      b.files,
				Cmd:        b.cmd,
				LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
					{PostStarts: buildContainerHooks(b.postStart)},
				},
				WaitingFor: wait.ForAll(b.waitingFor...),
			},
			Started: true,
		},
	)
}

// buildContainerHook constructs a testcontainers.ContainerHook that executes a
// provided command within a test container.
//
// This function is used to create hooks for running specific commands inside a
// test container. The hook captures the command's exit code and standard
// output. If the command exits with a non-zero status code, an error is
// returned with details including the command, exit code, and standard output.
func buildContainerHook(cmd []string) testcontainers.ContainerHook {
	return func(ctx context.Context, container testcontainers.Container) error {
		status, reader, err := container.Exec(ctx, cmd)
		if err != nil {
			return err
		}

		if status != 0 {
			buf := new(strings.Builder)
			if _, err := io.Copy(buf, reader); err != nil {
				return err
			}

			return errors.Errorf(
				"%s failed (exit code: %d): %s",
				strings.Join(cmd, " "),
				status,
				buf.String(),
			)
		}

		return nil
	}
}

func buildContainerHooks(cmds [][]string) []testcontainers.ContainerHook {
	var hooks []testcontainers.ContainerHook

	for _, cmd := range cmds {
		hooks = append(hooks, buildContainerHook(cmd))
	}

	return hooks
}
