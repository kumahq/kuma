package matchers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func MatchProto(expected any) types.GomegaMatcher {
	return &ProtoMatcher{
		Expected: expected,
	}
}

type ProtoMatcher struct {
	Expected any
}

func (p *ProtoMatcher) Match(actual any) (bool, error) {
	if actual == nil && p.Expected == nil {
		return true, nil
	}
	if actual == nil && p.Expected != nil {
		return false, errors.New("Actual object is nil, but Expected object is not.")
	}
	if actual != nil && p.Expected == nil {
		return false, errors.New("Actual object is not nil, but Expected object is.")
	}

	actualProto, actualIsProto := actual.(proto.Message)
	expectedProto, expectedIsProto := p.Expected.(proto.Message)

	switch {
	case actualIsProto && expectedIsProto:
		return proto.Equal(actualProto, expectedProto), nil
	case !actualIsProto && !expectedIsProto:
		// Resource specs are being converted from protobuf to Go structs, so this
		// matcher compares whichever of the two it is handed as long as both sides
		// are the same kind. EquateEmpty keeps proto.Equal's reading of an unset
		// list or map, which the JSON both forms are stored as cannot tell apart
		// either.
		return cmp.Diff(p.Expected, actual, protocmp.Transform(), cmpopts.EquateEmpty()) == "", nil
	case actualIsProto:
		return false, errors.New("Actual object is a proto.Message, but Expected object is not.")
	default:
		return false, errors.New("Expected object is a proto.Message, but Actual object is not.")
	}
}

func (p *ProtoMatcher) FailureMessage(actual any) string {
	differences := cmp.Diff(p.Expected, actual, protocmp.Transform(), cmpopts.EquateEmpty())
	return "Expected matching message:\n" + differences
}

func (p *ProtoMatcher) NegatedFailureMessage(actual any) string {
	return "Expected different protobuf but was the same"
}

var _ types.GomegaMatcher = &ProtoMatcher{}
