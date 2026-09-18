export const ATTEMPTS = 3;

export function decisions({ labels, draft }) {
  return `json=${JSON.stringify(labels)}\ndraft=${draft}`;
}

export async function readPullRequest({ get, wait, attempts = ATTEMPTS }) {
  let last;

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      const pull = await get();
      if (!Array.isArray(pull?.labels) || typeof pull?.draft !== "boolean") {
        throw new Error(`pull request payload has no labels or draft: ${JSON.stringify(pull)}`);
      }
      return { labels: pull.labels.map((label) => label.name), draft: pull.draft };
    } catch (cause) {
      last = cause;
      if (attempt < attempts) {
        await wait(attempt * 5000);
      }
    }
  }

  throw last;
}

export async function hold({ post, sha, log }) {
  try {
    await post(sha);
    log(`held distributions on ${sha} until this run reports`);
    return true;
  } catch (cause) {
    log(
      `::warning title=meta::could not hold the distributions check on ${sha}, so a result from an earlier run stands until this one finishes: ${cause.message}`,
    );
    return false;
  }
}

export async function main({ event, number, sha, get, post, wait, write, log }) {
  if (event !== "pull_request") {
    const decided = decisions({ labels: [], draft: "" });
    write(decided);
    log(decided);
    return 0;
  }

  let pull;
  try {
    pull = await readPullRequest({ get, wait });
  } catch (cause) {
    log(`::error title=meta::could not read pull request ${number} after ${ATTEMPTS} attempts: ${cause.message}`);
    return 1;
  }

  const decided = decisions(pull);
  write(decided);
  log(decided);
  await hold({ post, sha, log });
  return 0;
}

if (import.meta.filename === process.argv[1]) {
  const { GH_TOKEN, GH_REPO, PR, HEAD_SHA, GITHUB_EVENT_NAME, GITHUB_OUTPUT } = process.env;
  const { appendFileSync } = await import("node:fs");
  const headers = { authorization: `Bearer ${GH_TOKEN}`, accept: "application/vnd.github+json" };

  const send = async (path, init) => {
    const response = await fetch(`https://api.github.com${path}`, { ...init, headers });
    if (!response.ok) {
      throw new Error(`${response.status} ${response.statusText} for ${path}`);
    }
    return response.json();
  };

  process.exit(
    await main({
      event: GITHUB_EVENT_NAME,
      number: PR,
      sha: HEAD_SHA,
      get: () => send(`/repos/${GH_REPO}/pulls/${PR}`),
      post: (head_sha) =>
        send(`/repos/${GH_REPO}/check-runs`, {
          method: "POST",
          body: JSON.stringify({ name: "distributions", head_sha, status: "in_progress" }),
        }),
      wait: (ms) => new Promise((resolve) => setTimeout(resolve, ms)),
      write: (decided) => appendFileSync(GITHUB_OUTPUT, `${decided}\n`),
      log: (line) => console.log(line),
    }),
  );
}
