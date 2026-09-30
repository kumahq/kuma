package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type PullRequest struct {
	Labels []Label `json:"labels"`
	Draft  *bool   `json:"draft"`
	Base   struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

const attempts = 3

func Decisions(labels []string, draft string) string {
	encoded, _ := json.Marshal(append([]string{}, labels...))

	return fmt.Sprintf("json=%s\ndraft=%s", encoded, draft)
}

func ReadPullRequest(get func() (*PullRequest, error), wait func(time.Duration)) (*PullRequest, error) {
	var last error

	for attempt := 1; attempt <= attempts; attempt++ {
		pull, err := get()
		switch {
		case err != nil:
			last = err
		case pull == nil || pull.Labels == nil || pull.Draft == nil:
			last = fmt.Errorf("pull request payload has no labels or draft")
		default:
			return pull, nil
		}

		if !Transient(last) {
			return nil, last
		}

		if attempt < attempts {
			wait(time.Duration(attempt) * 5 * time.Second)
		}
	}

	return nil, last
}

func BaseHeld(base, now string) error {
	if now == base {
		return nil
	}

	return fmt.Errorf("this run was started against %s and the pull request now targets %s, so nothing it did describes it. Push a commit to start a run against %s.", base, now, now)
}

type Refusal struct {
	Status string
	Code   int
	URL    string
}

func (r Refusal) Error() string { return fmt.Sprintf("%s for %s", r.Status, r.URL) }

func Transient(err error) bool {
	refusal := Refusal{}
	if !errors.As(err, &refusal) {
		return true
	}

	return refusal.Code >= 500 || refusal.Code == http.StatusTooManyRequests
}

func Hold(post func(string) error, sha string, wait func(time.Duration), log func(string)) error {
	var last error

	for attempt := 1; attempt <= attempts; attempt++ {
		last = post(sha)
		switch {
		case last == nil:
			log(fmt.Sprintf("held distributions on %s until this run reports", sha))

			return nil
		case !Transient(last):
			return last
		}

		if attempt < attempts {
			wait(time.Duration(attempt) * 5 * time.Second)
		}
	}

	return last
}

func Meta(event, number, sha, base string, get func() (*PullRequest, error), post func(string) error, wait func(time.Duration), write func(string), log func(string)) int {
	if event != "pull_request" {
		decided := Decisions(nil, "")
		write(decided)
		log(decided)

		return 0
	}

	pull, err := ReadPullRequest(get, wait)
	if err != nil {
		log(fmt.Sprintf("::error title=meta::could not read pull request %s after %d attempts: %s", number, attempts, err))

		return 1
	}

	if base != "" {
		if err := BaseHeld(base, pull.Base.Ref); err != nil {
			log(fmt.Sprintf("::error title=meta::%s", err))

			return 1
		}
	}

	names := make([]string, 0, len(pull.Labels))
	for _, label := range pull.Labels {
		names = append(names, label.Name)
	}

	decided := Decisions(names, strconv.FormatBool(*pull.Draft))
	write(decided)
	log(decided)

	if err := Hold(post, sha, wait, log); err != nil {
		log(fmt.Sprintf("::warning title=meta::could not hold the distributions check on %s, so a result from an earlier run stands until this one finishes: %s", sha, err))
	}

	return 0
}
