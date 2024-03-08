package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kumahq/kuma/tools/ci/update-vulnerable-dependencies/leastcommonversion"
)

func main() {
	inBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	in := &leastcommonversion.Input{}
	if err := json.Unmarshal(inBytes, &in); err != nil {
		panic(err)
	}

	in.Current = strings.TrimSuffix(in.Current, "\n")
	version, err := leastcommonversion.Deduct(in)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, "null")
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
	_, _ = fmt.Fprintln(os.Stdout, version)
}
