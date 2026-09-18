const REQUIRED = [
  { label: "test / test_unit", pattern: /^test \/ test_unit$/ },
  { label: "test / e2e ...", pattern: /^test \/ e2e/ },
  { label: "build_publish / digest-images", pattern: /^build_publish \/ digest-images$/ },
  { label: "build_publish / build-binaries", pattern: /^build_publish \/ build-binaries/ },
  { label: "build_publish / publish-binaries", pattern: /^build_publish \/ publish-binaries$/ },
  { label: "build_publish / build-images ...", pattern: /^build_publish \/ build-images/ },
  { label: "build_publish / publish-helm", pattern: /^build_publish \/ publish-helm$/ },
];

export function candidates(runs) {
  return runs
    .filter((run) => run.conclusion === "success")
    .sort((a, b) => String(b.created_at).localeCompare(String(a.created_at)));
}

export function missingJobs(jobs) {
  return REQUIRED.filter(({ pattern }) => {
    const seen = jobs.filter((job) => pattern.test(job.name)).map((job) => job.conclusion);
    return seen.length === 0 || !seen.every((conclusion) => conclusion === "success");
  }).map(({ label }) => label);
}

export async function gate({ sha, workflow, listRuns, listJobs, log }) {
  const runs = candidates(await listRuns());

  if (runs.length === 0) {
    return {
      ok: false,
      error: `No successful completed push run of ${workflow} found for tagged SHA ${sha}. Every tag push requires a prior green branch push CI run for the tagged commit. Recovery: re-run the branch push CI on this SHA (re-tests and re-publishes the preview), then re-tag.`,
    };
  }

  let unread = 0;
  let inspected = 0;

  for (const run of runs) {
    let jobs;
    try {
      jobs = await listJobs(run.id);
    } catch (cause) {
      unread += 1;
      log(`::warning title=release_sha_gate::gate could not read the jobs of run ${run.html_url}: ${cause.message}`);
      continue;
    }

    inspected += 1;
    const missing = missingJobs(jobs);

    if (missing.length === 0) {
      return { ok: true, run };
    }

    log(`Run ${run.html_url} is missing required successful job(s): ${missing.join(", ")}`);
  }

  if (unread > 0) {
    return {
      ok: false,
      error: `gate could not read the jobs of ${unread} run(s) for SHA ${sha} and found no green run among the ${inspected} it did read. Some of this is a GitHub API failure rather than a missing run. Recovery: re-run this job.`,
    };
  }

  return {
    ok: false,
    error: `No push run of ${workflow} on SHA ${sha} has every required job green. Recovery: re-run the branch push CI on this SHA, then re-tag.`,
  };
}

async function paged(url, token) {
  const collected = [];
  let next = url;

  while (next) {
    const response = await fetch(next, {
      headers: { authorization: `Bearer ${token}`, accept: "application/vnd.github+json" },
    });
    if (!response.ok) {
      throw new Error(`${response.status} ${response.statusText} for ${next}`);
    }
    collected.push(await response.json());
    const link = response.headers.get("link") ?? "";
    next = link.match(/<([^>]+)>;\s*rel="next"/)?.[1] ?? null;
  }

  return collected;
}

if (import.meta.filename === process.argv[1]) {
  const { GH_TOKEN, REPO, SHA, WORKFLOW } = process.env;
  const api = "https://api.github.com";
  const log = (line) => console.log(line);

  const result = await gate({
    sha: SHA,
    workflow: WORKFLOW,
    listRuns: async () => {
      const pages = await paged(
        `${api}/repos/${REPO}/actions/workflows/${WORKFLOW}/runs?head_sha=${SHA}&status=completed&event=push&per_page=100`,
        GH_TOKEN,
      );
      return pages.flatMap((page) => page.workflow_runs);
    },
    listJobs: async (id) => {
      const pages = await paged(`${api}/repos/${REPO}/actions/runs/${id}/jobs?per_page=100`, GH_TOKEN);
      return pages.flatMap((page) => page.jobs);
    },
    log,
  });

  if (!result.ok) {
    log(`::error title=release_sha_gate::${result.error}`);
    process.exit(1);
  }

  log(`Trusted source run: ${result.run.html_url}`);
}
