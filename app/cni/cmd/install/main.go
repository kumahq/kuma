package main

import (
	"context"

	"github.com/kumahq/kuma/v3/app/cni/pkg/install"
)

func main() {
	install.Run(context.Background())
}
