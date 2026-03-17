package template

import (
	"maps"
	"strings"

	"github.com/hoisie/mustache"
)

type contextMap map[string]any

func (cm contextMap) merge(other contextMap) {
	maps.Copy(cm, other)
}

func newContextMap(key, value string) contextMap {
	if !strings.Contains(key, ".") {
		return map[string]any{
			key: value,
		}
	}

	parts := strings.SplitAfterN(key, ".", 2)
	return map[string]any{
		parts[0][:len(parts[0])-1]: newContextMap(parts[1], value),
	}
}

func Render(template string, values map[string]string) []byte {
	ctx := contextMap{}
	for k, v := range values {
		ctx.merge(newContextMap(k, v))
	}
	data := mustache.Render(template, ctx)
	return []byte(data)
}
