package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// indicates that all slices that start with this prefix will be appended, not replaced
const appendSlicesPrefix = "Append"

func MergeConfs(confs []interface{}) (interface{}, error) {
	if len(confs) == 0 {
		return nil, nil
	}

	resultBytes := []byte{}
	for _, conf := range confs {
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

	confType := reflect.TypeOf(confs[0])
	result, err := newConf(confType)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return nil, err
	}

	valueResult := reflect.ValueOf(result)
	// clear appendable slices, so we won't duplicate values of the last conf
	clearAppendSlices(valueResult)
	for i := range confs {
		// call .Elem() to unwrap interface{}
		appendSlices(valueResult, reflect.ValueOf(&confs[i]).Elem())
	}

	if err := handleMergeByKeyFields(valueResult); err != nil {
		return nil, err
	}

	v := valueResult.Elem().Interface()

	return v, nil
}

type acc struct {
	Key      interface{}
	Defaults []interface{}
}

const (
	defaultFieldName = "Default"
	policyMergeTag   = "policyMerge"
	mergeValuesByKey = "mergeValuesByKey"
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
	return nil
}

func mergeByKey(vals reflect.Value) (reflect.Value, error) {
	if vals.Len() == 0 {
		return reflect.Zero(vals.Type()), nil
	}
	valType := vals.Index(0).Type()
	key, ok := findKeyAndSpec(valType)
	if !ok {
		return reflect.Value{}, fmt.Errorf("a merge by key field must have a field tagged as %s and a Default field", mergeKey)
	}
	var defaultsByKey []acc
	for valueIndex := 0; valueIndex < vals.Len(); valueIndex++ {
		value := vals.Index(valueIndex)
		mergeKeyValue := value.FieldByName(key.Name).Interface()
		var found bool
		for accIndex, accRule := range defaultsByKey {
			if !reflect.DeepEqual(accRule.Key, mergeKeyValue) {
				continue
			}
			defaultsByKey[accIndex] = acc{
				Key:      accRule.Key,
				Defaults: append(accRule.Defaults, value.FieldByName(defaultFieldName).Interface()),
			}
			found = true
		}
		if !found {
			defaultsByKey = append(defaultsByKey, acc{
				Key:      mergeKeyValue,
				Defaults: []interface{}{value.FieldByName(defaultFieldName).Interface()},
			})
		}
	}
	keyValues := reflect.Zero(vals.Type())
	for _, confs := range defaultsByKey {
		merged, err := MergeConfs(confs.Defaults)
		if err != nil {
			return reflect.Value{}, err
		}

		keyValueP := reflect.New(valType)
		keyValue := keyValueP.Elem()

		keyValue.FieldByName(key.Name).Set(reflect.ValueOf(confs.Key))
		keyValue.FieldByName(defaultFieldName).Set(reflect.ValueOf(merged))

		keyValues = reflect.Append(keyValues, keyValue)
	}
	return keyValues, nil
}

func findKeyAndSpec(typ reflect.Type) (reflect.StructField, bool) {
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
