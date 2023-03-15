package v1alpha1

import (
	"encoding/json"
	"strconv"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// JsonPatchBlock is one json patch operation block.
type JsonPatchBlock struct {
	// Op is a jsonpatch operation string.
	// +required
	// +kubebuilder:validation:Enum=add;remove;replace;move;copy
	Op string `json:"op"`
	// Path is a jsonpatch path string.
	// +required
	Path string `json:"path"`
	// Value must be a valid json object used by replace and add operations.
	Value json.RawMessage `json:"value,omitempty"`
	// From is a jsonpatch from string, used by move and copy operations.
	From string `json:"from,omitempty"`
}

func ToJsonPatch(in []JsonPatchBlock) jsonpatch.Patch {
	var res []jsonpatch.Operation

	for _, o := range in {
		op := json.RawMessage(strconv.Quote(o.Op))
		path := json.RawMessage(strconv.Quote(o.Path))
		from := json.RawMessage(strconv.Quote(o.From))
		value := o.Value

		res = append(res, jsonpatch.Operation{
			"op":    &op,
			"path":  &path,
			"from":  &from,
			"value": &value,
		})
	}

	return res
}

func MergeJsonPatch(message proto.Message, patchBlock []JsonPatchBlock) error {
	resourceJson, err := util_proto.ToJSON(message)
	if err != nil {
		return err
	}

	options := jsonpatch.NewApplyOptions()
	options.EnsurePathExistsOnAdd = true
	options.AllowMissingPathOnRemove = true

	patchedJson, err := ToJsonPatch(patchBlock).ApplyWithOptions(resourceJson, options)
	if err != nil {
		return err
	}

	mod := message.ProtoReflect().New().Interface()

	if err := util_proto.FromJSON(patchedJson, mod); err != nil {
		return err
	}

	util_proto.Replace(message, mod)

	return nil
}

func MergeJsonPatchAny(dst *anypb.Any, patchBlock []JsonPatchBlock) (*anypb.Any, error) {
	// TypeURL in Any contains type.googleapis.com/ prefix, but in Proto registry
	// it does not have this prefix.
	msgTypeName := strings.ReplaceAll(dst.TypeUrl, util_proto.GoogleApisTypeUrl, "")
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgTypeName))
	if err != nil {
		return nil, err
	}

	dstMsg := msgType.New().Interface()
	if err := proto.Unmarshal(dst.Value, dstMsg); err != nil {
		return nil, err
	}

	resourceJson, err := util_proto.ToJSON(dstMsg)
	if err != nil {
		return nil, err
	}

	options := jsonpatch.NewApplyOptions()
	options.EnsurePathExistsOnAdd = true
	options.AllowMissingPathOnRemove = true

	patchedJson, err := ToJsonPatch(patchBlock).ApplyWithOptions(resourceJson, options)
	if err != nil {
		return nil, err
	}

	mod := msgType.New().Interface()

	if err := util_proto.FromJSON(patchedJson, mod); err != nil {
		return nil, err
	}

	util_proto.Replace(dstMsg, mod)

	return util_proto.MarshalAnyDeterministic(dstMsg)
}
