# Extended Global Insights

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/7707

## Context and Problem Statement

We would like to extend current Global Insight to add more meaningful metrics to it. We would like to add info about number of policies, zones,
services, dataplanes and their status.

## Considered Options

* Extend `GlobalInsight` with data from `MeshInsight` and `ZoneInsight`.

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
  "services": {
    "external": {
      "total": 1
    },
    "internal": {
      "online": 2,
      "offline": 1,
      "partiallyDegraded": 1,
      "total": 4
    },
    "gatewayBuiltin": {
      "online": 2,
      "offline": 1,
      "partiallyDegraded": 1,
      "total": 4
    },
    "gatewayDelegated": {
      "online": 2,
      "offline": 1,
      "partiallyDegraded": 1,
      "total": 4
    }
  },
  "zones": {
    "controlPlanes": {
      "online": 1,
      "total": 1
    },
    "zoneEgresses": {
      "online": 1,
      "total": 1
    },
    "zoneIngresses": {
      "online": 1,
      "total": 1
    }
  },
  "dataplanes": {
    "standard": {
      "online": 23,
      "offline": 10,
      "partiallyDegraded": 17,
      "total": 50
    },
    "gatewayBuiltin": {
      "online": 23,
      "offline": 10,
      "partiallyDegraded": 17,
      "total": 50
    },
    "gatewayDelegated": {
      "online": 23,
      "offline": 10,
      "partiallyDegraded": 17,
      "total": 50
    }
  },
  "policies": {
    "total": 100
  },
  "meshes": {
    "total": 3
  }
}
```

It is worth point out that we are removing old response entirely since most of this information is present in new response,
and old response was schema less. With introduction of OpenAPI schema we want to use explicit types and if any information
from old response will be needed we can easily extend new schema and add them.

#### How to get data

Most of this data is already computed. To assemble this, we need:
- `services` We can easily extract this information from `ServiceInsight`, we only need to aggregate them to get the overall count. 
- `dataplanes` are present in `MeshInsight.Dataplanes`, we only need to aggregate them to get the overall count.
- `zones` info can be extracted and aggregated from `ZoneInsight`
- `policies` are present in `MeshInsight.Policies` we just need to aggregate them to get the overall count.
- `meshes` can be easily extracted from `MeshInsight`

In order to get data needed for `dataplanes` object we would have to extend `MeshInsight.DataplanesByType` object. 
To make this change backward compatible we would have to add two fields `gatewayBuiltin` and `gatewayDelegated`. MeshInsight
dataplanesByType would look like this: 

```json
{
  "dataplanesByType": {
    "standard": {
      "total": 2,
      "online": 2
    },
    "gateway": {
      "total": 1,
      "online": 1
    },
    "gatewayBuiltin": {
      "total": 0,
      "online": 0
    },
    "gatewayDelegated": {
      "total": 1,
      "online": 1
    }
  }
}
```

#### Computing Global Insight

Computing Global Insight is not heavy, since we have all the data computed already, so we will create it when asked for. 
We will add some small cache for it so don't recompute it too often.

#### Generating openapi schema 

We would like to create new endpoint and generate Go types from OpenAPI schema. OpenAPI recommends to write schema first 
and then generate language specific code from it. [Source](https://learn.openapis.org/best-practices.html)

We will be adding new endpoint named `/global-insight`

#### Extending API for custom Kuma distributions

OpenAPI schema can be easily extended in custom distributions. Custom distribution specs can link end extend Kuma specs.
Types generated in custom distros can reuse types built in Kuma using [import mapping](https://github.com/deepmap/oapi-codegen#import-mappings).
Then we can use `APIWebServiceCustomize` from Runtime so we can remove Kuma implementation of Global Insight service and replace it with
custom one.
