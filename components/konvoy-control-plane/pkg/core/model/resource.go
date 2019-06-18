package model

type Resource interface {
	GetType() ResourceType
	GetMeta() ResourceMeta
	GetSpec() ResourceSpec
}

type ResourceType string

type ResourceMeta interface {
	GetName() string
	GetNamespace() string
	GetVersion() string
}

type ResourceSpec interface {
}

type ResourceList interface {
	GetItemType() ResourceType
	GetItems() []Resource
}
