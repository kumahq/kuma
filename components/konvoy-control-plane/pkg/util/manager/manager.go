package manager

import "sigs.k8s.io/controller-runtime/pkg/manager"

func Add(mgr manager.Manager, rs ...manager.Runnable) error {
	for _, r := range rs {
		if err := mgr.Add(r); err != nil {
			return err
		}
	}
	return nil
}

type Manageable interface {
	SetupWithManager(mgr manager.Manager) error
}

func SetupWithManager(mgr manager.Manager, ms ...Manageable) error {
	for _, m := range ms {
		if err := m.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
