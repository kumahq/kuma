package manager

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Add(mgr manager.Manager, rs ...manager.Runnable) error {
	for _, r := range rs {
		if err := mgr.Add(r); err != nil {
			return err
		}
	}
	return nil
}
