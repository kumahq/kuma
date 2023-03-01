package secrets

type ChangeKind int

const (
	IdentityChange ChangeKind = iota
	OwnMeshChange
	OtherMeshChange
)

type UpdateKinds map[ChangeKind]struct{}

func UpdateEverything() UpdateKinds {
	return map[ChangeKind]struct{}{
		IdentityChange:  {},
		OwnMeshChange:   {},
		OtherMeshChange: {},
	}
}

func (kinds UpdateKinds) HasType(kind ChangeKind) bool {
	_, ok := kinds[kind]
	return ok
}

func (kinds UpdateKinds) AddKind(kind ChangeKind) {
	kinds[kind] = struct{}{}
}
