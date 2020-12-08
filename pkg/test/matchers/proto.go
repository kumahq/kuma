package matchers

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/onsi/gomega/types"
)

func MatchProto(expected interface{}) types.GomegaMatcher {
	return &ProtoMatcher{
		Expected: expected,
	}
}

type ProtoMatcher struct {
	Expected interface{}
}

func (p *ProtoMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && p.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}

	actualProto, ok := actual.(proto.Message)
	if !ok {
		return false, errors.New("you can only compare proto with this matcher")
	}

	expectedProto, ok := p.Expected.(proto.Message)
	if !ok {
		return false, errors.New("you can only compare proto with this matcher")
	}

	return proto.Equal(actualProto, expectedProto), nil
}

func (p *ProtoMatcher) FailureMessage(actual interface{}) (message string) {
	return "proto are not equal" // todo
}

func (p *ProtoMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return "proto are not equal" // todo
}

var _ types.GomegaMatcher = &ProtoMatcher{}
