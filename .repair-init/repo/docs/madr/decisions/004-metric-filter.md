# Metrics filtering

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/1755

## Context and Problem Statement

Envoy by default produces a lot of metrics. You may be interested only in subset of them to not stress your metric infrastructure.
We need to have a mechanism to filter metrics.

## Considered Options

* Filter metrics dynamically on scrape request
* Filter metrics statically on Envoy bootstrap
* Customizable bootstrap

## Decision Outcome

Chosen option:
"Filter metrics dynamically on scrape request", because it follows Kuma philosophy of being as dynamic as possible.
Additionally, as a followup we should implement "Customizable bootstrap", because it both enables static metrics and other advanced use cases.
This is a subject for another MADR that will take into account more use cases.

## Pros and Cons of the Options

### Filter metrics dynamically on scrape request

UX:

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  metrics:
    enabledBackend: prometheus-1
    backends:
    - name: prometheus-1
      type: prometheus
      conf:
        skipMTLS: false
        port: 5670
        path: /metrics
        envoy:
          filterRegex: "envoy.cluster.*"
          usedOnly: false
```

Implementation:

Envoy can receive one [filter with regex](https://www.envoyproxy.io/docs/envoy/latest/operations/admin#get--stats?filter=regex)
_Note:_ filter works on statsd names, not prometheus names, when configuring filter you need to take that into account. 
You can only provide one filter, so `envoyFilterRegex` cannot be a list.

Then, in prometheus_endpoint_generator.go we can configure route to add this query param, so the `:5670` is still `/metrics`, but it adds the filter under the hood.
Then, metrics hijacker receives a request with query param and there are two options:
* It passes it only to Envoy.
* It passes it to Envoy and all the apps specified in aggregate.
I think the first option is safer.

Used only is a feature of Envoy to only publish metrics [that are used](https://www.envoyproxy.io/docs/envoy/latest/operations/admin#get--stats?usedonly).
It's useful in the environments when you don't set reachable services,
or you want to restrict drastically set of published metrics and you understand that you will be missing 0 values.
However, this might be confusing when you start with Kuma and metrics, so I think it should be opt-in.

#### Advantages
* It's dynamic. You can see the result right away.
  You can adjust this quicker.
  You may want to enable all the metrics for a minute or two and then switch it back.

#### Disadvantages
* Complex filter is hard to write (you cannot have a list of filters).
* Even if you don't care about some metrics, they are still kept in Envoy memory.

### Filter metrics statically on Envoy bootstrap

UX:

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  metrics:
    enabledBackend: prometheus-1
    backends:
    - name: prometheus-1
      type: prometheus
      conf:
        skipMTLS: false
        port: 5670
        path: /metrics
        tags: # tags that can be referred in Traffic Permission when metrics are secured by mTLS 
          kuma.io/service: dataplane-metrics
        envoy:
          staticFilter:
            rejectAll: true|false
            inclusion:
            - match:
                exact: ...
                prefix: ...
                suffix: ...
                contains: ...
                regex: ...
            exclusion: [] # same ^
        aggregate:
          - name: my-service # name of the metric, required to later disable/override with pod annotations 
            path: "/metrics/prometheus"
            port: 8888
          - name: other-sidecar
            # default path is going to be used, default: /metrics
            port: 8000
```

After applying the setting, you need to restart the proxy for setting to be applied.

Implementation:

Envoy exposes a setting to [manipulate statistics on Bootstrap config](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/stats.proto#envoy-v3-api-msg-config-metrics-v3-statsconfig).

#### Advantages
* It cuts the metrics in Envoy itself reducing memory usage.
* It's easier to write inclusion or exclusion lists.

#### Disadvantages
* You can by accident introduce undefined behaviour.
  From docs: _Excluding stats may affect Envoy’s behavior in undocumented ways._
* It requires proxy restart

### Customizable bootstrap

Like second option, but more general. It would be similar to the ContainerPatch when you can manipulate bootstrap using json patch.
It gives users freedom to manipulate bootstrap the way they want including metrics.

#### Advantages
* All advantages from the second option
* Brings more functionality to users

#### Disadvantages
* All disadvantages from the second option
* More complex to use, although it is a separate feature.
