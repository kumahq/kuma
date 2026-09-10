package merge

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/pkg/errors"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// indicates that all slices that start with this prefix will be appended, not replaced
const appendSlicesPrefix = "Append"

// Confs returns list of confs that may be apply to separate sets of refs.
// In the usual case it has a single element but for MeshHTTPRoute it is keyed
// by hostname.
func Confs(confs []any) ([]any, error) {
	if len(confs) == 0 {
		return nil, nil
	}

	// Sort the confs (potentially) into sets grouped by `mergeValues` fields
	taggedConfsList, err := handleMergeValues(confs)
	if err != nil {
		return nil, err
	}

	var interfaces []any

	// Merge each tagged sets of confs
	for _, taggedConfs := range taggedConfsList {
		confs := taggedConfs.Confs

		result, err := mergeConfs(confs)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't merge policy configurations")
		}

		valueResult := reflect.ValueOf(result)
		// clear appendable slices, so we won't duplicate values of the last conf
		clearAppendSlices(valueResult)
		for i := range confs {
			appendSlices(valueResult, confs[i])
		}

		if err := handleMergeByKeyFields(valueResult); err != nil {
			return nil, err
		}

		interfaces = append(interfaces, valueResult.Elem().Interface())
	}

	return interfaces, nil
}

// mergeConfs merges a list of confs to a single conf using the algorithm described in https://www.rfc-editor.org/rfc/rfc7396
func mergeConfs(confs []reflect.Value) (any, error) {
	if getConfMeta(derefType(confs[0].Type())).supported {
		return mergeTyped(confs)
	}
	return mergeViaJSON(confs)
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// mergeTyped applies RFC 7396 semantics directly on conf structs, avoiding JSON round-trips entirely. EXC:FILE011:explains-why-not-what
func mergeTyped(confs []reflect.Value) (any, error) {
	confType := confs[0].Type()
	result, err := newConf(confType)
	if err != nil {
		return nil, err
	}

	acc := reflect.ValueOf(result).Elem()
	if acc.Kind() == reflect.Pointer {
		acc.Set(reflect.New(acc.Type().Elem()))
		acc = acc.Elem()
	}
	for i := range confs {
		v := confs[i]
		for v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return nil, errors.New("cannot merge nil conf")
			}
			v = v.Elem()
		}
		mergeInto(acc, v)
	}

	return result, nil
}

func mergeInto(acc, patch reflect.Value) {
	meta := getConfMeta(patch.Type())
	for _, f := range meta.fields {
		pv := patch.Field(f.index)
		if f.omitEmpty && isEmptyValue(pv) {
			continue
		}
		mergeValue(acc.Field(f.index), pv, f)
	}
}

func mergeValue(acc, patch reflect.Value, f confField) {
	if patch.Kind() == reflect.Pointer {
		if patch.IsNil() {
			acc.Set(reflect.Zero(acc.Type()))
			return
		}
		if f.leaf {
			acc.Set(deepCopyValue(patch))
			return
		}
		if acc.Kind() == reflect.Pointer && !acc.IsNil() {
			mergeInto(acc.Elem(), patch.Elem())
			return
		}
		acc.Set(deepCopyValue(patch))
		return
	}
	if f.leaf {
		acc.Set(deepCopyValue(patch))
		return
	}
	mergeInto(acc, patch)
}

func deepCopyValue(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		c := reflect.New(v.Type().Elem())
		c.Elem().Set(deepCopyValue(v.Elem()))
		return c
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		c := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := range v.Len() {
			c.Index(i).Set(deepCopyValue(v.Index(i)))
		}
		return c
	case reflect.Struct:
		c := reflect.New(v.Type()).Elem()
		c.Set(v)
		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				c.Field(i).Set(deepCopyValue(v.Field(i)))
			}
		}
		return c
	default:
		return v
	}
}

type confField struct {
	index     int
	omitEmpty bool
	leaf      bool
}

type confMeta struct {
	fields    []confField
	supported bool
}

var confMetaCache sync.Map

func getConfMeta(t reflect.Type) *confMeta {
	if meta, ok := confMetaCache.Load(t); ok {
		return meta.(*confMeta)
	}
	meta := buildConfMeta(t)
	confMetaCache.Store(t, meta)
	return meta
}

var scalarJSONLeaves = []reflect.Type{
	reflect.TypeFor[k8s.Duration](),
	reflect.TypeFor[intstr.IntOrString](),
}

var marshalerType = reflect.TypeFor[json.Marshaler]()

