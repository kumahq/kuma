# Internal listeners naming convention
* status: accepted

## Context and Problem statement
At the moment we don't clearly distinguish between listeners used for service to service communication
and internal listeners, like Prometheus listeners.

## Decision Outcome

We will add prefix `_` to internal listeners.

## Solution
At the moments internal listeners names start with `kuma` prefix, problem is that service can also have this prefix.
To be able to easily distinguish between service listeners and internal listeners, we can
add `_` prefix to listener name