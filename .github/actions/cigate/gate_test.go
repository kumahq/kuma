package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func green() []Job {
	return []Job{
		{"test / test_unit", "success"},
		{"test / e2e (default, v1.35, amd64)", "success"},
		{"test / e2e (gatewayapi, v1.35, amd64)", "success"},
		{"build_publish / digest-images", "success"},
		{"build_publish / build-binaries (amd64)", "success"},
		{"build_publish / publish-binaries", "success"},
		{"build_publish / build-images (amd64)", "success"},
		{"build_publish / publish-helm", "success"},
	}
}

func replace(name, conclusion string) []Job {
	jobs := green()
	for i := range jobs {
		if jobs[i].Name == name {
			jobs[i].Conclusion = conclusion
		}
	}

	return jobs
}

func drop(prefix string) []Job {
	jobs := []Job{}
	for _, job := range green() {
		if !strings.HasPrefix(job.Name, prefix) {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func TestMissingJobsAcceptsAnAllGreenRun(t *testing.T) {
	if missing := MissingJobs(green()); len(missing) != 0 {
		t.Fatalf("Given every required job green, When checked, Then nothing is missing; got %v", missing)
	}
}

func TestMissingJobsReportsARenamedJobByItsLabel(t *testing.T) {
	jobs := green()
	jobs[0].Name = "test / unit_test"

	if missing := MissingJobs(jobs); !reflect.DeepEqual(missing, []string{"test / test_unit"}) {
		t.Fatalf("Given a renamed job, When checked, Then its label is missing; got %v", missing)
	}
}

func TestMissingJobsRejectsAnyLegThatIsNotSuccess(t *testing.T) {
	for _, conclusion := range []string{"failure", "cancelled", "skipped", ""} {
		missing := MissingJobs(replace("test / e2e (gatewayapi, v1.35, amd64)", conclusion))
		if !reflect.DeepEqual(missing, []string{"test / e2e ..."}) {
			t.Fatalf("Given an e2e leg %q beside a green sibling, When checked, Then the family is missing; got %v", conclusion, missing)
		}
	}
}

func TestMissingJobsRejectsATagRunWhichRunsNeitherFamily(t *testing.T) {
	want := []string{"test / test_unit", "test / e2e ..."}
	if missing := MissingJobs(drop("test / ")); !reflect.DeepEqual(missing, want) {
		t.Fatalf("Given a tag run, When checked, Then both test families are missing; got %v", missing)
	}
}

func TestMissingJobsAcceptsTheOlderKongMeshE2ENaming(t *testing.T) {
	jobs := drop("test / e2e")
	jobs = append(jobs,
		Job{"test / test_e2e_env (calico, v1.35) / e2e (0)", "success"},
		Job{"test / test_e2e (default) / e2e (1)", "success"},
	)

	if missing := MissingJobs(jobs); len(missing) != 0 {
		t.Fatalf("Given release branch job names, When checked, Then the e2e family still matches; got %v", missing)
	}
}

func TestMissingJobsStillRejectsAFailedLegUnderTheOlderNaming(t *testing.T) {
	jobs := drop("test / e2e")
	jobs = append(jobs,
		Job{"test / test_e2e_env (calico, v1.35) / e2e (0)", "success"},
		Job{"test / test_e2e_env (flannel, v1.35) / e2e (1)", "failure"},
	)

	if missing := MissingJobs(jobs); len(missing) != 1 {
		t.Fatalf("Given a failed leg under the older naming, When checked, Then the family is missing; got %v", missing)
	}
}

func TestMissingJobsRejectsAnEmptyJobList(t *testing.T) {
	if missing := MissingJobs(nil); len(missing) != len(required) {
		t.Fatalf("Given no jobs at all, When checked, Then every requirement is missing; got %v", missing)
	}
}

func TestCandidatesKeepsSuccessfulRunsNewestFirst(t *testing.T) {
	runs := []Run{
		{ID: 1, Conclusion: "success", CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: 2, Conclusion: "failure", CreatedAt: "2026-01-03T00:00:00Z"},
		{ID: 3, Conclusion: "success", CreatedAt: "2026-01-02T00:00:00Z"},
	}

	got := Candidates(runs)
	if len(got) != 2 || got[0].ID != 3 || got[1].ID != 1 {
		t.Fatalf("Given a failed run among successes, When selected, Then only successes remain newest first; got %v", got)
	}
}

func TestCandidatesSortsARunWithNoTimestampLast(t *testing.T) {
	runs := []Run{
		{ID: 1, Conclusion: "success", CreatedAt: "2026-05-05T00:00:00Z"},
		{ID: 2, Conclusion: "success"},
	}

	if got := Candidates(runs); got[0].ID != 1 {
		t.Fatalf("Given a run with no created_at, When selected, Then it does not sort first; got %v", got)
	}
}

func run(id int64, created string) Run {
	return Run{ID: id, Conclusion: "success", CreatedAt: created, HTMLURL: "https://x/" + created}
}

func TestGateRefusesWhenNoRunSucceeded(t *testing.T) {
	_, err := Gate("abc", "w", func() ([]Run, error) { return nil, nil }, nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "No successful completed push run") {
		t.Fatalf("Given no successful run, When gated, Then it says none was found; got %v", err)
	}
}

func TestGateFallsThroughANewerTagRunToAnOlderGreenBranchRun(t *testing.T) {
	logged := []string{}
	got, err := Gate("abc", "w",
		func() ([]Run, error) { return []Run{run(100, "2026-01-01"), run(200, "2026-01-02")}, nil },
		func(id int64) ([]Job, error) {
			if id == 200 {
				return nil, nil
			}

			return green(), nil
		},
		func(line string) { logged = append(logged, line) },
	)
	if err != nil || got.ID != 100 {
		t.Fatalf("Given a newer tag run and an older green branch run, When gated, Then it passes on the older; got %v %v", got, err)
	}
	if len(logged) != 1 || !strings.Contains(logged[0], "https://x/2026-01-02 is missing") {
		t.Fatalf("Then it says what the newer run was missing; got %v", logged)
	}
}

func TestGateReportsAnAPIFailureRatherThanAMissingRun(t *testing.T) {
	_, err := Gate("abc", "w",
		func() ([]Run, error) { return []Run{run(100, "2026-01-01"), run(200, "2026-01-02")}, nil },
		func(int64) ([]Job, error) { return nil, errors.New("502") },
		func(string) {},
	)
	if err == nil || !strings.Contains(err.Error(), "GitHub API failure") {
		t.Fatalf("Given every job query fails, When gated, Then it blames the API; got %v", err)
	}
}

func TestGateStillBlamesTheAPIWhenOneRunWasReadable(t *testing.T) {
	_, err := Gate("abc", "w",
		func() ([]Run, error) { return []Run{run(200, "2026-01-02"), run(100, "2026-01-01")}, nil },
		func(id int64) ([]Job, error) {
			if id == 200 {
				return nil, errors.New("502")
			}

			return nil, nil
		},
		func(string) {},
	)
	if err == nil || !strings.Contains(err.Error(), "could not read the jobs of 1 run(s)") {
		t.Fatalf("Given one unreadable run and one incomplete run, When gated, Then it blames the API; got %v", err)
	}
}

func TestGateSaysNoRunIsGreenWhenEveryRunWasReadable(t *testing.T) {
	_, err := Gate("abc", "w",
		func() ([]Run, error) { return []Run{run(100, "2026-01-01")}, nil },
		func(int64) ([]Job, error) { return nil, nil },
		func(string) {},
	)
	if err == nil || !strings.Contains(err.Error(), "has every required job green") {
		t.Fatalf("Given every run readable and incomplete, When gated, Then it says no run is green; got %v", err)
	}
}

func TestGateSurfacesAFailureToListRuns(t *testing.T) {
	_, err := Gate("abc", "w", func() ([]Run, error) { return nil, errors.New("503") }, nil, func(string) {})
	if err == nil || !strings.Contains(err.Error(), "could not list runs") {
		t.Fatalf("Given the run list cannot be read, When gated, Then it says so; got %v", err)
	}
}
