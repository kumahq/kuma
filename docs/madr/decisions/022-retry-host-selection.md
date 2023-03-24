# Retry Host Selection

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5897

## Context and Problem Statement

We want to add some additional retry fields plumbed through from Envoy to our Retry policy based on user demand. This is an HTTP only policy (no GRPC).

```
retry_priority
retry_host_predicate
host_selection_retry_max_attempts
```

### Design 

Envoy has "host predicates" and "priority predicates" to enforce an additional logic 
when selecting a host for the attempt. With `retry_host_predicate` we get:

* don't retry on previously retried hosts
* don't retry on canary hosts (the host should be preliminarily marked with `canary: true`)
* don't retry on hosts with specified metadata key/value

And with `retry_priority` we get:

* don't retry on previous priorities

These predicates can be combined into a single `hostSelection` section:

```yaml
hostSelection:
  - predicate: OMIT_PREVIOUS_HOSTS
  - predicate: OMIT_CANARY_HOSTS
  - predicate: OMIT_HOSTS_WITH_TAGS
    tags:
      env: dev
  - predicate: OMIT_PREVIOUS_PRIORITIES
    updateFrequency: 2                                                                                  
```

Having configurable `hostSelection` implies exposing the `host_selection_retry_max_attempts` parameter as well.
It allows configuring the maximum number of times host selection will be reattempted before giving up.

### Implementation

```go
type HTTP struct {
	// ...
  HostSelection *[]Predicate `json:"hostSelection,omitempty"`
  // ...
}

type Predicate struct {
	// Type is requested predicate mode. Available values are OMIT_PREVIOUS_HOSTS, OMIT_CANARY_HOSTS, 
  // OMIT_HOSTS_WITH_TAGS, and OMIT_PREVIOUS_PRIORITIES.
	Type *string `json:"predicate,omitempty"`
	// Tags is a map of metadata to match against for selecting the omitted hosts. Required if Type is 
  // OMIT_HOSTS_WITH_TAGS
	Tags map[string]string `json:"tags,omitempty"`
  // UpdateFrequency is how often the priority load should be updated based on previously attempted priorities. 
  // Required if Type is OMIT_PREVIOUS_PRIORITIES. 
  UpdateFrequency *uint32 `json:"tags,omitempty"`
}

```

## Considered Options

* Implement these 3 fields / options

## Decision Outcome

Implement these 3 fields / options.