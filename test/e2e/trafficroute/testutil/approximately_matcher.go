package testutil

import (
	"fmt"
	"math"
	"reflect"

	"github.com/onsi/gomega/types"
)

const defaultErrorRate = 3

// ApproximatelyEqual checks that 'actual' differs from 'expected' by no more than 'errorRate'.
// By default 'errorRate' is equal to 3, but it could be passed as the first 'args':
//     Expect(78).To(ApproximatelyEqual(80, 3))
// This line checks that 80-3 <= 78 <= 80+3
func ApproximatelyEqual(value interface{}, args ...interface{}) types.GomegaMatcher {
	var errorRate float64 = defaultErrorRate
	if len(args) > 0 {
		if fl, err := getFloat(args[0]); err == nil {
			errorRate = fl
		}
	}
	return &ApproximatelyMatcher{Expected: value, ErrorRate: errorRate}
}

type ApproximatelyMatcher struct {
	Expected  interface{}
	ErrorRate float64
}

func (matcher *ApproximatelyMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}

	actualFloat, err := getFloat(actual)
	if err != nil {
		return false, err
	}

	expectedFloat, err := getFloat(matcher.Expected)
	if err != nil {
		return false, err
	}

	diff := math.Abs(expectedFloat - actualFloat)
	return diff <= matcher.ErrorRate, nil
}

func (matcher *ApproximatelyMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%v is approximately equal to %v", actual, matcher.Expected)
}

func (matcher *ApproximatelyMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("%v is not approximately equal to %v", actual, matcher.Expected)
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}
