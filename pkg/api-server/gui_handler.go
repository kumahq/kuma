package api_server

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig/v3"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
)

type GuiConfig struct {
	ApiUrl      string `json:"apiUrl"`
	BaseGuiPath string `json:"baseGuiPath"`
	Version     string `json:"version"`
	BasedOnKuma string `json:"basedOnKuma,omitempty"`
	Product     string `json:"product"`
	Mode        string `json:"mode"`
	Environment string `json:"environment"`
	StoreType   string `json:"storeType"`
	ReadOnly    bool   `json:"apiReadOnly"`
}

func NewGuiHandler(guiPath string, disableGui bool, guiConfig GuiConfig) (http.Handler, error) {
	guiFs := resources.GuiFS()
	f := "index.html"
	if disableGui {
		f = "no-gui.html"
	}
	tmpl := template.Must(template.New(f).Funcs(sprig.HtmlFuncMap()).ParseFS(guiFs, f))
	buf := bytes.Buffer{}
	err := tmpl.Execute(&buf, guiConfig)
	if err != nil {
		return nil, err
	}
	return http.StripPrefix(guiPath, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := guiFs.Open(request.URL.Path)
		if err == nil {
			http.FileServer(http.FS(guiFs)).ServeHTTP(writer, request)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write(buf.Bytes())
	})), nil
}
