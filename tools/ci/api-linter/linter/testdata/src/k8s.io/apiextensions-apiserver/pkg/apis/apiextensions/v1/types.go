// Package v1 stands in for the real apiextensions types, so the linter's raw JSON
// check can be exercised on the type it actually matches on.
package v1

type JSON struct {
	Raw []byte `json:"-"`
}
