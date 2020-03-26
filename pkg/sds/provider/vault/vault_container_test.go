// +build integration

package vault_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type VaultContainer struct {
	Token     string
	container testcontainers.Container
}

func (v *VaultContainer) Start() error {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "vault",
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  v.Token,
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
		},

		ExposedPorts: []string{"8200/tcp"},
		WaitingFor:   wait.ForListeningPort("8200"),
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

func (v *VaultContainer) Stop() error {
	if v.container != nil {
		return v.container.Terminate(context.Background())
	}
	return nil
}

func (v *VaultContainer) Address() (string, error) {
	ip, err := v.container.Host(context.Background())
	if err != nil {
		return "", err
	}
	port, err := v.container.MappedPort(context.Background(), "8200")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%s", ip, port.Port()), nil
}
