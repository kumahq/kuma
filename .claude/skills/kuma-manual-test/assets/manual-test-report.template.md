# manual test report - **RUN_ID**

Compactness rules:

- keep this report concise and summary-oriented
- do not paste raw command output or full YAML dumps here
- reference evidence files in `artifacts/`, `state/`, and `manifests/`
- run `scripts/report-compactness-check.sh` before closing the run

## metadata

- run id: `__RUN_ID__`
- created at (utc): `__CREATED_AT__`
- operator: ``
- feature scope: ``

## environment summary

- profile: ``
- kubeconfig(s): ``
- cluster lifecycle commands: ``
- local build revision: ``
- local `kumactl` path/version: ``

## preflight status

| check                     | status    | evidence        |
| ------------------------- | --------- | --------------- |
| docker reachable          | PASS/FAIL | `artifacts/...` |
| cluster reachable         | PASS/FAIL | `artifacts/...` |
| kuma control plane ready  | PASS/FAIL | `artifacts/...` |
| sidecar injection healthy | PASS/FAIL | `artifacts/...` |
| local kumactl confirmed   | PASS/FAIL | `artifacts/...` |

## test execution log

| test id | title         | status    | manifests       | key evidence    | notes |
| ------- | ------------- | --------- | --------------- | --------------- | ----- |
| G1      | Resource CRUD | PASS/FAIL | `manifests/...` | `artifacts/...` |       |

## failures and triage

### failure n

- test id:
- expected:
- observed:
- classification: manifest issue / environment issue / product bug
- evidence:
- decision and next step:

## bug candidates

### bug n

- symptom:
- minimal repro:
- probable scope:
- probable root cause:
- files or components to inspect:

## conclusions

- overall status:
- confidence level:
- unresolved risks:
- recommended follow-up:
