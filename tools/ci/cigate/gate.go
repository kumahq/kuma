package main

import (
	"fmt"
	"regexp"
	"sort"
)

type Job struct {
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
}

type Run struct {
	ID         int64  `json:"id"`
	Conclusion string `json:"conclusion"`
	CreatedAt  string `json:"created_at"`
	HTMLURL    string `json:"html_url"`
}

type requirement struct {
	label   string
	pattern *regexp.Regexp
}

var required = []requirement{
	{"test / test_unit", regexp.MustCompile(`^test / test_unit$`)},
	{"test / e2e ...", regexp.MustCompile(`^test / e2e`)},
	{"build_publish / digest-images", regexp.MustCompile(`^build_publish / digest-images$`)},
	{"build_publish / build-binaries", regexp.MustCompile(`^build_publish / build-binaries`)},
	{"build_publish / publish-binaries", regexp.MustCompile(`^build_publish / publish-binaries$`)},
	{"build_publish / build-images ...", regexp.MustCompile(`^build_publish / build-images`)},
	{"build_publish / publish-helm", regexp.MustCompile(`^build_publish / publish-helm$`)},
}

func Candidates(runs []Run) []Run {
	kept := []Run{}
	for _, run := range runs {
		if run.Conclusion == "success" {
			kept = append(kept, run)
		}
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].CreatedAt > kept[j].CreatedAt })

	return kept
}

func MissingJobs(jobs []Job) []string {
	missing := []string{}

	for _, want := range required {
		matched, allGreen := 0, true
		for _, job := range jobs {
			if want.pattern.MatchString(job.Name) {
				matched++
				if job.Conclusion != "success" {
					allGreen = false
				}
			}
		}
		if matched == 0 || !allGreen {
			missing = append(missing, want.label)
		}
	}

	return missing
}

func Gate(sha, workflow string, listRuns func() ([]Run, error), listJobs func(int64) ([]Job, error), log func(string)) (Run, error) {
	runs, err := listRuns()
	if err != nil {
		return Run{}, fmt.Errorf("gate could not list runs for SHA %s: %w", sha, err)
	}

	candidates := Candidates(runs)
	if len(candidates) == 0 {
		return Run{}, fmt.Errorf("No successful completed push run of %s found for tagged SHA %s. Every tag push requires a prior green branch push CI run for the tagged commit. Recovery: re-run the branch push CI on this SHA (re-tests and re-publishes the preview), then re-tag.", workflow, sha)
	}

	unread, inspected := 0, 0

	for _, run := range candidates {
		jobs, err := listJobs(run.ID)
		if err != nil {
			unread++
			log(fmt.Sprintf("::warning title=release_sha_gate::gate could not read the jobs of run %s: %s", run.HTMLURL, err))

			continue
		}

		inspected++
		missing := MissingJobs(jobs)
		if len(missing) == 0 {
			return run, nil
		}

		log(fmt.Sprintf("Run %s is missing required successful job(s): %s", run.HTMLURL, join(missing)))
	}

	if unread > 0 {
		return Run{}, fmt.Errorf("gate could not read the jobs of %d run(s) for SHA %s and found no green run among the %d it did read. Some of this is a GitHub API failure rather than a missing run. Recovery: re-run this job.", unread, sha, inspected)
	}

	return Run{}, fmt.Errorf("No push run of %s on SHA %s has every required job green. Recovery: re-run the branch push CI on this SHA, then re-tag.", workflow, sha)
}
