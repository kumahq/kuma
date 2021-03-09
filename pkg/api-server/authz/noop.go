package authz

import "github.com/emicklei/go-restful"

func NoAuth(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	chain.ProcessFilter(request, response)
}
