// Package extensions stands in for the real registry, so the linter's raw JSON
// check can read a Point declared the way a resource declares one.
package extensions

type Point struct {
	ResourceType   string
	SchemaPath     []string
	Discriminator  string
	ConfigProperty string
}
