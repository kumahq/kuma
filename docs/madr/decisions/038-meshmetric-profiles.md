# MeshMetric profiles (predefined subsets of all metrics)

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/8845

## Context and Problem Statement

There is a lot of default Envoy metrics, even with `usedOnly` enabled.
Users might be overwhelmed by the number of metrics and not know what it's important what is less important.
Hosted providers usually charge based on the number of metrics ingested.
When using trial versions of hosted metrics users run out of free credits pretty fast.

To make it more digestible for the users and cheaper we should introduce profiles that only contain subsets of all Envoy metrics.

## Decision Drivers <!-- optional -->

* {driver 1, e.g., a force, facing concern, …}
* {driver 2, e.g., a force, facing concern, …}
* … <!-- numbers of drivers can vary -->

## Considered Options

* Base profiles on expert knowledge and current grafana dashboards
* Base profiles on stats popularity

## Decision Outcome

Chosen option: "{option 1}", because {justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force {force} | … | comes out best (see below)}.

### Positive Consequences <!-- optional -->

* {e.g., improvement of quality attribute satisfaction, follow-up decisions required, …}
* …

### Negative Consequences <!-- optional -->

* {e.g., compromising quality attribute, follow-up decisions required, …}
* …

## Pros and Cons of the Options

### Base profiles on expert knowledge and current grafana dashboards

We built the dashboards to showcase what is important to look at, we could 

```bash
cat kuma-dataplane.json | jq 'try .panels[] | try .targets[] | try .expr' | grep -E -o '\benvoy_[a-zA-Z0-9_]+\b' | sort | uniq
```


### {option 2}

{example | description | pointer to more information | …} <!-- optional -->

* Good, because {argument a}
* Good, because {argument b}
* Bad, because {argument c}
* … <!-- numbers of pros and cons can vary -->

### {option 3}

{example | description | pointer to more information | …} <!-- optional -->

* Good, because {argument a}
* Good, because {argument b}
* Bad, because {argument c}
* … <!-- numbers of pros and cons can vary -->

## Links <!-- optional -->

* {Link type} {Link to ADR} <!-- example: Refined by [ADR-0005](0005-example.md) -->
* … <!-- numbers of links can vary -->
