package authn

import (
	"github.com/emicklei/go-restful/v3"
)

type Authenticator = restful.FilterFunction

func NoopAuthenticator(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	chain.ProcessFilter(request, response)
}