// buildConfMeta decides whether a conf type can be merged without JSON
// round-trips: only plain structs, pointers, slices and scalar leaves are
// supported. Types with custom JSON marshaling (except known scalar leaves
// like k8s.Duration), maps, interfaces and embedded fields fall back to
// mergeViaJSON to preserve exact semantics. EXC:FILE011:documents-a-non-obvious-invariant
func buildConfMeta(t reflect.Type) *confMeta {
	meta := &confMeta{supported: true}
	for f := range t.Fields() {
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts := parseJSONTag(tag)
		if f.Anonymous && name == "" {
			meta.supported = false
			return meta
		}
		leaf, supported := classifyField(f.Type)
		if !supported {
			meta.supported = false
			return meta
		}
		meta.fields = append(meta.fields, confField{
			index:     f.Index[0],
			omitEmpty: opts.Contains("omitempty"),
			leaf:      leaf,
		})
	}
	return meta
}

func classifyField(t reflect.Type) (bool, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if slices.Contains(scalarJSONLeaves, t) {
		return true, true
	}
	if t.Implements(marshalerType) || reflect.PointerTo(t).Implements(marshalerType) {
		return false, false
	}
	switch t.Kind() {
	case reflect.Struct:
		for f := range t.Fields() {
			if !f.IsExported() {
				continue
			}
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			if f.Anonymous && tag == "" {
				return false, false
			}
			if _, supported := classifyField(f.Type); !supported {
				return false, false
			}
		}
		return false, true
	case reflect.Slice, reflect.Array:
		_, supported := classifyField(t.Elem())
		return true, supported
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true, true
	default:
		return false, false
	}
}

func parseJSONTag(tag string) (string, tagOptions) {
	name, opts, _ := strings.Cut(tag, ",")
	return name, tagOptions(opts)
}

type tagOptions string

func (o tagOptions) Contains(option string) bool {
	for s := range strings.SplitSeq(string(o), ",") {
		if s == option {
			return true
		}
	}
	return false
}

// isEmptyValue mirrors encoding/json's omitempty emptiness check. EXC:FILE011:documents-a-non-obvious-invariant
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	default:
		return false
	}
}

// mergeViaJSON merges a list of confs to a single conf using the algorithm described in https://www.rfc-editor.org/rfc/rfc7396
func mergeViaJSON(confs []reflect.Value) (any, error) {
	confType := confs[0].Type()
	result, err := newConf(confType)
	if err != nil {
		return nil, err
	}

	firstBytes, err := json.Marshal(confs[0].Interface())
	if err != nil {
		return nil, err
	}
	if len(confs) == 1 {
		if err := json.Unmarshal(firstBytes, result); err != nil {
			return nil, err
		}
		return result, nil
	}

	// Decode every conf once and merge on a tree of raw documents: round-tripping the accumulated document through jsonpatch.MergePatch on every step is quadratic in its size. EXC:FILE011:explains-why-not-what
	acc, err := decodeDoc(firstBytes)
	if err != nil {
		return nil, err
	}
	for i := 1; i < len(confs); i++ {
		confBytes, err := json.Marshal(confs[i].Interface())
		if err != nil {
			return nil, err
		}
		patch, err := decodeDoc(confBytes)
		if err != nil {
			return nil, err
		}
		if patch == nil {
			acc = nil
			continue
		}
		if acc == nil {
			return nil, errors.New("invalid JSON document")
		}
		mergePatchDocs(acc, patch)
	}

	resultBytes, err := json.Marshal(acc)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resultBytes, result); err != nil {
		return nil, err
	}

	return result, nil
}

// rawDoc is a decoded JSON object. Values are nil (JSON null), json.RawMessage
// (verbatim leaf, including unmerged objects) or rawDoc (an object that took
// part in a merge). Keeping untouched subtrees as RawMessage avoids
// decode/encode round-trips for them. EXC:FILE011:documents-a-non-obvious-invariant
type rawDoc map[string]any

func decodeDoc(data []byte) (rawDoc, error) {
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	res := make(rawDoc, len(doc))
	for k, v := range doc {
		res[k] = v
	}
	return res, nil
}

func isNullNode(v any) bool {
	if v == nil {
		return true
	}
	raw, ok := v.(json.RawMessage)
	return ok && isNullJSON(raw)
}

func isNullJSON(raw json.RawMessage) bool {
	return string(raw) == "null"
}

