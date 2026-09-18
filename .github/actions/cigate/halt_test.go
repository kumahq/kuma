package main

import (
	"maps"
	"strings"
	"testing"
)

func needs(over map[string]string) map[string]string {
	base := map[string]string{
		"release_sha_gate": "success", "meta": "success", "refuse_fork_publish": "skipped",
		"build_check": "success", "build_publish": "skipped", "check": "success",
		"test": "success", "provenance": "skipped",
	}
	maps.Copy(base, over)

	return base
}

func TestVerdictPassesWhenEveryDependencySucceededOrSkipped(t *testing.T) {
	if err := Verdict(needs(nil), false); err != nil {
		t.Fatalf("Given a healthy run, When judged, Then it passes; got %v", err)
	}
}

func TestVerdictNamesEveryBrokenDependencyInOrder(t *testing.T) {
	err := Verdict(needs(map[string]string{"check": "failure", "test": "cancelled"}), false)
	if err == nil || err.Error() != "these jobs failed or were cancelled: check, test" {
		t.Fatalf("Given a failure and a cancellation, When judged, Then both are named in order; got %v", err)
	}
}

func TestVerdictRefusesADraftThatSkippedItsGates(t *testing.T) {
	err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "skipped"}), true)
	if err == nil || !strings.Contains(err.Error(), "this pull request is a draft") {
		t.Fatalf("Given a draft that skipped its gates, When judged, Then it refuses; got %v", err)
	}
	if !strings.Contains(err.Error(), "Mark it ready for review") {
		t.Fatalf("Then it says how to test it; got %v", err)
	}
}

func TestVerdictRefusesSkippedGatesOnAPullRequestThatIsNoLongerADraft(t *testing.T) {
	err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "skipped"}), false)
	if err == nil || !strings.Contains(err.Error(), "skipped build_check, check") {
		t.Fatalf("Given a ready pull request whose gates skipped, When judged, Then it refuses; got %v", err)
	}
	if !strings.Contains(err.Error(), "re-running this one replays the event") {
		t.Fatalf("Then it says a re-run is not the recovery; got %v", err)
	}
}

func TestVerdictPrefersAFailureOverTheDraftAllowance(t *testing.T) {
	err := Verdict(needs(map[string]string{"build_check": "skipped", "check": "failure"}), true)
	if err == nil || !strings.Contains(err.Error(), "failed or were cancelled: check") {
		t.Fatalf("Given a failure on a draft, When judged, Then the failure wins; got %v", err)
	}
}

func TestVerdictRefusesWhenAGateIsNotAmongTheNeeds(t *testing.T) {
	for _, missing := range gates {
		without := needs(nil)
		delete(without, missing)

		err := Verdict(without, false)
		if err == nil || !strings.Contains(err.Error(), missing) {
			t.Fatalf("Given %s absent from needs, When judged, Then it refuses naming it; got %v", missing, err)
		}
	}
}

func TestVerdictRefusesAnEmptyNeeds(t *testing.T) {
	err := Verdict(map[string]string{}, true)
	if err == nil || !strings.Contains(err.Error(), "reports nothing about") {
		t.Fatalf("Given nothing in needs, When judged, Then it refuses; got %v", err)
	}
}

func TestHaltFailsClosedOnUnreadableNeeds(t *testing.T) {
	for _, raw := range []string{"", "not json", "[]"} {
		lines := []string{}
		if code := Halt(raw, "false", func(l string) { lines = append(lines, l) }); code != 1 {
			t.Fatalf("Given needs %q, When halted, Then it exits 1; got %d", raw, code)
		}
		if len(lines) == 0 || !strings.Contains(lines[0], "::error title=distributions::") {
			t.Fatalf("Then it annotates the failure; got %v", lines)
		}
	}
}

func TestHaltExitsOneAndAnnotatesASkippedGate(t *testing.T) {
	lines := []string{}
	code := Halt(`{"build_check":{"result":"skipped"},"check":{"result":"success"},"test":{"result":"success"}}`, "false", func(l string) { lines = append(lines, l) })
	if code != 1 {
		t.Fatalf("Given a skipped gate on a ready pull request, When halted, Then it exits 1; got %d", code)
	}
	if !strings.Contains(lines[0], "results: ") || !strings.Contains(lines[1], "::error title=distributions::") {
		t.Fatalf("Then it prints the results and the error; got %v", lines)
	}
}

func TestHaltExitsZeroOnAHealthyRun(t *testing.T) {
	lines := []string{}
	if code := Halt(`{"build_check":{"result":"success"},"check":{"result":"success"},"test":{"result":"success"}}`, "", func(l string) { lines = append(lines, l) }); code != 0 {
		t.Fatalf("Given a healthy run, When halted, Then it exits 0; got %d", code)
	}
	if lines[len(lines)-1] != "All dependent jobs succeeded" {
		t.Fatalf("Then it says so; got %v", lines)
	}
}
