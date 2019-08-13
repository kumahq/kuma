package controllers

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

// +kubebuilder:rbac:groups=mesh.getkonvoy.io,resources=dataplanes,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=mesh.getkonvoy.io,resources=dataplaneinsights,verbs=get;list;watch;create;update;patch;delete
