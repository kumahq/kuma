package lua

import (
	"embed"
	"io/fs"
)

//go:embed service-identity-injection.lua
var ServiceIdentityInjectionFile embed.FS

func ServiceIdentityInjectionScript() (string, error) {
	script, err := fs.ReadFile(ServiceIdentityInjectionFile, "service-identity-injection.lua")
	if err != nil {
		return "", err
	}
	return string(script), nil
}
