# Error Wrapping

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/2604

## Context and Problem Statement

There are some inappropriate golang error usages in Kuma due to early golang standard library deficiency, 
such as golang error assertion by relying on string prefixes matching, and error target checking by using `reflect` package.
These error handling solutions are improper, and it is recommended to use [golang's error wrapping](https://go.dev/blog/go1.13-errors) capabilities instead.

For example, these are patterns that should be avoided:

```text
func IsResourceNotFound(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "resource not found")
}

func IsResourceAlreadyExists(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "resource already exists")
}

func (a *PreconditionError) Is(err error) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(err)
}
```

This document proposes replacing the above error handling solutions with golang native error wrapping API.

## Design

Extracted from the issue, we have the following goals:
* Each error should be a distinct type.
* Every `ErrorResourceFoobar()` should have a matching `IsResourceFoobarr()`.
* The `IsResource***` APIs should be implemented in terms of `errors.Is`.

### Custom Error Type

Define a struct that implements the `error` interface by providing an `Error()` method. Create a function `Is***` uses `errors.Is` to check 
if the error is of the pointer type. 

```text
type ResourceAlreadyExistsError struct {
	msg string
}

func (r ResourceAlreadyExistsError) Error() string {
	return r.msg
}

func ErrorResourceAlreadyExists(rt, name, mesh string) error {
	return &ResourceAlreadyExistsError{msg: fmt.Sprintf("resource already exists: type=%q name=%q mesh=%q", rt, name, mesh)}
}

func IsResourceAlreadyExists(err error) bool {
	return errors.Is(err, &ResourceAlreadyExistsError{})
}
```

#### Pros

* Typed-Based Error Handling: Since it uses a custom struct, you can check for the specific error type using `errors.Is`, which is robust for identifying this exact error.
* Extensibility: The struct can be extended to include additional fields (like `ResourceType`, `Name`, `Mesh`) for more context, allowing downstream code to extract structured data from the error.

#### Cons

* Verbosity: Requires defining a struct and implementing the error interface, which is more code compared to simpler approaches.
* Limited Wrapping Support: This implementation doesn't use `errors.Wrap` or `fmt.Errorf` with `%w`, so it doesn't natively support error wrapping for adding context or stack traces.
* Inconsistent Logging: The error message would be ambiguous unless you add the prefix like `resource already exists` or `resource conflict` in the head of the message.

### Sentinel Error with Wrapping

Define a singleton error using `errors.New` which serves as a unique identifier(sentinel) for the error condition, 
and use `fmt.Errorf` with the `%w` verb that includes additional context. Create a function `Is***` uses `errors.Is` to check
if the error is in the wrapped error chain.

```text
var ErrIsConflict = errors.New("conflict")

func ErrorResourceConflict(rt, name, mesh string) error {
	return fmt.Errorf("resource %w: type=%q name=%q mesh=%q", ErrIsConflict, rt, name, mesh)
}

func IsResourceConflict(err error) bool {
	return errors.Is(err, ErrIsConflict)
}
```

#### Pros

* Simplicity: Uses a lightweight sentinel error, requiring less code than a custom struct.
* Wrapping Support: Uses `fmt.Errorf` with `%w`, enabling error wrapping. This allows the error to be wrapped multiple times while still being identifiable via `errors.Is`.
* Consistent Logging: As singleton error defined, the brief error message would be put in the head for every new error wrapping like `fmt.Errorf("%w: ***")`.

#### Cons

* Less Extensible: Since it’s just a sentinel error, it doesn’t carry structured data (like `ResourceType`, `Name`, `Mesh`). Extracting them requires parsing the error message, which is fragile.

## Implications for Kong Mesh

None

## Decision

* Option 2 - Sentinel Error with Wrapping. We will define multiple singleton error by using `errors.New`, 
and create new errors using `fmt.Errorf` with the `%w` verb that includes additional context.

### Convention

1. return error with the defined function.
    ```text
    return ErrorResourceConflict(core_mesh.DataplaneType, objName, model.DefaultMesh)
    ```

2. return error with customized wrapping.
    ```text
    return fmt.Errorf("%w: %v", ErrIsConflict, err)
   ```

3. error assertion by the existing function.
   ```text
   if IsResourceConflict(err){
      ...
   }
   ```
