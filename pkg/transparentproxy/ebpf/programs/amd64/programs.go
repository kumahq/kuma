//go:build linux && amd64

package amd64

import "embed"

//go:embed *
var Programs embed.FS
