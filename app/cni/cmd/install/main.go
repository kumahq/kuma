package main

import (
	"context"

	"github.com/kumahq/kuma/v2/app/cni/pkg/install"
)

func main() {
	install.Run(context.Background())
}
