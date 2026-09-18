package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type Need struct {
	Result string `json:"result"`
}

var gates = []string{"build_check", "check"}

func join(parts []string) string { return strings.Join(parts, ", ") }

func Verdict(results map[string]string, isDraft bool) error {
	broken := slices.DeleteFunc(slices.Sorted(maps.Keys(results)), func(job string) bool {
		return results[job] != "failure" && results[job] != "cancelled"
	})

	if len(broken) > 0 {
		return fmt.Errorf("these jobs failed or were cancelled: %s", join(broken))
	}

	silent := slices.DeleteFunc(slices.Clone(gates), func(job string) bool { return results[job] != "" })
	if len(silent) > 0 {
		return fmt.Errorf("this run reports nothing about %s, which it has to check before it can pass. They belong in this job's needs.", join(silent))
	}

	skipped := slices.DeleteFunc(slices.Clone(gates), func(job string) bool { return results[job] != "skipped" })

	if len(skipped) > 0 {
		if isDraft {
			return fmt.Errorf("this pull request is a draft, so %s did not run and this check has tested nothing. Mark it ready for review to test it.", join(skipped))
		}

		return fmt.Errorf("this run skipped %s, but the pull request is not a draft, so these results do not describe it. Push a commit to start a run that tests it - re-running this one replays the event it was started with and skips them again.", join(skipped))
	}

	return nil
}

func Halt(rawNeeds, isDraft, base string, liveBase func() (string, error), log func(string)) int {
	needs := map[string]Need{}
	if err := json.Unmarshal([]byte(rawNeeds), &needs); err != nil {
		log(fmt.Sprintf("::error title=distributions::could not read the results of this run's dependencies: %s", err))

		return 1
	}

	results := map[string]string{}
	for job, need := range needs {
		results[job] = need.Result
	}

	encoded, _ := json.Marshal(results)
	log(fmt.Sprintf("results: %s", encoded))

	if base != "" {
		now, err := liveBase()
		if err != nil {
			log(fmt.Sprintf("::error title=distributions::could not read which branch this pull request targets: %s", err))

			return 1
		}
		if now != base {
			log(fmt.Sprintf("::error title=distributions::this run was started against %s and the pull request now targets %s, so nothing it did describes it. Push a commit to start a run against %s.", base, now, now))

			return 1
		}
	}

	if err := Verdict(results, isDraft == "true"); err != nil {
		log(fmt.Sprintf("::error title=distributions::%s", err))

		return 1
	}

	log("All dependent jobs succeeded")

	return 0
}
