package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Need struct {
	Result string `json:"result"`
}

var gates = []string{"build_check", "check", "test"}

func join(parts []string) string { return strings.Join(parts, ", ") }

func Verdict(needs map[string]Need, isDraft bool) (map[string]string, error) {
	results := map[string]string{}
	for job, need := range needs {
		results[job] = need.Result
	}

	broken := []string{}
	for job, result := range results {
		if result == "failure" || result == "cancelled" {
			broken = append(broken, job)
		}
	}
	sort.Strings(broken)

	if len(broken) > 0 {
		return results, fmt.Errorf("these jobs failed or were cancelled: %s", join(broken))
	}

	skipped := []string{}
	for _, job := range gates {
		if results[job] == "skipped" {
			skipped = append(skipped, job)
		}
	}

	if !isDraft && len(skipped) > 0 {
		return results, fmt.Errorf("this run skipped %s, but the pull request is not a draft, so these results do not describe it. Push a commit to start a run that tests it - re-running this one replays the event it was started with and skips them again.", join(skipped))
	}

	return results, nil
}

func Halt(rawNeeds, isDraft string, log func(string)) int {
	needs := map[string]Need{}
	if err := json.Unmarshal([]byte(rawNeeds), &needs); err != nil {
		log(fmt.Sprintf("::error title=distributions::could not read the results of this run's dependencies: %s", err))

		return 1
	}

	results, err := Verdict(needs, isDraft == "true")

	encoded, _ := json.Marshal(results)
	log(fmt.Sprintf("results: %s", encoded))

	if err != nil {
		log(fmt.Sprintf("::error title=distributions::%s", err))

		return 1
	}

	log("All dependent jobs succeeded")

	return 0
}
