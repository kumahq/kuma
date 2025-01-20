package main

import (
	"context"

	"github.com/kumahq/kuma/app/cni/pkg/install"
)

func main() {
	install.Run(context.Background())
}
