# Include Envoy resource name and transparent proxy configuration in `_layout` endpoint 

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/13847

## Context and Problem Statement

While working on naming resources based on KRI ([MADR](https://github.com/kumahq/kuma/pull/13485/files)), 
we have discovered that we are missing certain information in Inspect API,
precisely in `_layout` endpoint. Which was designed in [Inspect API redesign MADR](https://docs.google.com/document/d/1EzZpk3wwneIxQNPK7WXJqhW3Y3CS1AWMXIcXix9LEoc/edit?tab=t.g7ooo2ntj4jj)

Right now schema for `_layout` endpoint look like this:

```yaml
kri: <dataplane_resource_identifier>
name: <dataplane_name>
labels: <dataplane_labels>
inbounds:
  - kri: <resource_identifier> # Inbound KRI
    port: <port_number> # port number
    protocol: http # port protocol 
outbounds:
  - kri: <resource_identifier> # MS/MMZS/MES resource identifier
    port: <port_number> # MS/MMZS/MES exposed port number
    protocol: TCP
  - kri: <resource_identifier> # MS/MMZS/MES resource identifier
    port: <port_number> # MS/MMZS/MES exposed port number
    protocol: HTTP
```

While trying to name Envoy resources using KRI, we've discovered that this will drastically increase the cardinality of metrics.
We want to generate a simplified name for inbound that will look like: `<common_prefix>_sectionName`, with this new name
it will be hard for GUI or other API users to correlate information from `_layout` endpoint with Envoy resources or stats.

### Passthrough inbound/outbound

While using transparent proxy, we add passthrough inbound and outbound which does not fit our current inbound/outbound schema
since it does not have KRI it is not a real resource and cannot be targeted in policies. We need to include these extra inbound/outbound
in our current `_layout` endpoint.

## Design

### Generated inbound name

To address the issue with simplified naming on inbounds for Envoy resources and stats, we want to add this name to existing inbounds in api.
There is question on how to name this field, here are the ideas:

1. xdsName
2. statsName
3. configurationName
4. generatedName
5. resourceName
6. configResourceName
7. proxyResourceName

Pros and cons on these names:

1. This name exactly describes what this field contains and how it is used, but it leaks Envoy internals in our API (maybe this does not matter?)
2. This is a good descriptive name as it points to where it can be used, but this omits the aspect of envoy config
3. This name is too generic, like what configuration we are talking about or how to use it? 
4. This is also too generic and does not say much
5. This could work, but most of the time we use resource in the context of Kuma resources Dataplane/MeshService etc. 
6. This precisely points to what this name, although it might not be obvious at first what it references
7. This points exactly to what this name is used for, and it does not leak Envoy internals

#### Decision

We decided to use `proxyResourceName` as this is descriptive enough and does not leak Envoy internals. Moreover, after 
discussion with the GUI team, we've decided that this field will be added not only to inbounds but also to outbounds to
make it consistent and easier to work with. For now, this field in outbound will just have outbound kri

Inbounds and Outbounds will now look like this:

```yaml
inbounds:
  - kri: <resource_identifier> # Inbound KRI
    port: <port_number> # port number
    protocol: http # port protocol
    proxyResourceName: <common_prefix>_sectionName
outbounds:
  - kri: <resource_identifier> # MS/MMZS/MES resource identifier
    port: <port_number> # MS/MMZS/MES exposed port number
    protocol: TCP
    proxyResourceName: <outbound_kri> # MS/MMZS/MES resource identifier
```

### Passthrough inbound/outbound

Since passthrough inbound/outbound are specific, and don't have KRI, they do not fit in our current schema. We should extend
our current api with extra field `transparentproxy` that will contain this information:

```yaml
kri: <dataplane_resource_identifier>
name: <dataplane_name>
labels: <dataplane_labels>
inbounds: []
outbounds: []
transparentproxy:
  inbound:
    port: <port_number>
    proxyResourceName: <generated_name> # generated passthrough inbound name that will be used to match envoy resources and stats
  outbound: 
    port: <port_number>
    proxyResourceName: <generated_name> # generated passthrough outbound name that will be used to match envoy resources and stats
```

With this addition, we will be able to show in GUI these inbound/outbound and correlate it with Envoy resources and stats.

#### Alternative solution

We could try fitting this special case inbound/outbound into our existing API, but will make API worse with lots of optional fields
since these have in common only port with actual inbounds/outbounds.

## Decision outcome

We will extend `_layout` endpoint with generated proxyResourceName field in inbounds anb outbounds for easier correlation with 
Envoy resources names and stats, and whole new section with transparent proxy configuration.

New full schema will look like:

```yaml
kri: <dataplane_resource_identifier>
name: <dataplane_name>
labels: <dataplane_labels>
inbounds:
  - kri: <resource_identifier> # Inbound KRI
    port: <port_number> # port number
    protocol: http # port protocol
    proxyResourceName: <common_prefix>_sectionName
outbounds:
  - kri: <resource_identifier> # MS/MMZS/MES resource identifier
    port: <port_number> # MS/MMZS/MES exposed port number
    protocol: TCP
    proxyResourceName: <outbound_kri> # MS/MMZS/MES resource identifier
transparentproxy:
  inbound:
    port: <port_number>
    proxyResourceName: <generated_name> # generated passthrough inbound name that will be used to match envoy resources and stats
  outbound: 
    port: <port_number>
    proxyResourceName: <generated_name> # generated passthrough outbound name that will be used to match envoy resources and stats
```
