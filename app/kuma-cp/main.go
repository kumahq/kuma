package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/pkg/profile"

	"github.com/kumahq/kuma/app/kuma-cp/cmd"
)

func main() {
	defer profile.Start(profile.MemProfile).Stop()

	go func() {
		http.ListenAndServe(":8080", nil)
	}()
	cmd.Execute()
}
