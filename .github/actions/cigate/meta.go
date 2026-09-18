package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Labels []Label `json:"labels"`
	Draft  *bool   `json:"draft"`
}

const attempts = 3

func Decisions(labels []string, draft string) string {
	if labels == nil {
		labels = []string{}
	}
	encoded, _ := json.Marshal(labels)

	return fmt.Sprintf("json=%s\ndraft=%s", encoded, draft)
}

func ReadPullRequest(get func() (*PullRequest, error), wait func(time.Duration)) ([]string, string, error) {
	var last error

	for attempt := 1; attempt <= attempts; attempt++ {
		pull, err := get()
		switch {
		case err != nil:
			last = err
		case pull == nil || pull.Labels == nil || pull.Draft == nil:
			last = fmt.Errorf("pull request payload has no labels or draft")
		default:
			names := make([]string, 0, len(pull.Labels))
			for _, label := range pull.Labels {
				names = append(names, label.Name)
			}

			return names, strconv.FormatBool(*pull.Draft), nil
		}

		if attempt < attempts {
			wait(time.Duration(attempt) * 5 * time.Second)
		}
	}

	return nil, "", last
}

func Hold(post func(string) error, sha string, log func(string)) bool {
	if err := post(sha); err != nil {
		log(fmt.Sprintf("::warning title=meta::could not hold the distributions check on %s, so a result from an earlier run stands until this one finishes: %s", sha, err))

		return false
	}

	log(fmt.Sprintf("held distributions on %s until this run reports", sha))

	return true
}

func Meta(event, number, sha string, get func() (*PullRequest, error), post func(string) error, wait func(time.Duration), write func(string), log func(string)) int {
	if event != "pull_request" {
		decided := Decisions(nil, "")
		write(decided)
		log(decided)

		return 0
	}

	labels, draft, err := ReadPullRequest(get, wait)
	if err != nil {
		log(fmt.Sprintf("::error title=meta::could not read pull request %s after %d attempts: %s", number, attempts, err))

		return 1
	}

	decided := Decisions(labels, draft)
	write(decided)
	log(decided)
	Hold(post, sha, log)

	return 0
}
