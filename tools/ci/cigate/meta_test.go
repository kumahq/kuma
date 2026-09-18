package main

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func pull(names []string, draft bool) *PullRequest {
	labels := []Label{}
	for _, name := range names {
		labels = append(labels, Label{Name: name})
	}

	return &PullRequest{Labels: labels, Draft: &draft}
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
	var decoded []string
	if err := json.Unmarshal([]byte(line[len("json="):len(line)-len("\ndraft=false")]), &decoded); err != nil {
		t.Fatalf("Given a label needing escaping, When written, Then it survives a round trip; got %v", err)
	}
	if len(decoded) != 2 || decoded[0] != `a"b` {
		t.Fatalf("Then the names come back intact; got %v", decoded)
	}
}

func TestReadPullRequestDoesNotRetryAfterASuccess(t *testing.T) {
	calls := 0
	labels, draft, err := ReadPullRequest(func() (*PullRequest, error) {
		calls++

		return pull([]string{"ci/run-build"}, true), nil
	}, func(time.Duration) {})

	if err != nil || calls != 1 || draft != "true" || len(labels) != 1 {
		t.Fatalf("Given the read succeeds, When reading, Then it does not retry; got %v %v %v %d", labels, draft, err, calls)
	}
}

func TestReadPullRequestReturnsTheThirdAnswerAfterTwoFailures(t *testing.T) {
	calls := 0
	_, draft, err := ReadPullRequest(func() (*PullRequest, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("502")
		}

		return pull(nil, false), nil
	}, func(time.Duration) {})

	if err != nil || calls != 3 || draft != "false" {
		t.Fatalf("Given two failures then a success, When reading, Then it returns the third; got %v %v %d", draft, err, calls)
	}
}

func TestReadPullRequestRetriesABodyMissingLabelsOrDraft(t *testing.T) {
	for _, body := range []*PullRequest{{}, {Labels: []Label{}}, nil} {
		calls := 0
		_, _, err := ReadPullRequest(func() (*PullRequest, error) {
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
	_, _, _ = ReadPullRequest(
		func() (*PullRequest, error) { return nil, errors.New("502") },
		func(d time.Duration) { waits = append(waits, d) },
	)

	if len(waits) != 2 || waits[0] != 5*time.Second || waits[1] != 10*time.Second {
		t.Fatalf("Given repeated failures, When retrying, Then it waits longer each time; got %v", waits)
	}
}

func TestHoldWarnsAndCarriesOnWhenRefused(t *testing.T) {
	lines := []string{}
	if Hold(func(string) error { return errors.New("403 Forbidden") }, "abc", func(l string) { lines = append(lines, l) }) {
		t.Fatal("Given a refused hold, as on a fork, When held, Then it reports failure")
	}
	if !contains(lines[0], "::warning title=meta::could not hold") || !contains(lines[0], "403 Forbidden") {
		t.Fatalf("Then it warns with the cause; got %v", lines)
	}
}

func meta(t *testing.T, event string, get func() (*PullRequest, error), post func(string) error) (int, []string, []string) {
	t.Helper()
	written, logs := []string{}, []string{}
	code := Meta(event, "7", "abc", get, post, func(time.Duration) {},
		func(d string) { written = append(written, d) },
		func(l string) { logs = append(logs, l) })

	return code, written, logs
}

func TestMetaWritesTheDecisionsAndHoldsTheCheck(t *testing.T) {
	posted := ""
	code, written, _ := meta(t, "pull_request",
		func() (*PullRequest, error) { return pull([]string{"ci/skip-test"}, true), nil },
		func(sha string) error { posted = sha; return nil })

	if code != 0 || len(written) != 1 || written[0] != "json=[\"ci/skip-test\"]\ndraft=true" || posted != "abc" {
		t.Fatalf("Given a pull request, When decided, Then it writes and holds; got %d %v %q", code, written, posted)
	}
}

func TestMetaOnAPushWritesEmptyDecisionsAndHoldsNothing(t *testing.T) {
	held := false
	code, written, _ := meta(t, "push", nil, func(string) error { held = true; return nil })

	if code != 0 || written[0] != "json=[]\ndraft=" || held {
		t.Fatalf("Given a push, When decided, Then nothing is held; got %d %v %v", code, written, held)
	}
}

func TestMetaWritesNoOutputsWhenTheReadNeverSucceeds(t *testing.T) {
	code, written, logs := meta(t, "pull_request",
		func() (*PullRequest, error) { return nil, errors.New("503") }, nil)

	if code != 1 || len(written) != 0 {
		t.Fatalf("Given the read never succeeds, When decided, Then it exits 1 writing nothing; got %d %v", code, written)
	}
	if !contains(logs[0], "could not read pull request 7 after 3 attempts") {
		t.Fatalf("Then it says how many attempts it made; got %v", logs)
	}
}

func TestMetaKeepsTheDecisionsWhenTheHoldFails(t *testing.T) {
	code, written, logs := meta(t, "pull_request",
		func() (*PullRequest, error) { return pull(nil, false), nil },
		func(string) error { return errors.New("403") })

	if code != 0 || len(written) != 1 {
		t.Fatalf("Given a failed hold, When decided, Then the decisions still stand; got %d %v", code, written)
	}
	if !contains(logs[len(logs)-1], "::warning title=meta::") {
		t.Fatalf("Then it warns; got %v", logs)
	}
}
