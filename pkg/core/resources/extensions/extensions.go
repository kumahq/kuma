// Package extensions collects the configurations that plug into the freeform
// extension points of Kuma resources.
//
// An extension point is a `config` field typed as raw JSON, picked by a sibling
// discriminator such as HostnameGenerator's `spec.extension.type`. The control
// plane validates the payload by unmarshalling it into a Go struct, but nothing
// in the OpenAPI document says which structs are accepted, so `config` shows up
// as "anything at all". Registering a configuration here hands the OpenAPI
// generator the Go type to describe.
//
// Registration lives next to the extension it describes, so an extension that
// works but is undocumented is not a state you can reach by forgetting a marker.
package extensions

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// Point is an extension site on a resource.
type Point struct {
	// ResourceType is the resource that carries the extension point.
	ResourceType model.ResourceType
	// SchemaPath is the chain of property names leading to the extension object
	// inside the resource's OpenAPI item schema, for example {"spec", "extension"}.
	SchemaPath []string
	// Discriminator names the property of the extension object whose value picks
	// the configuration, for example "type" or "name".
	Discriminator string
}

func (p Point) equal(other Point) bool {
	if p.ResourceType != other.ResourceType || p.Discriminator != other.Discriminator {
		return false
	}
	if len(p.SchemaPath) != len(other.SchemaPath) {
		return false
	}
	for i := range p.SchemaPath {
		if p.SchemaPath[i] != other.SchemaPath[i] {
			return false
		}
	}
	return true
}

// Extension is one configuration plugged into a Point.
type Extension struct {
	// Point is the extension site this configuration plugs into.
	Point Point
	// Value is the discriminator value selecting this configuration, for example "Route53".
	Value string
	// Config is a pointer to a zero value of the struct describing the `config`
	// payload, for example &Route53{}. Only its type is read.
	Config any
}

// ConfigType is the struct type of the `config` payload.
func (e Extension) ConfigType() reflect.Type {
	return reflect.TypeOf(e.Config).Elem()
}

var (
	mtx        sync.Mutex
	registered = map[string]Extension{}
)

func key(resourceType model.ResourceType, value string) string {
	return string(resourceType) + "/" + value
}

// Register adds an extension configuration to the global registry. Call it from an
// init function in the package that owns the configuration. It panics on anything
// that would produce an unusable schema, because there is no way to recover from a
// bad registration at generation time.
func Register(ext Extension) {
	if err := validate(ext); err != nil {
		panic(fmt.Errorf("invalid extension registration: %w", err))
	}

	mtx.Lock()
	defer mtx.Unlock()

	k := key(ext.Point.ResourceType, ext.Value)
	if existing, ok := registered[k]; ok {
		panic(fmt.Errorf("extension %q is already registered on %s with config %s",
			ext.Value, ext.Point.ResourceType, existing.ConfigType()))
	}
	for _, existing := range registered {
		if existing.Point.ResourceType == ext.Point.ResourceType && !existing.Point.equal(ext.Point) {
			panic(fmt.Errorf("extension %q declares a different extension point on %s than %q does",
				ext.Value, ext.Point.ResourceType, existing.Value))
		}
	}
	registered[k] = ext
}

func validate(ext Extension) error {
	switch {
	case ext.Point.ResourceType == "":
		return fmt.Errorf("point resource type must not be empty")
	case len(ext.Point.SchemaPath) == 0:
		return fmt.Errorf("point schema path must not be empty")
	case ext.Point.Discriminator == "":
		return fmt.Errorf("point discriminator must not be empty")
	case ext.Value == "":
		return fmt.Errorf("value must not be empty")
	case ext.Config == nil:
		return fmt.Errorf("config must not be nil")
	}
	t := reflect.TypeOf(ext.Config)
	if t.Kind() != reflect.Pointer || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct, got %s", t)
	}
	if t.Elem().Name() == "" {
		return fmt.Errorf("config must be a named struct type")
	}
	return nil
}

// Registered returns every registered extension, ordered by resource type then
// discriminator value so generated output does not depend on map iteration.
func Registered() []Extension {
	mtx.Lock()
	defer mtx.Unlock()

	all := make([]Extension, 0, len(registered))
	for _, ext := range registered {
		all = append(all, ext)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Point.ResourceType != all[j].Point.ResourceType {
			return all[i].Point.ResourceType < all[j].Point.ResourceType
		}
		return all[i].Value < all[j].Value
	})
	return all
}
