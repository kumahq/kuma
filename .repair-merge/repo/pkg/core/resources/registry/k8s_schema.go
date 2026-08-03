package registry

import "k8s.io/apimachinery/pkg/runtime"

var AllKubeSchemes []func(*runtime.Scheme) error

func AddKubeScheme(fn func(scheme *runtime.Scheme) error) {
	AllKubeSchemes = append(AllKubeSchemes, fn)
}

func AddToScheme(s *runtime.Scheme) error {
	for i := range AllKubeSchemes {
		if err := AllKubeSchemes[i](s); err != nil {
			return err
		}
	}
	return nil
}