// asDoc decodes v into a rawDoc if v is a JSON object. EXC:FILE011:documents-a-non-obvious-invariant
func asDoc(v any) (rawDoc, bool) {
	switch t := v.(type) {
	case rawDoc:
		return t, true
	case json.RawMessage:
		doc, err := decodeDoc(t)
		if err != nil {
			return nil, false
		}
		return doc, doc != nil
	default:
		return nil, false
	}
}

// mergePatchDocs applies RFC 7396 semantics of jsonpatch.MergePatch: null patch values delete the key, object values present on both sides merge recursively, anything else replaces. EXC:FILE011:documents-a-non-obvious-invariant
func mergePatchDocs(doc, patch rawDoc) {
	for k, v := range patch {
		if isNullNode(v) {
			delete(doc, k)
			continue
		}
		cur, ok := doc[k]
		if !ok || isNullNode(cur) {
			doc[k] = pruneNulls(v)
			continue
		}
		curDoc, curIsDoc := asDoc(cur)
		patchDoc, patchIsDoc := asDoc(v)
		switch {
		case curIsDoc && patchIsDoc:
			mergePatchDocs(curDoc, patchDoc)
			doc[k] = curDoc
		case !curIsDoc:
			doc[k] = pruneNulls(v)
		default:
			doc[k] = v
		}
	}
}

// pruneNulls mirrors json-patch: nulls are pruned recursively, including inside arrays — stricter than RFC 7396 (arrays atomic), but byte-compatible with the previous jsonpatch.MergePatch chain, which is the compatibility bar. EXC:FILE011:documents-a-non-obvious-invariant
func pruneNulls(v any) any {
	switch t := v.(type) {
	case rawDoc:
		for k, val := range t {
			if isNullNode(val) {
				delete(t, k)
			} else {
				t[k] = pruneNulls(val)
			}
		}
		return t
	case json.RawMessage:
		if doc, ok := asDoc(t); ok {
			return pruneNulls(doc)
		}
		var ary []json.RawMessage
		if err := json.Unmarshal(t, &ary); err == nil {
			for i, elem := range ary {
				if elem == nil || isNullJSON(elem) {
					continue
				}
				switch pruned := pruneNulls(elem).(type) {
				case json.RawMessage:
					ary[i] = pruned
				default:
					out, err := json.Marshal(pruned)
					if err != nil {
						continue
					}
					ary[i] = out
				}
			}
			out, err := json.Marshal(ary)
			if err != nil {
				return t
			}
			return json.RawMessage(out)
		}
		return t
	default:
		return v
	}
}

type acc struct {
	// Skip defines whether this item can be ignored because it appears later in
	// the list
	Skip bool
	// The value of the `mergeKey`
	Key any
	// The `Default` value
	Defaults []any
}

const (
	defaultFieldName = "Default"
	policyMergeTag   = "policyMerge"
	mergeValuesByKey = "mergeValuesByKey"
	mergeValues      = "mergeValues"
	mergeKey         = "mergeKey"
)

func handleMergeByKeyFields(valueResult reflect.Value) error {
	confType := valueResult.Elem().Type()
	for fieldIndex := 0; fieldIndex < confType.NumField(); fieldIndex++ {
		field := confType.Field(fieldIndex)
		if !strings.Contains(field.Tag.Get(policyMergeTag), mergeValuesByKey) {
			continue
		}
		if field.Type.Kind() != reflect.Slice && field.Type.Elem().Kind() != reflect.Struct {
			return errors.New("a merge by key field must be a slice of structs")
		}

		entriesValue := valueResult.Elem().Field(fieldIndex)

		merged, err := mergeByKey(entriesValue)
		if err != nil {
			return err
		}
		entriesValue.Set(merged)
	}
	// Some policies transform their values when `GetDefault` is called. See
	// `MeshHTTPRoute`. Basically the order when merging is increasing
	// precedence but with MeshHTTPRoute/Gateway API we have "the first rule"
	// wins as a fallback ordering.
	// So we need to basically unreverse the transformed rules.
	if withSet, ok := reflect.TypeAssert[core_model.TransformDefaultAfterMerge](valueResult); ok {
		withSet.Transform()
	}
	return nil
}

type GroupedConfs struct {
	Confs []reflect.Value
}

