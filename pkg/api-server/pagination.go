package api_server

import (
	"net/url"
	"strconv"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/types"
)

const (
	maxPageSize     = 1000
	defaultPageSize = 100
)

type page struct {
	size   int
	offset string
}

func pagination(request *restful.Request) (page, error) {
	pageSize := defaultPageSize
	if request.QueryParameter("size") != "" {
		p, err := strconv.Atoi(request.QueryParameter("size"))
		if err != nil {
			return page{}, types.InvalidPageSize
		}
		pageSize = p
		if pageSize > maxPageSize {
			return page{}, types.NewMaxPageSizeExceeded(pageSize, maxPageSize)
		}
	}
	offset := request.QueryParameter("offset")
	return page{
		size:   pageSize,
		offset: offset,
	}, nil
}

func nextLink(request *restful.Request, nextOffset string) *string {
	if nextOffset == "" {
		return nil
	}

	query := request.Request.URL.Query()
	query.Set("offset", nextOffset)

	nextURL := &url.URL{}
	if request.Request.TLS == nil {
		nextURL.Scheme = "http"
	} else {
		nextURL.Scheme = "https"
	}
	nextURL.Host = request.Request.Host
	nextURL.Path = request.Request.URL.Path
	nextURL.RawQuery = query.Encode()
	urlString := nextURL.String()
	return &urlString
}
