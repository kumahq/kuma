package model

const (
	// NamePattern is the pattern every resource name must match. It lives here
	// rather than next to the validator so code generators can read it without
	// depending on the packages they generate.
	NamePattern = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	// MeshNamePattern is laxer than NamePattern: a Mesh created before the
	// stricter rule may still contain '_', and mesh-scoped resources have to keep
	// referring to it.
	MeshNamePattern = `^[0-9a-z-_.]*$`
	// MaxNameLength is the maximum length of a resource name or mesh reference.
	MaxNameLength = 253
)