// handleMergeValues takes conf objects and returns
// * a list of the confs keyed by `mergeValues` tagged fields
// Note that if there is no `mergeValues` field, given values are returned in a
// single element list.
// See merge_test.go for an example.
func handleMergeValues(confs []any) ([]GroupedConfs, error) {
	confType := reflect.TypeOf(confs[0])

	// We construct a map of strings to confs
	var keyFieldIndex *int
	// Find a field tagged with `mergeValues`
	for fieldIndex := 0; fieldIndex < confType.NumField(); fieldIndex++ {
		field := confType.Field(fieldIndex)
		if field.Tag.Get(policyMergeTag) != mergeValues {
			continue
		}
		if field.Type.Kind() != reflect.Slice || field.Type.Elem().Kind() != reflect.String {
			return nil, fmt.Errorf("a mergeValues field must be a slice of strings")
		}

		keyFieldIndex = pointer.To(fieldIndex)
	}

	intermediateMergeMap := map[string][]reflect.Value{}

	// Track ordered mergeValues field values
	var orderedIntermediateKeys []string

	// Put every conf into the map for every `mergeValues` field value it has
	for _, conf := range confs {
		confVal := reflect.ValueOf(conf)
		// If there is no such `mergeValues` field, put everything under the
		// empty string
		keys := []string{""}
		if keyFieldIndex != nil {
			confKeys := confVal.Field(*keyFieldIndex)
			// Treat the empty list of values like "" in our map
			if confKeys.Len() > 0 {
				keys = confKeys.Interface().([]string)
			}
		}

		for _, key := range keys {
			keyedConf := reflect.New(confVal.Type())
			keyedConf.Elem().Set(confVal)
			values, ok := intermediateMergeMap[key]
			if !ok {
				orderedIntermediateKeys = append(orderedIntermediateKeys, key)
			}
			// Set the singular key on copies of these confs and add them to the
			// accumulator
			if key != "" {
				singleKeys := reflect.ValueOf([]string{key})
				keyedConf.Elem().Field(*keyFieldIndex).Set(singleKeys)
			}
			intermediateMergeMap[key] = append(values, keyedConf.Elem())
		}
	}

	var taggedValues []GroupedConfs
	for _, key := range orderedIntermediateKeys {
		confs, ok := intermediateMergeMap[key]
		if !ok {
			panic("internal merge error")
		}

		taggedValues = append(taggedValues, GroupedConfs{
			Confs: confs,
		})
	}

	return taggedValues, nil
}

func mergeByKey(vals reflect.Value) (reflect.Value, error) {
	if vals.Len() == 0 {
		return reflect.Zero(vals.Type()), nil
	}
	valType := vals.Index(0).Type()
	key, ok := findMergeKeyField(valType)
	if !ok {
		return reflect.Value{}, fmt.Errorf("a merge by key field must have a field tagged as %s and a Default field", mergeKey)
	}
	var defaultsByKey []acc
	// EXC:FILE011:explains-why-not-what — candidates for DeepEqual are bucketed by JSON encoding because matches can't be map keys: DeepEqual values always marshal identically, so only same-bucket entries can match. This avoids scanning every accumulated key with DeepEqual per value, which is quadratic.
	candidatesByJSON := map[string][]int{}
	for i := 0; i < vals.Len(); i++ {
		value := vals.Index(i)

		mergeKeyValue := value.FieldByName(key.Name).Interface()
		valueDef := []any{value.FieldByName(defaultFieldName).Interface()}

		keyJSON, err := json.Marshal(mergeKeyValue)
		if err != nil {
			return reflect.Value{}, err
		}

		// EXC:FILE011:pre-existing-comment — we define the order of matches by where it appears with the most precedence (i.e. the last appearance)
		bucketKey := string(keyJSON)
		bucket := candidatesByJSON[bucketKey]
		active := bucket[:0]
		for _, accIndex := range bucket {
			accRule := defaultsByKey[accIndex]
			if !reflect.DeepEqual(accRule.Key, mergeKeyValue) {
				active = append(active, accIndex)
				continue
			}
			valueDef = append(accRule.Defaults, valueDef...)
			// EXC:FILE011:pre-existing-comment — later rules overwrite earlier ones but we also want the order of the later rule to take priority so we skip this in the future
			defaultsByKey[accIndex] = acc{
				Skip: true,
			}
		}
		candidatesByJSON[bucketKey] = append(active, len(defaultsByKey))
		defaultsByKey = append(defaultsByKey, acc{
			Key:      mergeKeyValue,
			Defaults: valueDef,
		})
	}
	keyValues := reflect.Zero(vals.Type())
	for _, confs := range defaultsByKey {
		if confs.Skip {
			continue
		}
		merged, err := Confs(confs.Defaults)
		if err != nil {
			return reflect.Value{}, err
		}

		for _, mergedConf := range merged {
			keyValueP := reflect.New(valType)
			keyValue := keyValueP.Elem()

			keyValue.FieldByName(key.Name).Set(reflect.ValueOf(confs.Key))
			keyValue.FieldByName(defaultFieldName).Set(reflect.ValueOf(mergedConf))

			keyValues = reflect.Append(keyValues, keyValue)
		}
	}
	return keyValues, nil
}

