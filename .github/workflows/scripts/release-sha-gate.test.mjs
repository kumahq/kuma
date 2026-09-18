import assert from "node:assert/strict";
import { test } from "node:test";
import { candidates, gate, missingJobs } from "./release-sha-gate.mjs";

const green = [
  { name: "test / test_unit", conclusion: "success" },
  { name: "test / e2e (default, v1.35, amd64)", conclusion: "success" },
  { name: "test / e2e (gatewayapi, v1.35, amd64)", conclusion: "success" },
  { name: "build_publish / digest-images", conclusion: "success" },
  { name: "build_publish / build-binaries (amd64)", conclusion: "success" },
  { name: "build_publish / publish-binaries", conclusion: "success" },
  { name: "build_publish / build-images (amd64)", conclusion: "success" },
  { name: "build_publish / publish-helm", conclusion: "success" },
];

const without = (name) => green.filter((job) => job.name !== name);
const replace = (name, conclusion) => green.map((job) => (job.name === name ? { ...job, conclusion } : job));

test("a run with every required job green is missing nothing", () => {
  assert.deepEqual(missingJobs(green), []);
});

test("a renamed job is reported by the label it was required under", () => {
  const jobs = green.map((job) => (job.name === "test / test_unit" ? { ...job, name: "test / unit_test" } : job));
  assert.deepEqual(missingJobs(jobs), ["test / test_unit"]);
});

test("one failed, cancelled or null leg fails the whole family", () => {
  for (const conclusion of ["failure", "cancelled", null]) {
    assert.deepEqual(missingJobs(replace("test / e2e (gatewayapi, v1.35, amd64)", conclusion)), ["test / e2e ..."]);
  }
});

test("a family that skipped beside a green sibling fails", () => {
  assert.deepEqual(missingJobs(replace("test / e2e (gatewayapi, v1.35, amd64)", "skipped")), ["test / e2e ..."]);
});

test("a family that produced no jobs at all fails", () => {
  const jobs = green.filter((job) => !job.name.startsWith("test / e2e"));
  assert.deepEqual(missingJobs(jobs), ["test / e2e ..."]);
});

test("an empty job list is missing every requirement", () => {
  assert.equal(missingJobs([]).length, 7);
});

test("a tag run, which runs neither the unit tests nor e2e, is missing both", () => {
  assert.deepEqual(missingJobs(without("test / test_unit").filter((job) => !job.name.startsWith("test / e2e"))), [
    "test / test_unit",
    "test / e2e ...",
  ]);
});

test("candidates keeps only successful runs, newest first", () => {
  const runs = [
    { id: 1, conclusion: "success", created_at: "2026-01-01T00:00:00Z" },
    { id: 2, conclusion: "failure", created_at: "2026-01-03T00:00:00Z" },
    { id: 3, conclusion: "success", created_at: "2026-01-02T00:00:00Z" },
  ];
  assert.deepEqual(candidates(runs).map((run) => run.id), [3, 1]);
});

const run = (id, created_at) => ({ id, created_at, conclusion: "success", html_url: `https://x/${id}` });

test("Given no successful run, When gated, Then it says none was found", async () => {
  const result = await gate({ sha: "abc", workflow: "w", listRuns: async () => [], listJobs: async () => [], log: () => {} });
  assert.equal(result.ok, false);
  assert.match(result.error, /No successful completed push run/);
});

test("Given a newer tag run and an older green branch run, When gated, Then it passes on the older one", async () => {
  const logs = [];
  const result = await gate({
    sha: "abc",
    workflow: "w",
    listRuns: async () => [run(100, "2026-01-01T00:00:00Z"), run(200, "2026-01-02T00:00:00Z")],
    listJobs: async (id) => (id === 200 ? [] : green),
    log: (line) => logs.push(line),
  });
  assert.equal(result.ok, true);
  assert.equal(result.run.id, 100);
  assert.match(logs[0], /Run https:\/\/x\/200 is missing/);
});

test("Given every job query fails, When gated, Then it reports an API failure rather than a missing run", async () => {
  const result = await gate({
    sha: "abc",
    workflow: "w",
    listRuns: async () => [run(100, "2026-01-01T00:00:00Z"), run(200, "2026-01-02T00:00:00Z")],
    listJobs: async () => {
      throw new Error("502");
    },
    log: () => {},
  });
  assert.equal(result.ok, false);
  assert.match(result.error, /GitHub API failure/);
});

test("Given one unreadable run and one genuinely incomplete run, When gated, Then it still reports the API failure", async () => {
  const result = await gate({
    sha: "abc",
    workflow: "w",
    listRuns: async () => [run(200, "2026-01-02T00:00:00Z"), run(100, "2026-01-01T00:00:00Z")],
    listJobs: async (id) => {
      if (id === 200) throw new Error("502");
      return [];
    },
    log: () => {},
  });
  assert.equal(result.ok, false);
  assert.match(result.error, /could not read the jobs of 1 run\(s\)/);
});

test("Given every run is readable and incomplete, When gated, Then it says no run is green", async () => {
  const result = await gate({
    sha: "abc",
    workflow: "w",
    listRuns: async () => [run(100, "2026-01-01T00:00:00Z")],
    listJobs: async () => [],
    log: () => {},
  });
  assert.equal(result.ok, false);
  assert.match(result.error, /has every required job green/);
});
