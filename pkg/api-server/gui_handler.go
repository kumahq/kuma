package api_server

import (
	"bytes"
	"html/template"
	"net/http"
	"path"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
)

var disabledPage = `
<!DOCTYPE html><html lang=en>
	<head>
		<style>
			.center {
				display: flex;
				justify-content: center;
				align-items: center;
				height: 200px;
				border: 3px solid green;
			}
		</style>
	</head>
	<body>
		<div class="center"><strong>
		GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP.
		If this isn't a Zone CP the GUI can be enabled by setting the configuration KUMA_API_SERVER_GUI_ENABLED=true.
		</strong></div>
	</body>
</html>
`

type GuiConfig struct {
	ApiUrl      string `json:"apiUrl"`
	BaseGuiPath string `json:"baseGuiPath"`
	Version     string `json:"version"`
}

func NewGuiHandler(guiPath string, enabledGui bool, guiConfig GuiConfig) (http.Handler, error) {
	if enabledGui {
		guiFs := resources.GuiFS()
		tmpl := template.Must(template.ParseFS(guiFs, "index.html"))
		buf := bytes.Buffer{}
		err := tmpl.Execute(&buf, guiConfig)
		if err != nil {
			return nil, err
		}
		childHandler := http.FileServer(http.FS(guiFs))
		return http.StripPrefix(guiPath, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/" && path.Ext(request.URL.Path) == "" {
				request.URL.Path = "/"
			}
			if request.URL.Path == "/" {
				writer.WriteHeader(http.StatusOK)
				writer.Header().Set("Content-Type", "text/html")
				_, _ = writer.Write(buf.Bytes())
				return
			}
			childHandler.ServeHTTP(writer, request)
		})), nil
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte(disabledPage))
		if err != nil {
			log.Error(err, "could not write the response")
		}
	}), nil
}
