package api_server

import (
	"net/url"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/api-server/types"

	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/core/resources/model"
)

const maxPageSize = 1000
const defaultPageSize = 100

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

func nextLink(request *restful.Request, publicURL string, list model.ResourceList) (*string, error) {
	if list.GetPagination().NextOffset == "" {
		return nil, nil
	}
	query := request.Request.URL.Query()
	query.Set("offset", list.GetPagination().NextOffset)
	nextURL, err := url.Parse(publicURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse public url")
	}
	nextURL.Path = request.Request.URL.Path
	nextURL.RawQuery = query.Encode()
	urlString := nextURL.String()
	return &urlString, nil
}
