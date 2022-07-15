package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/asaskevich/govalidator"
)

func parseFileToHashMap(file string) (map[string]interface{}, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return parseBytesToHashMap(contents)
}

func parseBytesToHashMap(bytes []byte) (map[string]interface{}, error) {
	var parsed map[string]interface{}
	err := json.Unmarshal(bytes, &parsed)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func guardIpv6Host(host string) string {
	safeHost := host
	if govalidator.IsIPv6(host) {
		if !(strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")) {
			safeHost = "[" + host + "]"
		}
	}

	return safeHost
}
