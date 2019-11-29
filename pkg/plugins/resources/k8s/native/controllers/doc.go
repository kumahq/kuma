package controllers

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch

// +kubebuilder:rbac:groups=kuma.io,resources=dataplanes,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=kuma.io,resources=dataplaneinsights,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=kuma.io,resources=healthchecks,verbs=get;list;watch

// +kubebuilder:rbac:groups=kuma.io,resources=trafficroutes,verbs=get;list;watch
