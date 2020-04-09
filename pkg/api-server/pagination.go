package api_server

import (
	"github.com/Kong/kuma/pkg/api-server/types"
	"github.com/pkg/errors"
	"net/url"
	"strconv"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/emicklei/go-restful"
)

const maxPageSize = 1000
const defaultPageSize = 100

func pagination(request *restful.Request) (int, string, error) {
	pageSize := defaultPageSize
	if request.QueryParameter("size") != "" {
		p, err := strconv.Atoi(request.QueryParameter("size"))
		if err != nil {
			return 0, "", types.InvalidPageSize
		}
		pageSize = p
		if pageSize > maxPageSize {
			return 0, "", types.NewMaxPageSizeExceeded(pageSize, maxPageSize)
		}
	}
	offset := request.QueryParameter("offset")
	return pageSize, offset, nil
}

func nextLink(request *restful.Request, publicURL string, list model.ResourceList) (*string, error) {
	if list.GetPagination() == nil || list.GetPagination().NextOffset == "" {
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
