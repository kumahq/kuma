## How to generate a new Kuma policy

Use the tool:

```shell
go run ./tools/policy-gen/bootstrap/... --name CaseNameOfPolicy --is-policy true
```

The output of the tool will tell you where the important files are!

## Add plugin name to the configuration

To enable policy you need to adjust configuration of two places:
* Remove `+kuma:policy:skip_registration=true` from your policy schema.
* `pkg/plugins/policies/core/ordered/ordered.go`. Plugins name is equals to `KumactlArg` in file `zz_generated.resource.go`. It's important to place the plugin in the correct place because the order of executions is important.

## Linter

There is a [liner](https://github.com/kumahq/kuma/blob/61f28e4bcca0b5e6aa1044a0e536abb6a6a53642/tools/ci/api-linter/linter/api-linter.go#L49-L48)
to check if the fields are correctly defined.

## How to map API to a Go struct

Field is _mergeable_ if it is directly in `default` struct or any sub-struct.

The exception is where the struct has only required properties (like [Rate](https://github.com/kumahq/kuma/blob/9f07f444e076433bfdff4ce5d26958006234dd3b/pkg/plugins/policies/meshratelimit/api/v1alpha1/meshratelimit.go#L107-L113)).
In that case the struct is mergable as a whole (needs to be annotated with `+kuma:non-mergeable-struct`) and properties are _not mergeable_.

To allow kubernetes "union types" we created a `+kuma:discriminator` annotation to mark a [discriminator field](https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/1027-api-unions/README.md?utm_source=chatgpt.com#discriminator-field) to allow "required" fields on mergeable structs.

Field is _not mergeable_ if it is inside the struct which is inside the list.

```yaml
default:
  a: value_a
  b: 
    c: value_c
    d:
      e: required
      f: required
  g:
    - h: value_e
    - i: 
        j: value_g
    - k: []
```

* Mergeable: a, b, c, d, g
* Not mergeable: e, f, h, i, j, k

**All mergeable fields are optional.**

### Mergeable

#### Optional

Need to be defined with "omitempty", have **no default kubernetes annotation** ("+kubebuilder:default") and be a pointer.

Example:
```yaml
ValidPtr  *string   `json:"valid_ptr,omitempty"`
```

#### Discriminator (Required)

```go
type Conf struct {
    Discriminator Discriminator `json:"discriminator"`
}

type Discriminator struct {
    // +kuma:discriminator
    Type string `json:"type"`

    OptionOne *OptionOne `json:"optionOne,omitempty"`
    OptionTwo *OptionTwo `json:"optionTwo,omitempty"`
}

```

#### Mergeable struct with required fields (+kuma:non-mergeable-struct)

This can be thought of as "mergable as a whole".

```go
type Conf struct {
    // +kuma:non-mergeable-struct
    NonMergeableStruct NonMergeableStruct `json:"nonMergeableStruct"`
}

type NonMergeableStruct struct {
    RequiredIntField int `json:"requiredIntField"`
    RequiredStrField string `json:"requiredStrField"`
}
```

### Not mergeable

Three types of fields are allowed:


#### User optional with default

```go
type OtherStruct struct {
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=false
    NonMergeableOptional string   `json:"nonNergeableOptional"`
}
```

#### User optional without default

```go
type OtherStruct struct {
    NonMergeablePtr      *string  `json:"nonMergeable,omitempty"`
}
```

#### User required

```go
type OtherStruct struct {
    NonMergeableRequired string   `json:"nonNergeable"`
}
```
