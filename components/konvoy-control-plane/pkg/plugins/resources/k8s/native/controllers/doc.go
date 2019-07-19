package controllers

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// +kubebuilder:rbac:groups=mesh.getkonvoy.io,resources=dataplanes,verbs=get;list;watch;create;update;patch;delete
