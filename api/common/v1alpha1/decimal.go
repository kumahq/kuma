package v1alpha1

import (
	"fmt"

	"github.com/shopspring/decimal"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewDecimalFromIntOrString(intOrString intstr.IntOrString) (decimal.Decimal, error) {
	switch intOrString.Type {
	case intstr.Int:
		return decimal.NewFromInt(int64(intOrString.IntVal)), nil
	case intstr.String:
		return decimal.NewFromString(intOrString.String())
	default:
		return decimal.Zero, fmt.Errorf("invalid IntOrString '%s'", intOrString.String())
	}
}
