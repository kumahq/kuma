package merge

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

// indicates that all slices that start with this prefix will be appended, not replaced
const appendSlicesPrefix = "Append"

// Confs returns list of confs that may be apply to separate sets of refs.
// In the usual case it has a single element but for MeshHTTPRoute it is keyed
// by hostname.
func Confs(confs []interface{}) ([]interface{}, error) {
	if len(confs) == 0 {
		return nil, nil
	}

	// Sort the confs (potentially) into sets grouped by `mergeValues` fields
	taggedConfsList, err := handleMergeValues(confs)
	if err != nil {
		return nil, err
	}

	var interfaces []interface{}

	// Merge each tagged sets of confs
	for _, taggedConfs := range taggedConfsList {
		confs := taggedConfs.Confs

		result, err := mergeJSONPatches(confs)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't merge JSON patches")
		}

		valueResult := reflect.ValueOf(result)
		// clear appendable slices, so we won't duplicate values of the last conf
		clearAppendSlices(valueResult)
		for i := 0; i < len(confs); i++ {
			appendSlices(valueResult, confs[i])
		}

		if err := handleMergeByKeyFields(valueResult); err != nil {
			return nil, err
		}

		interfaces = append(interfaces, valueResult.Elem().Interface())
	}

	return interfaces, nil
}

// mergeJSONPatches merges a list of confs to a single conf using the algorithm described in https://www.rfc-editor.org/rfc/rfc7396
func mergeJSONPatches(confs []reflect.Value) (interface{}, error) {
	resultBytes := []byte{}
	for i := 0; i < len(confs); i++ {
		conf := confs[i].Interface()
		confBytes, err := json.Marshal(conf)
		if err != nil {
			return nil, err
		}
		if len(resultBytes) == 0 {
			resultBytes = confBytes
			continue
		}
		resultBytes, err = jsonpatch.MergePatch(resultBytes, confBytes)
		if err != nil {
			return nil, err
		}
	}

	confType := confs[0].Type()
	result, err := newConf(confType)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return nil, err
	}

	return result, nil
}

type acc struct {
	// Skip defines whether this item can be ignored because it appears later in
	// the list
	Skip bool
	// The value of the `mergeKey`
	Key interface{}
	// The `Default` value
	Defaults []interface{}
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
	if withSet, ok := valueResult.Interface().(core_model.TransformDefaultAfterMerge); ok {
		withSet.Transform()
	}
	return nil
}

type SetConfField func(reflect.Value, []string)

type GroupedConfs struct {
	Confs []reflect.Value
}

// handleMergeValues takes conf objects and returns
// * a list of the confs keyed by `mergeValues` tagged fields
// Note that if there is no `mergeValues` field, given values are returned in a
// single element list.
// See merge_test.go for an example.
func handleMergeValues(confs []interface{}) ([]GroupedConfs, error) {
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
	for i := 0; i < vals.Len(); i++ {
		value := vals.Index(i)

		mergeKeyValue := value.FieldByName(key.Name).Interface()
		valueDef := []interface{}{value.FieldByName(defaultFieldName).Interface()}

		// We can't have a map keyed by matches so we use a slice and call
		// search through it calling `DeepEqual`. We define the order of matches
		// by where it appears with the most precedence (i.e. the last appearance)
		for accIndex, accRule := range defaultsByKey {
			if accRule.Skip {
				continue
			}
			if !reflect.DeepEqual(accRule.Key, mergeKeyValue) {
				continue
			}
			valueDef = append(accRule.Defaults, valueDef...)
			// Later rules overwrite earlier ones but we also want the order of
			// the later rule to take priority so we skip this in the future
			defaultsByKey[accIndex] = acc{
				Skip: true,
			}
		}
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
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
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

func newConf(t reflect.Type) (interface{}, error) {
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
					s := reflect.AppendSlice(dstField.Elem(), srcField.Elem())
					dstField.Elem().Set(s)
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

func Entries[B common.BaseEntry, T common.Entry[B]](items []T) ([]interface{}, error) {
	var confs []interface{}
	for _, item := range items {
		confs = append(confs, item.GetEntry().GetDefault())
	}
	return Confs(confs)
}
