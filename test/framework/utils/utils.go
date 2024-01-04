package utils

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"text/template"

	"github.com/onsi/gomega"
)

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}

func GetFreePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

func FromTemplate(g gomega.Gomega, tmpl string, data any) string {
	t, err := template.New("tmpl").Parse(tmpl)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	b := bytes.Buffer{}
	g.Expect(t.Execute(&b, data)).To(gomega.Succeed())
	return b.String()
}
