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

## How to map API to a Go struct

Field is _mergeable_ if it is directly in `default` struct or any sub-struct.
Field is _not mergeable_ if it is inside the struct which is inside the list.

```yaml
default:
  a: value_a
  b: 
    c: value_c
  d:
    - e: value_e
    - f: 
        g: value_g
    - h: []
```

* Mergeable: a, b, c, d
* Not mergeable: e, f, g, h

**All mergeable fields are optional.**

### List

#### Required

```go
type SomeStruct struct {
	MyListField []ItemType `json:"myListField"`
}
```

#### Optional

```go
type SomeStruct struct {
	MyListField *[]ItemType `json:"myListField,omitempty"`
}
```

### Struct

#### Required

```go
type Conf struct {
	MyStructField StructType `json:"myStructField"`
}
```

#### Optional

```go
type Conf struct {
	MyStructField *StructType `json:"myStructField,omitempty"`
}
```

### Basic types (string, int, bool, etc.)

#### Required

```go
type SomeStruct struct {
	MyField ItemType `json"myField"`
}
```

#### Optional

```go
type SomeStruct struct {
	MyField *ItemType `json"myField,omitempty"`
}
```
