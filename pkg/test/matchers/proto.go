package matchers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func MatchProto(expected interface{}) types.GomegaMatcher {
	return &ProtoMatcher{
		Expected: expected,
	}
}

type ProtoMatcher struct {
	Expected interface{}
}

func (p *ProtoMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil && p.Expected == nil {
		return true, nil
	}
	if actual == nil && p.Expected != nil {
		return false, errors.New("Actual object is nil, but Expected object is not.")
	}
	if actual != nil && p.Expected == nil {
		return false, errors.New("Actual object is not nil, but Expected object is.")
	}

	actualProto, ok := actual.(proto.Message)
	if !ok {
		return false, errors.New("You can only compare proto with this matcher. Make sure the object passed to MatchProto() implements proto.Message")
	}

	expectedProto, ok := p.Expected.(proto.Message)
	if !ok {
		return false, errors.New("You can only compare proto with this matcher. Make sure the object passed to Expect() implements proto.Message")
	}

	return proto.Equal(actualProto, expectedProto), nil
}

func (p *ProtoMatcher) FailureMessage(actual interface{}) string {
	differences := cmp.Diff(p.Expected, actual, protocmp.Transform())
	return "Expected matching protobuf message:\n" + differences
}

func (p *ProtoMatcher) NegatedFailureMessage(actual interface{}) string {
	return "Expected different protobuf but was the same"
}

var _ types.GomegaMatcher = &ProtoMatcher{}
