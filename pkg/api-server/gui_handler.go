package api_server

import (
	"encoding/json"
	"net/http"

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

func guiHandler(path string, enabledGui bool, apiUrl string, basePath string) http.Handler {
	if enabledGui {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == path+"config.json" {
				writer.WriteHeader(http.StatusOK)
				writer.Header().Add("Content-Type", "application/json")
				_ = json.NewEncoder(writer).Encode(struct {
					ApiUrl      string `json:"apiUrl"`
					BaseGuiPath string `json:"baseGuiPath"`
				}{
					ApiUrl:      apiUrl,
					BaseGuiPath: basePath,
				})
				return
			}
			http.StripPrefix(path, http.FileServer(http.FS(resources.GuiFS()))).ServeHTTP(writer, request)
		})
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte(disabledPage))
		if err != nil {
			log.Error(err, "could not write the response")
		}
	})
}
