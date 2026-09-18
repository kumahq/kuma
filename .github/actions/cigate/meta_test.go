package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func pull(names []string, draft bool) *PullRequest {
	labels := []Label{}
	for _, name := range names {
		labels = append(labels, Label{Name: name})
	}

	built := &PullRequest{Labels: labels, Draft: &draft}
	built.Base.Ref = "master"

	return built
}

func TestDecisionsWritesLabelsAsJSONBesideTheDraftFlag(t *testing.T) {
	if got := Decisions([]string{"ci/skip-test"}, "true"); got != "json=[\"ci/skip-test\"]\ndraft=true" {
		t.Fatalf("Given one label on a draft, When written, Then both lines are set; got %q", got)
	}
	if got := Decisions(nil, ""); got != "json=[]\ndraft=" {
		t.Fatalf("Given no pull request, When written, Then the labels are an empty array; got %q", got)
	}
}

func TestDecisionsEscapesALabelNameThatWouldBreakTheJSON(t *testing.T) {
	line := Decisions([]string{`a"b`, "c\nd"}, "false")
	encoded, _, _ := strings.Cut(strings.TrimPrefix(line, "json="), "\ndraft=")
	var decoded []string
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatalf("Given a label needing escaping, When written, Then it survives a round trip; got %v", err)
	}
	if len(decoded) != 2 || decoded[0] != `a"b` {
		t.Fatalf("Then the names come back intact; got %v", decoded)
	}
}

func TestReadPullRequestDoesNotRetryAfterASuccess(t *testing.T) {
	calls := 0
	got, err := ReadPullRequest(func() (*PullRequest, error) {
		calls++

		return pull([]string{"ci/run-build"}, true), nil
	}, func(time.Duration) {})

	if err != nil || calls != 1 || !*got.Draft || len(got.Labels) != 1 {
		t.Fatalf("Given the read succeeds, When reading, Then it does not retry; got %v %v %d", got, err, calls)
	}
}

func TestReadPullRequestReturnsTheThirdAnswerAfterTwoFailures(t *testing.T) {
	calls := 0
	got, err := ReadPullRequest(func() (*PullRequest, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("502")
		}

		return pull(nil, false), nil
	}, func(time.Duration) {})

	if err != nil || calls != 3 || *got.Draft {
		t.Fatalf("Given two failures then a success, When reading, Then it returns the third; got %v %v %d", got, err, calls)
	}
}

func TestReadPullRequestRetriesABodyMissingLabelsOrDraft(t *testing.T) {
	for _, body := range []*PullRequest{{}, {Labels: []Label{}}, nil} {
		calls := 0
		_, err := ReadPullRequest(func() (*PullRequest, error) {
			calls++

			return body, nil
		}, func(time.Duration) {})

		if err == nil || calls != 3 {
			t.Fatalf("Given a body with no labels or draft, When reading, Then it retries and fails; got %v after %d", err, calls)
		}
	}
}

func TestReadPullRequestBacksOffLongerEachTime(t *testing.T) {
	waits := []time.Duration{}
	_, _ = ReadPullRequest(
		func() (*PullRequest, error) { return nil, errors.New("502") },
		func(d time.Duration) { waits = append(waits, d) },
	)

	if len(waits) != 2 || waits[0] != 5*time.Second || waits[1] != 10*time.Second {
		t.Fatalf("Given repeated failures, When retrying, Then it waits longer each time; got %v", waits)
	}
}

func TestHoldReturnsTheCauseWhenRefused(t *testing.T) {
	err := Hold(func(string) error { return errors.New("403 Forbidden") }, "abc", func(string) {})
	if err == nil || !strings.Contains(err.Error(), "403 Forbidden") {
		t.Fatalf("Given a refused hold, as on a fork, When held, Then it returns the cause; got %v", err)
	}
}

func meta(t *testing.T, event, base string, get func() (*PullRequest, error), post func(string) error) (int, []string, []string) {
	t.Helper()
	written, logs := []string{}, []string{}
	code := Meta(event, "7", "abc", base, get, post, func(time.Duration) {},
		func(d string) { written = append(written, d) },
		func(l string) { logs = append(logs, l) })

	return code, written, logs
}

func TestMetaRefusesARunStartedAgainstAnotherBase(t *testing.T) {
	code, written, logs := meta(t, "pull_request", "release-2.14",
		func() (*PullRequest, error) { return pull(nil, false), nil },
		func(string) error { return nil })

	if code != 1 || len(written) != 0 {
		t.Fatalf("Given the pull request now targets another base, When decided, Then it writes nothing and exits 1; got %d %v", code, written)
	}
	if !strings.Contains(logs[0], "now targets master") {
		t.Fatalf("Then it names both bases; got %v", logs)
	}
}

func TestMetaWritesTheDecisionsAndHoldsTheCheck(t *testing.T) {
	posted := ""
	code, written, _ := meta(t, "pull_request", "master",
		func() (*PullRequest, error) { return pull([]string{"ci/skip-test"}, true), nil },
		func(sha string) error { posted = sha; return nil })

	if code != 0 || len(written) != 1 || written[0] != "json=[\"ci/skip-test\"]\ndraft=true" || posted != "abc" {
		t.Fatalf("Given a pull request, When decided, Then it writes and holds; got %d %v %q", code, written, posted)
	}
}

func TestMetaOnAPushWritesEmptyDecisionsAndHoldsNothing(t *testing.T) {
	held := false
	code, written, _ := meta(t, "push", "master", nil, func(string) error { held = true; return nil })

	if code != 0 || written[0] != "json=[]\ndraft=" || held {
		t.Fatalf("Given a push, When decided, Then nothing is held; got %d %v %v", code, written, held)
	}
}

func TestMetaWritesNoOutputsWhenTheReadNeverSucceeds(t *testing.T) {
	code, written, logs := meta(t, "pull_request", "master",
		func() (*PullRequest, error) { return nil, errors.New("503") }, nil)

	if code != 1 || len(written) != 0 {
		t.Fatalf("Given the read never succeeds, When decided, Then it exits 1 writing nothing; got %d %v", code, written)
	}
	if !strings.Contains(logs[0], "could not read pull request 7 after 3 attempts") {
		t.Fatalf("Then it says how many attempts it made; got %v", logs)
	}
}

func TestMetaKeepsTheDecisionsWhenTheHoldFails(t *testing.T) {
	code, written, logs := meta(t, "pull_request", "master",
		func() (*PullRequest, error) { return pull(nil, false), nil },
		func(string) error { return errors.New("403") })

	if code != 0 || len(written) != 1 {
		t.Fatalf("Given a failed hold, When decided, Then the decisions still stand; got %d %v", code, written)
	}
	if !strings.Contains(logs[len(logs)-1], "::warning title=meta::") {
		t.Fatalf("Then it warns; got %v", logs)
	}
}
