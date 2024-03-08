package leastcommonversion

import (
	"sort"

	"github.com/Masterminds/semver/v3"
)

type Input struct {
	Name          string     `json:"name"`
	Current       string     `json:"current"`
	FixedVersions [][]string `json:"fixedVersions"`
}

func Deduct(in *Input) (string, error) {
	current := semver.MustParse(in.Current)
	collections := parseCollections(in.FixedVersions)

	leastVersionPerCVE := []*semver.Version{}
	for _, c := range collections {
		filtered := filterNewerThanCurrent(c, current)
		if len(filtered) == 0 {
			// there is no version to update for the CVE
			continue
		}
		leastVersionPerCVE = append(leastVersionPerCVE, filtered[0])
	}

	if len(leastVersionPerCVE) == 0 {
		return "null", nil
	}

	// newest version out of all the least versions is the least common version to satisfy all the CVEs
	return newest(leastVersionPerCVE).String(), nil
}

func newest(c []*semver.Version) *semver.Version {
	sort.Stable(semver.Collection(c))
	return c[len(c)-1]
}

func parseCollections(ss [][]string) []semver.Collection {
	collections := []semver.Collection{}
	for _, s := range ss {
		collections = append(collections, parseCollection(s))
	}
	return collections
}

func parseCollection(ss []string) semver.Collection {
	collection := semver.Collection{}
	for _, s := range ss {
		collection = append(collection, semver.MustParse(s))
	}
	sort.Stable(collection)
	return collection
}

func filterNewerThanCurrent(c semver.Collection, current *semver.Version) semver.Collection {
	result := semver.Collection{}
	for _, v := range c {
		if v.GreaterThan(current) {
			result = append(result, v)
		}
	}
	return result
}
