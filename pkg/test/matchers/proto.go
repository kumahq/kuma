package matchers

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

func (p *ProtoMatcher) FailureMessage(actual interface{}) (message string) {
	actualYAML, expectedYAML, err := p.yamls(actual)
	if err != nil {
		return err.Error()
	}
	return format.Message(actualYAML, "to equal", expectedYAML)
}

func (p *ProtoMatcher) yamls(actual interface{}) (string, string, error) {
	actualProto := actual.(proto.Message)
	actualYAML, err := util_proto.ToYAML(actualProto)
	if err != nil {
		return "", "", fmt.Errorf("Proto are not equal (could not convert to YAML: %s)", err)
	}

	expectedProto := p.Expected.(proto.Message)
	expectedYAML, err := util_proto.ToYAML(expectedProto)
	if err != nil {
		return "", "", fmt.Errorf("Proto are not equal (could not convert to YAML: %s)", err)
	}
	return string(actualYAML), string(expectedYAML), nil
}

func (p *ProtoMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	actualYAML, expectedYAML, err := p.yamls(actual)
	if err != nil {
		return err.Error()
	}
	return format.Message(actualYAML, "not to equal", expectedYAML)
}

var _ types.GomegaMatcher = &ProtoMatcher{}
