package dns

import "strings"

func DnsNameToKumaCompliant(name string) (string, error) {
	// the request might be of the form:
	//  service-name.namespace.something-else.mesh.
	// it will always end with a dot, as specified by https://tools.ietf.org/html/rfc1034#section-3.1
	//  `Since a complete domain name ends with the root label, this leads to a printed form which ends in a dot.`
	countDots := strings.Count(name, ".")
	toReplace := countDots
	switch countDots {
	case 0:
		return name, nil
	case 1:
		if name[len(name)-1] == '.' {
			return name, nil
		}
	default:
		if name[len(name)-1] == '.' {
			toReplace--
		}
	}

	return strings.Replace(name, ".", "_", toReplace-1), nil
}
