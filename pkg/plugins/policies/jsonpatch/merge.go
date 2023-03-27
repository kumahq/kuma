package jsonpatch

import (
	"fmt"

	"github.com/evanphx/json-patch/v5"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var patchApplyOptions = &jsonpatch.ApplyOptions{
	AccumulatedCopySizeLimit: 0,
	SupportNegativeIndices:   true,
	AllowMissingPathOnRemove: true,
	EnsurePathExistsOnAdd:    true,
}

func MergeJsonPatch(dst proto.Message, patchBlock []common_api.JsonPatchBlock) error {
	return patch(patchBlock, dst, dst.ProtoReflect().New())
}

func MergeJsonPatchAny(
	dst *anypb.Any,
	patchBlock []common_api.JsonPatchBlock,
) (*anypb.Any, error) {
	msgType, err := util_proto.FindMessageType(dst.TypeUrl)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to find message type (%s): %w", dst.TypeUrl, err,
		)
	}

	dstMsg := msgType.New().Interface()
	if err := proto.Unmarshal(dst.Value, dstMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proto: %w", err)
	}

	if err := patch(patchBlock, dstMsg, msgType.New()); err != nil {
		return nil, err
	}

	marshaled, err := util_proto.MarshalAnyDeterministic(dstMsg)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal proto message to proto.Any: %w", err,
		)
	}

	return marshaled, nil
}

func patch(
	patchBlock []common_api.JsonPatchBlock,
	dst protoreflect.ProtoMessage,
	src protoreflect.Message,
) error {
	resource, err := util_proto.ToJSON(dst)
	if err != nil {
		return fmt.Errorf("failed to patch: parse proto resource to json: %w", err)
	}

	jsonPatch, err := common_api.ToJsonPatch(patchBlock)
	if err != nil {
		return fmt.Errorf(
			"failed to patch: convert block to json patch: %w",
			err,
		)
	}

	patched, err := jsonPatch.ApplyWithOptions(resource, patchApplyOptions)
	if err != nil {
		return fmt.Errorf("failed to patch: apply json patches: %w", err)
	}

	mod := src.Interface()

	if err := util_proto.FromJSON(patched, mod); err != nil {
		return fmt.Errorf(
			"failed to patch: parse json patched resource to proto: %w",
			err,
		)
	}

	util_proto.Replace(dst, mod)

	return nil
}
