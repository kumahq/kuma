package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

const maxPages = 50

var nextLink = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func api(method, url, token string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/vnd.github+json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= http.StatusBadRequest {
		response.Body.Close()

		return nil, fmt.Errorf("%s for %s", response.Status, url)
	}

	return response, nil
}

func paged[T any](url, token, key string, send func(string) (*http.Response, error)) ([]T, error) {
	collected := []T{}

	for page := 0; url != ""; page++ {
		if page >= maxPages {
			return nil, fmt.Errorf("more than %d pages", maxPages)
		}

		response, err := send(url)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, err
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(body, &envelope); err != nil {
			return nil, err
		}
		raw, ok := envelope[key]
		if !ok {
			return nil, fmt.Errorf("no %s array in the response from %s", key, url)
		}

		var items []T
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		collected = append(collected, items...)

		url = ""
		if match := nextLink.FindStringSubmatch(response.Header.Get("Link")); match != nil {
			url = match[1]
		}
	}

	return collected, nil
}

func postCheckRun(repo, token string) func(string) error {
	return func(head string) error {
		payload, _ := json.Marshal(map[string]string{"name": "distributions", "head_sha": head, "status": "in_progress"})
		response, err := api(http.MethodPost, fmt.Sprintf("https://api.github.com/repos/%s/check-runs", repo), token, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		defer response.Body.Close()

		return nil
	}
}

func main() {
	log := func(line string) { fmt.Println(line) }

	if len(os.Args) < 2 {
		log("::error::cigate needs a subcommand: release-gate, hold, halt or meta")
		os.Exit(1)
	}

	token := os.Getenv("GH_TOKEN")
	repo := os.Getenv("GH_REPO")
	send := func(url string) (*http.Response, error) { return api(http.MethodGet, url, token, nil) }

	switch os.Args[1] {
	case "halt":
		os.Exit(Halt(os.Getenv("NEEDS"), os.Getenv("IS_DRAFT"), log))

	case "hold":
		if err := Hold(postCheckRun(repo, token), os.Getenv("HEAD_SHA"), log); err != nil {
			log(fmt.Sprintf("::error title=hold::could not hold the distributions check, so this pull request would keep a result that no longer describes it: %s", err))
			os.Exit(1)
		}

	case "meta":
		number, sha := os.Getenv("PR"), os.Getenv("HEAD_SHA")
		output, err := os.OpenFile(os.Getenv("GITHUB_OUTPUT"), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			log(fmt.Sprintf("::error title=meta::could not open GITHUB_OUTPUT: %s", err))
			os.Exit(1)
		}

		os.Exit(Meta(
			os.Getenv("GITHUB_EVENT_NAME"), number, sha, os.Getenv("BASE_REF"),
			func() (*PullRequest, error) {
				response, err := send(fmt.Sprintf("https://api.github.com/repos/%s/pulls/%s", repo, number))
				if err != nil {
					return nil, err
				}
				defer response.Body.Close()
				pull := &PullRequest{}

				return pull, json.NewDecoder(response.Body).Decode(pull)
			},
			postCheckRun(repo, token),
			time.Sleep,
			func(decided string) { fmt.Fprintln(output, decided) },
			log,
		))

	case "release-gate":
		sha, workflow := os.Getenv("SHA"), os.Getenv("WORKFLOW")
		run, err := Gate(sha, workflow,
			func() ([]Run, error) {
				return paged[Run](fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/%s/runs?head_sha=%s&status=completed&event=push&per_page=100", repo, workflow, sha), token, "workflow_runs", send)
			},
			func(id int64) ([]Job, error) {
				return paged[Job](fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d/jobs?per_page=100", repo, id), token, "jobs", send)
			},
			log,
		)
		if err != nil {
			log(fmt.Sprintf("::error title=release_sha_gate::%s", err))
			os.Exit(1)
		}
		log(fmt.Sprintf("Trusted source run: %s", run.HTMLURL))

	default:
		log(fmt.Sprintf("::error::cigate does not know the subcommand %q", os.Args[1]))
		os.Exit(1)
	}
}
