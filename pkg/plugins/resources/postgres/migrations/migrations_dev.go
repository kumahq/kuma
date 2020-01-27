// +build dev

package migrations

import (
	"net/http"
)

var Migrations http.FileSystem = http.Dir(MigrationsDir())
