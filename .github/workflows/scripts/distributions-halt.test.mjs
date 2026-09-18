import assert from "node:assert/strict";
import { test } from "node:test";
import { main, verdict } from "./distributions-halt.mjs";

const needs = (overrides) => ({
  release_sha_gate: { result: "success" },
  meta: { result: "success" },
  refuse_fork_publish: { result: "skipped" },
  build_check: { result: "success" },
  build_publish: { result: "skipped" },
  check: { result: "success" },
  test: { result: "success" },
  provenance: { result: "skipped" },
  ...overrides,
});

test("Given every dependency succeeded or skipped, When judged, Then it passes", () => {
  assert.equal(verdict(needs(), false).ok, true);
});

test("Given a failed dependency, When judged, Then it names that job", () => {
  const outcome = verdict(needs({ check: { result: "failure" } }), false);
  assert.equal(outcome.ok, false);
  assert.equal(outcome.error, "these jobs failed or were cancelled: check");
});

test("Given a cancelled dependency, When judged, Then it names that job too", () => {
  const outcome = verdict(needs({ test: { result: "cancelled" } }), false);
  assert.match(outcome.error, /^these jobs failed or were cancelled: test$/);
});

test("Given several broken dependencies, When judged, Then it names all of them in order", () => {
  const outcome = verdict(needs({ check: { result: "failure" }, test: { result: "cancelled" } }), false);
  assert.equal(outcome.error, "these jobs failed or were cancelled: check, test");
});

test("Given a draft that skipped its gates, When judged, Then it passes", () => {
  const outcome = verdict(needs({ build_check: { result: "skipped" }, check: { result: "skipped" } }), true);
  assert.equal(outcome.ok, true);
});

test("Given a run that skipped its gates for a pull request that is no longer a draft, When judged, Then it refuses", () => {
  const outcome = verdict(needs({ build_check: { result: "skipped" }, check: { result: "skipped" } }), false);
  assert.equal(outcome.ok, false);
  assert.match(outcome.error, /skipped build_check, check/);
  assert.match(outcome.error, /re-running this one replays the event/);
});

test("Given a failure on a draft, When judged, Then the failure wins over the draft allowance", () => {
  const outcome = verdict(needs({ build_check: { result: "skipped" }, check: { result: "failure" } }), true);
  assert.equal(outcome.ok, false);
  assert.match(outcome.error, /failed or were cancelled: check/);
});

test("Given a push, where there is no pull request and nothing is skipped, When judged, Then it passes", () => {
  assert.equal(verdict(needs({ refuse_fork_publish: { result: "skipped" } }), false).ok, true);
});

test("Given a need with no result, When judged, Then it is neither broken nor a skipped gate", () => {
  assert.equal(verdict({ meta: {}, build_check: { result: "success" } }, false).ok, true);
});

test("Given the entrypoint runs, When a gate was skipped on a ready pull request, Then it exits 1 and annotates", () => {
  const lines = [];
  const code = main(
    { NEEDS: JSON.stringify(needs({ build_check: { result: "skipped" } })), IS_DRAFT: "false" },
    (line) => lines.push(line),
  );
  assert.equal(code, 1);
  assert.match(lines[0], /^results: /);
  assert.match(lines[1], /^::error title=distributions::/);
});

test("Given the entrypoint runs on a healthy run, When judged, Then it exits 0", () => {
  const lines = [];
  const code = main({ NEEDS: JSON.stringify(needs()), IS_DRAFT: "" }, (line) => lines.push(line));
  assert.equal(code, 0);
  assert.equal(lines.at(-1), "All dependent jobs succeeded");
});
