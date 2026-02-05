# POC: Helm Subchart for Zone Proxy Deployment

## Problem: The Cat-and-Mouse Game

The current Kuma Helm chart mirrors Kubernetes API fields into `values.yaml` one by one.
For **ingress alone** this means:

| File | Lines | What it does |
|------|------:|--------------|
| `values.yaml` (ingress section) | ~144 | Declares every supported field |
| `ingress-deployment.yaml` | 148 | Wires each value into the K8s spec |

Every time a user needs a field the chart doesn't expose (e.g. `shareProcessNamespace`,
`initContainers`, `topologySpreadConstraints`), someone must:

1. Add the field to `values.yaml`
2. Add the corresponding `{{- with }}` / `toYaml` block in the template
3. Open a PR, wait for review, wait for a release

This is the *cat-and-mouse problem* — the chart will never fully cover the Kubernetes API
surface, and users are blocked until the chart catches up.

## Solution: Subchart with Raw Spec Passthrough

A subchart accepts **raw Kubernetes spec sections** as values and merges them into a
minimal base Deployment. The template never needs to enumerate individual fields.

### Key insight

```yaml
# Instead of this (current approach — fixed fields):
ingress:
  nodeSelector: {}
  tolerations: []
  affinity: {}
  resources: {}
  # ... 140+ lines, still incomplete

# Do this (subchart approach — open-ended):
zone-proxy:
  podSpec: {}        # ANY PodSpec field
  containers: {}     # ANY container field
```

The subchart template uses Helm's `merge` to overlay user values onto sensible defaults:

```
defaultContainer  +  .Values.containers  →  final container spec
defaultPodSpec    +  .Values.podSpec      →  final pod spec
```

## Structure

```
094/
├── README.md
├── kuma/                          # Parent chart
│   ├── Chart.yaml                 # Declares zone-proxy subchart dependency
│   ├── values.yaml                # User-facing values (passed through)
│   └── templates/
│       └── _helpers.tpl
└── charts/
    └── zone-proxy/                # Subchart
        ├── Chart.yaml
        ├── values.yaml            # Accepts raw K8s spec fields
        └── templates/
            ├── deployment.yaml    # ~40 lines — merges user values into base spec
            └── service.yaml
```

## Running the POC

```bash
cd docs/madr/decisions/pocs/094/kuma

# Build subchart dependency
helm dependency update .

# Render with default values (from kuma/values.yaml)
helm template my-release .
```

### Add a new field without changing templates

```bash
# shareProcessNamespace — not in any values.yaml, just works:
helm template my-release . \
  --set 'zone-proxy.podSpec.shareProcessNamespace=true'

# terminationGracePeriodSeconds:
helm template my-release . \
  --set 'zone-proxy.podSpec.terminationGracePeriodSeconds=60'

# Container securityContext:
helm template my-release . \
  --set 'zone-proxy.containers.securityContext.readOnlyRootFilesystem=true'
```

All of these render into the correct Deployment spec positions with zero template changes.

## Comparison

### Current approach (ingress)

```
values.yaml            144 lines   Fixed set of fields
deployment template    148 lines   One {{- with }} block per field
────────────────────────────────
Total                  292 lines   Supports ~15 PodSpec/container fields
```

### Subchart approach (this POC)

```
values.yaml             27 lines   Open-ended: podSpec + containers
deployment template     44 lines   Single merge, renders entire spec
────────────────────────────────
Total                   71 lines   Supports ALL PodSpec/container fields
```

**4x fewer lines, unlimited field coverage.**

## What This Proves

1. **No chart changes needed for new fields** — `podSpec.shareProcessNamespace`,
   `podSpec.initContainers`, `containers.lifecycle`, etc. all work out of the box.
2. **Sensible defaults are preserved** — image, ports, probes, and args come from
   the template; user values are merged on top.
3. **Parent chart passthrough works** — values set under `zone-proxy:` in the
   parent chart flow through to the subchart without any extra wiring.
4. **Dramatically less template code** — 71 lines vs 292 lines for equivalent
   (actually superior) functionality.
