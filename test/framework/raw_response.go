package framework

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"sigs.k8s.io/yaml"
)

// ApplyResourceRawResponse sends a YAML resource via HTTP PUT to the kuma-cp
// REST API and returns the raw response formatted as
// "<proto> <status>\n\n<body>\n". The format is stable across runs, suitable
// for golden file comparison.
//
// resourcePath is the plural REST path segment (e.g., "meshtimeouts"). The URL
// is built from the resource's metadata as
// "{api}/meshes/{mesh}/{resourcePath}/{name}".
func ApplyResourceRawResponse(cluster Cluster, resourcePath, yamlBody string) string {
	jsonBody, err := yaml.YAMLToJSON([]byte(yamlBody))
	if err != nil {
		return fmt.Sprintf("yaml-to-json error: %v", err)
	}

	var meta struct {
		Name string `json:"name"`
		Mesh string `json:"mesh"`
	}
	if err := json.Unmarshal(jsonBody, &meta); err != nil {
		return fmt.Sprintf("parse name/mesh error: %v", err)
	}

	url := fmt.Sprintf("%s/meshes/%s/%s/%s",
		cluster.GetKuma().GetAPIServerAddress(),
		meta.Mesh, resourcePath, meta.Name)
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Sprintf("request create error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("request error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("%s %s\n\nbody read error: %v\n", resp.Proto, resp.Status, err)
	}

	return fmt.Sprintf("%s %s\n\n%s\n", resp.Proto, resp.Status, strings.TrimRight(string(body), "\n"))
}
