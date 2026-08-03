package authn

import (
	"strings"

	"github.com/emicklei/go-restful/v3"
)

const (
	MetadataAuthKey  = "auth"
	MetadataAuthSkip = "skip"
)

// UnauthorizedPathPrefixes is another way to add path prefixes as unauthorized endpoint.
// Prefer MetadataAuthKey, use UnauthorizedPathPrefixes if needed.
var UnauthorizedPathPrefixes = map[string]struct{}{
	"/gui": {},
}

func SkipAuth(request *restful.Request) bool {
	if route := request.SelectedRoute(); route != nil {
		if route.Metadata()[MetadataAuthKey] == MetadataAuthSkip {
			return true
		}
	}
	for prefix := range UnauthorizedPathPrefixes {
		if strings.HasPrefix(request.Request.RequestURI, prefix) {
			return true
		}
	}
	return false
}