func findMergeKeyField(typ reflect.Type) (reflect.StructField, bool) {
	var key *reflect.StructField
	for field := range typ.Fields() {
		if strings.Contains(field.Tag.Get(policyMergeTag), mergeKey) {
			key = &field
			break
		}
	}
	if key == nil {
		return reflect.StructField{}, false
	}
	if _, ok := typ.FieldByName(defaultFieldName); !ok {
		return reflect.StructField{}, false
	}
	return *key, true
}

func newConf(t reflect.Type) (any, error) {
	if t.Kind() == reflect.Pointer {
		return nil, errors.New("conf is expected to have a non-pointer type")
	}
	return reflect.New(t).Interface(), nil
}

func clearAppendSlices(val reflect.Value) {
	strVal := mustUnwrapStruct(val)
	if strVal == reflect.ValueOf(nil) {
		return
	}
	for i := 0; i < strVal.NumField(); i++ {
		valField := strVal.Field(i)
		field := strVal.Type().Field(i)
		switch valField.Kind() {
		case reflect.Slice:
			mergeByKey := strings.Contains(field.Tag.Get(policyMergeTag), mergeValuesByKey)
			if strings.HasPrefix(field.Name, appendSlicesPrefix) || mergeByKey {
				valField.Set(reflect.Zero(valField.Type()))
			}
		case reflect.Struct:
			clearAppendSlices(valField)
		case reflect.Pointer:
			if valField.Elem().Kind() == reflect.Struct {
				clearAppendSlices(valField)
			}
			if valField.Elem().Kind() == reflect.Slice {
				mergeByKey := strings.Contains(field.Tag.Get(policyMergeTag), mergeValuesByKey)
				if strings.HasPrefix(field.Name, appendSlicesPrefix) || mergeByKey {
					valField.Elem().Set(reflect.Zero(valField.Elem().Type()))
				}
			}
		}
	}
}

// dst and src has to be of a same type
func appendSlices(dst reflect.Value, src reflect.Value) {
	strDst := mustUnwrapStruct(dst)
	strSrc := mustUnwrapStruct(src)
	if strSrc == reflect.ValueOf(nil) || strDst == reflect.ValueOf(nil) {
		return
	}
	for i := 0; i < strDst.NumField(); i++ {
		dstField := strDst.Field(i)
		srcField := strSrc.Field(i)

		field := strDst.Type().Field(i)
		switch dstField.Kind() {
		case reflect.Slice:
			mergeByKey := strings.Contains(field.Tag.Get(policyMergeTag), mergeValuesByKey)
			if strings.HasPrefix(field.Name, appendSlicesPrefix) || mergeByKey {
				s := reflect.AppendSlice(dstField, srcField)
				dstField.Set(s)
			}
		case reflect.Struct:
			appendSlices(dstField, srcField)
		case reflect.Pointer:
			if dstField.Elem().Kind() == reflect.Struct {
				appendSlices(dstField, srcField)
			}
			if dstField.Elem().Kind() == reflect.Slice {
				mergeByKey := strings.Contains(field.Tag.Get(policyMergeTag), mergeValuesByKey)
				if strings.HasPrefix(field.Name, appendSlicesPrefix) || mergeByKey {
					if !srcField.IsNil() {
						s := reflect.AppendSlice(dstField.Elem(), srcField.Elem())
						dstField.Elem().Set(s)
					}
				}
			}
		}
	}
}

func mustUnwrapStruct(val reflect.Value) reflect.Value {
	resVal := val
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return reflect.ValueOf(nil)
		}
		resVal = val.Elem()
	}
	if resVal.Kind() != reflect.Struct {
		panic("expected struct or pointer to a struct got " + val.Kind().String())
	}
	return resVal
}

func Entries[B common.BaseEntry, T common.Entry[B]](items []T) ([]any, error) {
	var confs []any
	for _, item := range items {
		confs = append(confs, item.GetEntry().GetDefault())
	}
	return Confs(confs)
}
