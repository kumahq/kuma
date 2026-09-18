const GATES = ["build_check", "check", "test"];

export function verdict(needs, isDraft) {
  const results = Object.fromEntries(
    Object.entries(needs ?? {}).map(([job, value]) => [job, value?.result ?? null]),
  );

  const broken = Object.keys(results).filter(
    (job) => results[job] === "failure" || results[job] === "cancelled",
  );
  if (broken.length > 0) {
    return { ok: false, results, error: `these jobs failed or were cancelled: ${broken.join(", ")}` };
  }

  const gated = GATES.filter((job) => results[job] === "skipped");
  if (!isDraft && gated.length > 0) {
    return {
      ok: false,
      results,
      error: `this run skipped ${gated.join(", ")}, but the pull request is not a draft, so these results do not describe it. Push a commit to start a run that tests it - re-running this one replays the event it was started with and skips them again.`,
    };
  }

  return { ok: true, results };
}

export function main(env, log) {
  const needs = JSON.parse(env.NEEDS);
  const outcome = verdict(needs, env.IS_DRAFT === "true");
  log(`results: ${JSON.stringify(outcome.results)}`);

  if (!outcome.ok) {
    log(`::error title=distributions::${outcome.error}`);
    return 1;
  }

  log("All dependent jobs succeeded");
  return 0;
}

if (import.meta.filename === process.argv[1]) {
  process.exit(main(process.env, (line) => console.log(line)));
}
