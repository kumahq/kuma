package tags

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
)

type Tags map[string]string

func (t Tags) WithoutTags(tags ...string) Tags {
	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[t] = true
	}
	result := Tags{}
	for tagName, tagValue := range t {
		if !tagSet[tagName] {
			result[tagName] = tagValue
		}
	}
	return result
}

func (t Tags) WithTags(keysAndValues ...string) Tags {
	result := Tags{}
	for tagName, tagValue := range t {
		result[tagName] = tagValue
	}
	for i := 0; i < len(keysAndValues); {
		key, value := keysAndValues[i], keysAndValues[i+1]
		result[key] = value
		i += 2
	}
	return result
}

func (t Tags) Keys() TagKeys {
	var keys []string
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (t Tags) String() string {
	var pairs []string
	for _, key := range t.Keys() {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, t[key]))
	}
	return strings.Join(pairs, ",")
}

type TagsSlice []Tags
type TagKeys []string
type TagKeysSlice []TagKeys

func (t TagsSlice) ToTagKeysSlice() TagKeysSlice {
	out := []TagKeys{}
	for _, v := range t {
		out = append(out, v.Keys())
	}
	return out
}

// Transform applies each transformer to each TagKeys and returns a sorted unique TagKeysSlice.
func (t TagKeysSlice) Transform(transformers ...TagKeyTransformer) TagKeysSlice {
	allSlices := map[string]TagKeys{}
	for _, tagKeys := range t {
		res := tagKeys.Transform(transformers...)
		if len(res) > 0 {
			h := strings.Join(res, ", ")
			allSlices[h] = res
		}
	}
	out := TagKeysSlice{}
	for _, n := range allSlices {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool {
		for k := 0; k < len(out[i]) && k < len(out[j]); k++ {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return len(out[i]) < len(out[j])
	})
	return out
}

type TagKeyTransformer interface {
	Apply(slice TagKeys) TagKeys
}
type TagKeyTransformerFunc func(slice TagKeys) TagKeys

func (f TagKeyTransformerFunc) Apply(slice TagKeys) TagKeys {
	return f(slice)
}

// Transform applies a list of transformers on the tag keys and return a new set of keys (always return sorted, unique sets).
func (t TagKeys) Transform(transformers ...TagKeyTransformer) TagKeys {
	tmp := t
	for _, tr := range transformers {
		tmp = tr.Apply(tmp)
	}
	// Make tags unique and sorted
	tagSet := map[string]bool{}
	out := TagKeys{}
	for _, n := range tmp {
		if !tagSet[n] {
			tagSet[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

func Without(tags ...string) TagKeyTransformer {
	tagSet := map[string]bool{}
	for _, t := range tags {
		tagSet[t] = true
	}
	return TagKeyTransformerFunc(func(slice TagKeys) TagKeys {
		out := []string{}
		for _, t := range slice {
			if !tagSet[t] {
				out = append(out, t)
			}
		}
		return out
	})
}

func With(tags ...string) TagKeyTransformer {
	return TagKeyTransformerFunc(func(slice TagKeys) TagKeys {
		res := make([]string, len(tags)+len(slice))
		copy(res, slice)
		copy(res[len(slice):], tags)
		return res
	})
}

func TagsFromString(tagsString string) (Tags, error) {
	result := Tags{}
	tagPairs := strings.Split(tagsString, ",")
	for _, pair := range tagPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return nil, errors.New("invalid format of tags, pairs should be separated by , and key should be separated from value by =")
		}
		result[split[0]] = split[1]
	}
	return result, nil
}

func DistinctTags(tags []Tags) []Tags {
	used := map[string]bool{}
	var result []Tags
	for _, tag := range tags {
		str := tag.String()
		if !used[str] {
			result = append(result, tag)
			used[str] = true
		}
	}
	return result
}

func TagKeySlice(tags []Tags) TagKeysSlice {
	r := make([]TagKeys, len(tags))
	for i := range tags {
		r[i] = tags[i].Keys()
	}
	return r
}

func MatchingRegex(tags mesh_proto.SingleValueTagSet) (re string) {
	for _, key := range tags.Keys() {
		keyIsEqual := fmt.Sprintf(`&%s=`, key)
		var value string
		switch tags[key] {
		case "*":
			value = ``
		default:
			value = fmt.Sprintf(`[^&]*%s[,&]`, tags[key])
		}
		value = strings.ReplaceAll(value, ".", `\.`)
		expr := keyIsEqual + value + `.*`
		re += expr
	}
	re = `.*` + re
	return
}

func RegexOR(r ...string) string {
	if len(r) == 0 {
		return ""
	}
	if len(r) == 1 {
		return r[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(r, "|"))
}

func MatchSourceRegex(policy core_policy.ConnectionPolicy) string {
	var selectorRegexs []string
	for _, selector := range policy.Sources() {
		selectorRegexs = append(selectorRegexs, MatchingRegex(selector.Match))
	}
	return RegexOR(selectorRegexs...)
}
