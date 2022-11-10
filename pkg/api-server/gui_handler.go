package api_server

import (
	"encoding/json"
	"net/http"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
)

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
		_, err := writer.Write([]byte("" +
			"<!DOCTYPE html><html lang=en>" +
			"<head>\n<style>\n.center {\n  display: flex;\n  justify-content: center;\n  align-items: center;\n  height: 200px;\n  border: 3px solid green; \n}\n</style>\n</head>" +
			"<body><div class=\"center\"><strong>" +
			"GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP." +
			"</strong></div></body>" +
			"</html>"))
		if err != nil {
			log.Error(err, "could not write the response")
		}
	})
}
