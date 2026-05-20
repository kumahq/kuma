# Perf Test

Scripts for running Kuma control plane performance tests on an existing cluster
(typically k3d / kind for local runs, GKE / EKS for real scale).

The scripts capture [pprof](https://pkg.go.dev/net/http/pprof) profiles and
Prometheus metrics so you can compare Kuma CP behaviour across releases, configs,
or scale points.

## Contents

| Script                 | Purpose                                                           |
| ---------------------- | ----------------------------------------------------------------- |
| `deploy-wave.sh`       | Deploy a "wave" of fake-service namespaces (N × services × replicas). |
| `collect-profiles.sh`  | Grab pprof (cpu, heap, goroutine, mutex, block, trace) + metrics snapshot from one or more CP endpoints. |
| `run-waves.sh`         | Progressive load ramp: deploys waves 1..5, stabilizes, profiles after each wave. |
| `policy-scenarios.sh`  | Policy/KDS load scenarios: spread vs. single-DP MeshTimeout targeting, updates, bulk delete. |
| `dump-prometheus.sh`   | Pull a range of Kuma CP metrics from Prometheus as JSON for offline analysis. |

All scripts are driven by environment variables and take sensible defaults for
local runs.

## Prerequisites

- `kubectl` context pointing at a cluster with Kuma already deployed
- CP has `KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS=true` (enables `/debug/pprof`)
- CP diagnostics port reachable (typically via `kubectl port-forward`)
- `go` in `$PATH` (for `go tool pprof`)
- `curl`, `python3` (for `dump-prometheus.sh`)

## Quick start

```bash
# Port-forward the CP diagnostics port
kubectl -n kuma-system port-forward svc/kuma-control-plane 5680:5680 &

# Progressive wave test (50 → 2350 DPs), profiles after each wave
./run-waves.sh

# Phase 4: policy scenarios (requires at least wave1 DPs already deployed)
./policy-scenarios.sh

# Dump Prometheus metrics for the last 150 minutes
kubectl -n monitoring port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 &
./dump-prometheus.sh
```

## Multi-zone profiling

Profile zone and global CPs simultaneously by setting `ENDPOINTS`:

```bash
ENDPOINTS="zone=http://localhost:5680 global=http://34.10.179.32:5680" \
  ./run-waves.sh
```

Each collected profile is prefixed with its endpoint name (e.g. `zone-cpu.pb.gz`,
`global-heap.pb.gz`).

## Wave definitions

`run-waves.sh` defaults to this ramp (override with `WAVES="..."`):

| Wave | Namespaces | Services/ns | Replicas | Total DPs | Cumulative |
| ---- | ---------- | ----------- | -------- | --------- | ---------- |
| 1    | 5          | 5           | 2        | 50        | 50         |
| 2    | 10         | 5           | 2        | 100       | 150        |
| 3    | 20         | 5           | 3        | 300       | 450        |
| 4    | 30         | 10          | 3        | 900       | 1350       |
| 5    | 20         | 10          | 5        | 1000      | 2350       |

## Policy scenarios

`policy-scenarios.sh` runs these against an already-running DP fleet:

1. **baseline** — no user policies, just CPU + heap
2. **spread** — N `MeshTimeout` policies targeting N different Dataplane label
   selectors (load distributed)
3. **single-dp** — N policies all targeting the same selector (load concentrated)
4. **spread-update** — patch every spread policy
5. **single-dp-update** — patch every single-dp policy
6. **delete** — bulk delete

Assumes the mesh has `MeshService.Mode=Exclusive` and the CP runs with
`KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED=true` so policies target DPs via label
selectors rather than deprecated service references.

## Output layout

```
profiles/
├── wave1/                    — one dir per wave or scenario
│   ├── cpu.pb.gz
│   ├── heap.pb.gz
│   ├── goroutine.pb.gz
│   ├── mutex.pb.gz
│   ├── block.pb.gz
│   ├── trace.out
│   └── metrics.txt
├── p4-baseline/              — policy scenario output (one dir per scenario)
│   ├── zone-cpu.pb.gz        — prefix matches the ENDPOINTS name
│   ├── zone-heap.pb.gz
│   ├── zone-metrics.txt
│   ├── global-cpu.pb.gz
│   └── global-metrics.txt
└── prometheus/               — JSON range-query dumps
    ├── go_goroutines.json
    ├── xds_generation_count.json
    └── ...
```

`profiles/` is ignored by git (`.gitignore`) — results stay local.

## Analysis hints

```bash
# Compare CPU profiles across waves
go tool pprof -http=:8080 profiles/wave1/cpu.pb.gz
go tool pprof -base profiles/wave1/cpu.pb.gz profiles/wave3/cpu.pb.gz

# Heap growth
go tool pprof -diff_base profiles/wave1/heap.pb.gz profiles/wave3/heap.pb.gz

# Trace analysis
go tool trace profiles/wave3/trace.out
```
