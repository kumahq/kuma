# Extended Global Insights

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/7707

## Context and Problem Statement

We would like to extend current Global Insight to add more meaningful metrics to it. We would like to add info about number of policies, zones,
services, dataplanes and their status.

## Considered Options

* extend `GlobalInsight` with data from `MeshInsight` and `ZoneInsight`.

## Decision Outcome

We will extend Global Insight with detailed information about number of policies, zones, service, dataplanes and its status. It will be done using 
information from `MeshInsight` and `ZoneInsight`

### Solution

Right now Global Insight is really simple, it only returns count of resources, example response:

```json
{
  "type": "GlobalInsights",
  "creationTime": "2023-09-07T07:04:48.473293841Z",
  "resources": {
    "GlobalSecret": {
      "total": 6
    },
    "Mesh": {
      "total": 1
    },
    "Zone": {
      "total": 0
    },
    "ZoneEgress": {
      "total": 0
    },
    "ZoneIngress": {
      "total": 0
    }
  }
}
```

We would like to add more detailed information about zones, services and dataplanes. New extended Global Insight 
should look like this:

```json
{
  "resources": ..., // <== this stays the same
  "services": {
    "total": 5,
    "internal": 4,
    "external": 1,
    "internal_by_status": {
      "Online": 2,
      "Offline": 1,
      "pariallyDegrated": 1
     }
  },
  "zones": {
    "cps": {
      "online": 1,
      "total": 1
    }
  },
  "dataplanes": {
    "online": 23,
    "offline": 10,
    "partially_degraded": 17,
    "total": 50
  },
  "policies": {
    "total": 100
  }
}
```

#### How to get data

Most of this data is already computed. To assemble this, we need:
- `services` are mostly present in `MeshInsight`, missing part is `internal_by_status` data. We need to extend `MeshInsight.Services`
   with this information. Moreover, we need to aggregate this to get the overall count.
- `dataplanes` are present in `MeshInsight.Dataplanes` we only need to aggregate them to get the overall count.
- `zones` info can be extracted and aggregated from `ZoneInsight`
- `policies` are present in `MeshInsight.Policies` we just need to aggregate them to get the overall count.