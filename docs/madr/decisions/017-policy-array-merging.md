# Policy array merging

* Status: accepted

## Context and Problem Statement

[New policies](005-policy-matching.md) support merging of multiple configuration.
While merging regular fields is pretty straightforward, it is quite difficult to pick a strategy for merging list.
Our current strategy is to replace list from the most specific policy.

For example
```
type: Policy
targetRef:
  kind: Mesh
defaults:
  elements:
  - a
```
and
```
type: Policy
targetRef:
  kind: MeshService
  name: backend
defaults:
  elements:
  - b
```
would result in
```
defaults:
  elements:
  - b
```

But for some cases we would like to append list and end up with
```
defaults:
  elements:
  - a
  - b
```

Example use cases are:
* Composing ProxyTemplate modifications.
* Composing policies in Kong Mesh's OPA Policy
* Potentially composing backends in stuff like MeshTrafficLog

## Considered Options

* Name convention
* Field annotation
* Merge keyword

## Pros and Cons of the Options

### Name convention

We could have a name convention that indicates that given array is appendable not replaceable when we merge.

```
defaults:
  appendElements:
  - a
```

So every time there is the field with value array that starts with `append` we append.

Other potential prefixes:
* composed
* merged
* combined
* joined

To implement this, we would need to replace [jsonpatch](https://github.com/kumahq/kuma/blob/master/pkg/core/xds/rules.go#L207) with our reflection-based solution.

We could also express both strategies
```yaml
default:
  elements: []
  appendElements: []
```

#### Advantages
* Explicit for users
* Replacing jsonpatch will be more efficient, because we don't need to marshal to json just for the sake of merging
* Expressing both strategies

#### Disadvantages
* Longer field names
* Reserving the prefix as a field name (although it's very unlikely to see this as a prefix of any field)
* We need to roll out custom implementation of merging and maintain it.

### Field annotation

When building the policy model in Go, we could annotate the array
```
type Policy {
  Elements []string `json:"elements", kumapolicy:"appendable"`
}
```
or use `omitempty` for that. If it's present, the array is appendable.
```
type Policy {
  Elements []string `json:"elements,omitempty"`
}
```

The rest is the same with previous solution, we need to utilize this annotation when we travel the fields with reflection.

#### Advantages
* Cleaner API names.
* Replacing jsonpatch will be more efficient.

#### Disadvantages
* Users need to read the policy spec and remember it to understand that the elements are appended, not replaced.
  This might be surprising for them.
* We need to roll out custom implementation of merging.

### Merge keyword

We could introduce additional keyword, so instead of `default` we use `merge` which switches the merging algo to append lists

```
merge:
  elements:
  - a
```

#### Advantages
* Explicit and easy to understand

#### Disadvantages
* We cannot control merging on a single list. It's all append or all replace for the whole policy.
* Additional keyword, which may confuse service owners. Which one they should use by default? `defaults` or `merge`?

## Decision Outcome

Chosen option: "Name convention", because it follows the _Principle Of The Least Surprise_.
Explicit prefix has a higher chance of avoiding bug when trying to compose policies.
