package main

import "testing"

func needs(over map[string]string) map[string]Need {
	base := map[string]string{
		"release_sha_gate": "success", "meta": "success", "refuse_fork_publish": "skipped",
		"build_check": "success", "build_publish": "skipped", "check": "success",
		"test": "success", "provenance": "skipped",
	}
	for job, result := range over {
		base[job] = result
	}

	built := map[string]Need{}
	for job, result := range base {
		built[job] = Need{Result: result}
	}

	return built
}

func TestVerdictPassesWhenEveryDependencySucceededOrSkipped(t *testing.T) {
	if _, err := Verdict(needs(nil), false); err != nil {
		t.Fatalf("Given a healthy run, When judged, Then it passes; got %v", err)
	}
}

func TestVerdictNamesEveryBrokenDependencyInOrder(t *testing.T) {
	_, err := Verdict(needs(map[string]string{"check": "failure", "test": "cancelled"}), false)
	if err == nil || err.Error() != "these jobs failed or were cancelled: check, test" {
		t.Fatalf("Given a failure and a cancellation, When judged, Then both are named in order; got %v", err)
	}
}

func TestVerdictAllowsADraftToSkipItsGates(t *testing.T) {
	if _, err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "skipped"}), true); err != nil {
		t.Fatalf("Given a draft that skipped its gates, When judged, Then it passes; got %v", err)
	}
}

func TestVerdictRefusesSkippedGatesOnAPullRequestThatIsNoLongerADraft(t *testing.T) {
	_, err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "skipped"}), false)
	if err == nil || !contains(err.Error(), "skipped build_check, check") {
		t.Fatalf("Given a ready pull request whose gates skipped, When judged, Then it refuses; got %v", err)
	}
	if !contains(err.Error(), "re-running this one replays the event") {
		t.Fatalf("Then it says a re-run is not the recovery; got %v", err)
	}
}

func TestVerdictPrefersAFailureOverTheDraftAllowance(t *testing.T) {
	_, err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "failure"}), true)
	if err == nil || !contains(err.Error(), "failed or were cancelled: check") {
		t.Fatalf("Given a failure on a draft, When judged, Then the failure wins; got %v", err)
	}
}

func TestHaltFailsClosedOnUnreadableNeeds(t *testing.T) {
	for _, raw := range []string{"", "not json", "[]"} {
		lines := []string{}
		if code := Halt(raw, "false", func(l string) { lines = append(lines, l) }); code != 1 {
			t.Fatalf("Given needs %q, When halted, Then it exits 1; got %d", raw, code)
		}
		if len(lines) == 0 || !contains(lines[0], "::error title=distributions::") {
			t.Fatalf("Then it annotates the failure; got %v", lines)
		}
	}
}

func TestHaltExitsOneAndAnnotatesASkippedGate(t *testing.T) {
	lines := []string{}
	code := Halt(`{"build_check":{"result":"skipped"},"check":{"result":"success"}}`, "false", func(l string) { lines = append(lines, l) })
	if code != 1 {
		t.Fatalf("Given a skipped gate on a ready pull request, When halted, Then it exits 1; got %d", code)
	}
	if !contains(lines[0], "results: ") || !contains(lines[1], "::error title=distributions::") {
		t.Fatalf("Then it prints the results and the error; got %v", lines)
	}
}

func TestHaltExitsZeroOnAHealthyRun(t *testing.T) {
	lines := []string{}
	if code := Halt(`{"check":{"result":"success"}}`, "", func(l string) { lines = append(lines, l) }); code != 0 {
		t.Fatalf("Given a healthy run, When halted, Then it exits 0; got %d", code)
	}
	if lines[len(lines)-1] != "All dependent jobs succeeded" {
		t.Fatalf("Then it says so; got %v", lines)
	}
}
