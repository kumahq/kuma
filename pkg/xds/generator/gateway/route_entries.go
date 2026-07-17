package gateway

import (
	"strings"

	"github.com/kumahq/kuma/v3/pkg/xds/generator/gateway/route"
)

// HandlePrefixMatches expands prefix path matches into a matching pair of
// prefix and exact matches.
func HandlePrefixMatches(exactEntries, prefixEntries map[string][]route.Entry, entries []route.Entry) []route.Entry {
	// The Kubernetes Ingress and Gateway APIs define prefix matching
	// to match in terms of path components, so we follow suit here.
	// Envoy path prefix matching is byte-wise, so we need to do some
	// transformations. Unless there is already an exact match for the
	// path in question, we expand each prefix path to both a prefix and
	// an exact path, duplicating the route.
	for path, pathEntries := range prefixEntries {
		exactPath := strings.TrimRight(path, "/")

		_, hasExactMatch := exactEntries[exactPath]

		for _, e := range pathEntries {
			var exactPathRewrite *string
			if rw := e.Rewrite; rw != nil && rw.ReplacePrefixMatch != nil {
				rewrite := strings.TrimRight(*rw.ReplacePrefixMatch, "/")
				exactPathRewrite = &rewrite
			}
			var exactPathRedirectPathRewrite *string
			if redir := e.Action.Redirect; redir != nil {
				if rw := redir.PathRewrite; rw != nil && rw.ReplacePrefixMatch != nil {
					rewrite := strings.TrimRight(*rw.ReplacePrefixMatch, "/")
					exactPathRedirectPathRewrite = &rewrite
				}
			}

			// Make sure the prefix has a trailing '/' so that it only matches
			// complete path components.
			e.Match.PrefixPath = exactPath + "/"

			// We need to make sure the prefix replacement
			// _also_ gets a trailing slash
			if exactPathRewrite != nil {
				replace := *exactPathRewrite + "/"
				e.Rewrite.ReplacePrefixMatch = &replace
			}
			if exactPathRedirectPathRewrite != nil {
				replace := *exactPathRedirectPathRewrite + "/"
				e.Action.Redirect.PathRewrite.ReplacePrefixMatch = &replace
			}
			entries = append(entries, e)

			// Duplicate the route to an exact match only if there
			// isn't already an exact match for this path.
			if !hasExactMatch {
				// For the root prefix ("/"), the prefix match already becomes "/",
				// which matches every path including "/", so synthesizing an extra
				// exact "/" match is pure redundancy. Keep emitting it when a prefix
				// rewrite or redirect path-rewrite is present, because there the exact
				// match carries different full-path rewrite semantics for "/".
				if exactPath == "" && exactPathRewrite == nil && exactPathRedirectPathRewrite == nil {
					continue
				}
				exactMatch := e
				exactMatch.Match.PrefixPath = ""
				exactMatch.Match.ExactPath = exactPath
				if exactPath == "" {
					exactMatch.Match.ExactPath = "/"
				}

				// We need to make sure this prefix replacement
				// does _not_ get a trailing slash (unless it is "/")
				if exactPathRewrite != nil {
					path := *exactPathRewrite
					if path == "" {
						path = "/"
					}
					exactMatch.Rewrite = &route.Rewrite{
						ReplaceFullPath: &path,
					}
				}
				if exactPathRedirectPathRewrite != nil {
					path := *exactPathRedirectPathRewrite
					if path == "" {
						path = "/"
					}
					redir := *exactMatch.Action.Redirect
					redir.PathRewrite = &route.Rewrite{
						ReplaceFullPath: &path,
					}
					exactMatch.Action.Redirect = &redir
				}
				exactEntries[exactPath] = append(exactEntries[exactPath], exactMatch)
			}
		}
	}

	for _, pathEntries := range exactEntries {
		entries = append(entries, pathEntries...)
	}

	return entries
}
