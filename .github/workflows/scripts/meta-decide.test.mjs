import assert from "node:assert/strict";
import { test } from "node:test";
import { decisions, hold, main, readPullRequest } from "./meta-decide.mjs";

const pull = (labels, draft = false) => ({ labels: labels.map((name) => ({ name })), draft });
const nowait = async () => {};

test("decisions writes the labels as JSON and the draft flag beside them", () => {
  assert.equal(decisions({ labels: ["ci/skip-test"], draft: true }), 'json=["ci/skip-test"]\ndraft=true');
  assert.equal(decisions({ labels: [], draft: false }), "json=[]\ndraft=false");
});

test("Given a label whose name needs escaping, When written, Then the JSON survives a round trip", () => {
  const written = decisions({ labels: ['a"b', "c\nd"], draft: false });
  assert.deepEqual(JSON.parse(written.split("\n")[0].slice("json=".length)), ['a"b', "c\nd"]);
});

test("Given the read succeeds first time, When reading, Then it does not retry", async () => {
  let calls = 0;
  const result = await readPullRequest({
    get: async () => {
      calls += 1;
      return pull(["ci/run-build"], true);
    },
    wait: nowait,
  });
  assert.equal(calls, 1);
  assert.deepEqual(result, { labels: ["ci/run-build"], draft: true });
});

test("Given two failures then a success, When reading, Then it returns the third answer", async () => {
  let calls = 0;
  const result = await readPullRequest({
    get: async () => {
      calls += 1;
      if (calls < 3) throw new Error("502");
      return pull([], false);
    },
    wait: nowait,
  });
  assert.equal(calls, 3);
  assert.deepEqual(result, { labels: [], draft: false });
});

test("Given every attempt fails, When reading, Then it throws the last cause", async () => {
  await assert.rejects(
    readPullRequest({
      get: async () => {
        throw new Error("503");
      },
      wait: nowait,
    }),
    /503/,
  );
});

test("Given a body with no labels or no draft flag, When reading, Then it retries rather than trusting it", async () => {
  for (const body of [{}, { labels: [] }, { draft: false }, { labels: "x", draft: false }, null]) {
    let calls = 0;
    await assert.rejects(
      readPullRequest({
        get: async () => {
          calls += 1;
          return body;
        },
        wait: nowait,
      }),
      /payload has no labels or draft/,
    );
    assert.equal(calls, 3);
  }
});

test("Given the read backs off, When retrying, Then it waits longer each time", async () => {
  const waits = [];
  await assert.rejects(
    readPullRequest({
      get: async () => {
        throw new Error("502");
      },
      wait: async (ms) => waits.push(ms),
    }),
  );
  assert.deepEqual(waits, [5000, 10000]);
});

test("Given the hold posts, When held, Then it reports the commit it held", async () => {
  const lines = [];
  assert.equal(await hold({ post: async () => ({ id: 1 }), sha: "abc", log: (l) => lines.push(l) }), true);
  assert.match(lines[0], /^held distributions on abc/);
});

test("Given the hold is refused, as on a fork, When held, Then it warns and carries on", async () => {
  const lines = [];
  const held = await hold({
    post: async () => {
      throw new Error("403 Forbidden");
    },
    sha: "abc",
    log: (l) => lines.push(l),
  });
  assert.equal(held, false);
  assert.match(lines[0], /^::warning title=meta::could not hold/);
  assert.match(lines[0], /403 Forbidden/);
});

const run = (over = {}) => {
  const state = { written: [], logs: [], posted: [] };
  const deps = {
    event: "pull_request",
    number: "7",
    sha: "abc",
    get: async () => pull(["ci/skip-test"], true),
    post: async (sha) => state.posted.push(sha),
    wait: nowait,
    write: (d) => state.written.push(d),
    log: (l) => state.logs.push(l),
    ...over,
  };
  return { state, code: main(deps) };
};

test("Given a pull request, When decided, Then it writes the outputs and holds the check", async () => {
  const { state, code } = run();
  assert.equal(await code, 0);
  assert.deepEqual(state.written, ['json=["ci/skip-test"]\ndraft=true']);
  assert.deepEqual(state.posted, ["abc"]);
});

test("Given a push, When decided, Then it writes empty decisions and holds nothing", async () => {
  const { state, code } = run({ event: "push" });
  assert.equal(await code, 0);
  assert.deepEqual(state.written, ["json=[]\ndraft="]);
  assert.deepEqual(state.posted, []);
});

test("Given the read never succeeds, When decided, Then it exits 1 and writes no outputs", async () => {
  const { state, code } = run({
    get: async () => {
      throw new Error("503");
    },
  });
  assert.equal(await code, 1);
  assert.deepEqual(state.written, []);
  assert.match(state.logs[0], /^::error title=meta::could not read pull request 7 after 3 attempts/);
});

test("Given the hold fails, When decided, Then the decisions still stand and the job survives", async () => {
  const { state, code } = run({
    post: async () => {
      throw new Error("403");
    },
  });
  assert.equal(await code, 0);
  assert.equal(state.written.length, 1);
  assert.match(state.logs.at(-1), /^::warning title=meta::/);
});
