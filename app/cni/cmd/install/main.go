package main

import (
	"context"
	"os"

	"github.com/kumahq/kuma/v3/app/cni/pkg/install"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "ready" {
		install.CheckReady()
		return
	}

	install.Run(context.Background())
}
