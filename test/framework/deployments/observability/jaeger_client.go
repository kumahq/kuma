package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type jaegerServicesOutput struct {
	Data []string `json:"data"`
}

func tracedServices(url string) ([]string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/api/services", url), http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	output := &jaegerServicesOutput{}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return nil, err
	}
	sort.Strings(output.Data)
	return output.Data, nil
}
