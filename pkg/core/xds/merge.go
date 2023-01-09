package xds

import (
	"encoding/json"
	"errors"
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

	result, err := newConf(reflect.TypeOf(confs[0]))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resultBytes, result); err != nil {
		return nil, err
	}

	// clear appendable slices, so we won't duplicate values of the last conf
	clearAppendSlices(reflect.ValueOf(result))
	for _, conf := range confs {
		// call .Elem() to unwrap interface{}
		appendSlices(reflect.ValueOf(result), reflect.ValueOf(&conf).Elem())
	}

	v := reflect.ValueOf(result).Elem().Interface()

	return v, nil
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
		fieldName := strVal.Type().Field(i).Name
		switch valField.Kind() {
		case reflect.Slice:
			if strings.HasPrefix(fieldName, appendSlicesPrefix) {
				valField.Set(reflect.Zero(valField.Type()))
			}
		case reflect.Struct:
			clearAppendSlices(valField)
		case reflect.Pointer:
			if valField.Elem().Kind() == reflect.Struct {
				clearAppendSlices(valField)
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

		fieldName := strDst.Type().Field(i).Name
		switch dstField.Kind() {
		case reflect.Slice:
			if strings.HasPrefix(fieldName, appendSlicesPrefix) {
				s := reflect.AppendSlice(dstField, srcField)
				dstField.Set(s)
			}
		case reflect.Struct:
			appendSlices(dstField, srcField)
		case reflect.Pointer:
			if dstField.Elem().Kind() == reflect.Struct {
				appendSlices(dstField, srcField)
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
