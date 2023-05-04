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
~~* don't retry on canary hosts (the host should be preliminarily marked with `canary: true`)~~ (doesn't make sense for us to implement)
* don't retry on hosts with specified metadata key/value

And with `retry_priority` we get:

* don't retry on previous priorities

These predicates can be combined into a single `hostSelection` section:

```yaml
hostSelection:
  - predicate: OMIT_PREVIOUS_HOSTS
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
    // HostSelection is a list of predicates that dictate how hosts should be selected
    // when requests are retried.
    HostSelection *[]Predicate `json:"hostSelection,omitempty"`
    // HostSelectionMaxAttempts is the maximum number of times host selection will be
    // reattempted before giving up, at which point the host that was last selected will
    // be routed to. If unspecified, this will default to retrying once.
    HostSelectionMaxAttempts *int64 `json:"hostSelectionMaxAttempts,omitempty"`
}

type PredicateType string

var (
    OmitPreviousHosts      PredicateType = "OmitPreviousHosts"
    OmitHostsWithTags      PredicateType = "OmitHostsWithTags"
    OmitPreviousPriorities PredicateType = "OmitPreviousPriorities"
)

type Predicate struct {
    // Type is requested predicate mode. Available values are OmitPreviousHosts, OmitHostsWithTags,
    // and OmitPreviousPriorities.
    PredicateType PredicateType `json:"predicate"`
    // Tags is a map of metadata to match against for selecting the omitted hosts. Required if Type is
    // OmitHostsWithTags
    Tags map[string]string `json:"tags,omitempty"`
    // UpdateFrequency is how often the priority load should be updated based on previously attempted priorities.
    // Used for OmitPreviousPriorities. Default is 2 if not set.
    UpdateFrequency int32 `json:"updateFrequency,omitempty"`
}
```

## Considered Options

* Implement these 3 fields / options.

## Decision Outcome

Implement these 3 fields / options.