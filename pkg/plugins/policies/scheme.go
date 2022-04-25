package policies

import (
	"k8s.io/apimachinery/pkg/runtime"
)

func AddToScheme(s *runtime.Scheme) error {
	// Example:
	// if err := my_new_policy.AddToScheme(s); err != nil {
	//    return err
	//}
	return nil
}
