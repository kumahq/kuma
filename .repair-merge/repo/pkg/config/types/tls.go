package types

import (
	"crypto/tls"
	"fmt"
	"sort"
)

func TLSVersion(minVersion string) (uint16, error) {
	if minVersion == "" {
		return 0, nil
	}
	v, ok := versions[minVersion]
	if !ok {
		return 0, fmt.Errorf("unsupported tls version: %s supported versions:%v", minVersion, versionNames)
	}
	return v, nil
}

func TLSCiphers(ciphers []string) ([]uint16, error) {
	if len(ciphers) == 0 {
		return nil, nil
	}
	var res []uint16
	for _, cipher := range ciphers {
		v, ok := secureCiphers[cipher]
		if !ok {
			v, ok = insecureCiphers[cipher]
			if !ok {
				return nil, fmt.Errorf("unsupported tls cipher: %s supported ciphers insecure:%v, secure:%v", cipher, insecureCiphersNames, secureCiphersNames)
			}
		}
		res = append(res, v)
	}
	return res, nil
}

var versions = map[string]uint16{
	"TLSv1_2": tls.VersionTLS12,
	"TLSv1_3": tls.VersionTLS13,
}

var (
	versionNames         []string
	secureCiphers        map[string]uint16
	secureCiphersNames   []string
	insecureCiphers      map[string]uint16
	insecureCiphersNames []string
)

func init() {
	secureCiphers = map[string]uint16{}
	for _, v := range tls.CipherSuites() {
		secureCiphers[v.Name] = v.ID
		secureCiphersNames = append(secureCiphersNames, v.Name)
	}
	insecureCiphers = map[string]uint16{}
	for _, v := range tls.InsecureCipherSuites() {
		insecureCiphers[v.Name] = v.ID
		insecureCiphersNames = append(insecureCiphersNames, v.Name)
	}
	for k := range versions {
		versionNames = append(versionNames, k)
	}
	sort.Strings(versionNames)
}
