package validators

import (
	"encoding/json"
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func ValidateJsonPatchBlock(
	rootPath validators.PathBuilder,
	patchBlock []common_api.JsonPatchBlock,
) validators.ValidationError {
	var err validators.ValidationError

	if len(patchBlock) == 0 {
		err.AddViolationAt(rootPath, validators.MustNotBeEmpty)
	}

	for i, patch := range patchBlock {
		path := rootPath.Index(i)

		err.AddErrorAt(path.Field("op"), validateOp(patch.Op))
		err.AddErrorAt(path.Field("from"), validateFrom(patch.From, patch.Op))
		err.AddErrorAt(path.Field("value"), validateValue(patch.Value, patch.Op))
		err.AddErrorAt(path.Field("path"), validatePath(patch.Path, patch.Op))
	}

	return err
}

// validateOf checks if "op" field is valid json patch operation
func validateOp(op string) validators.ValidationError {
	var err validators.ValidationError

	switch op {
	case "add", "remove", "replace", "move", "copy":
		break
	default:
		err.Add(validators.MakeFieldMustBeOneOfErr("op",
			"add", "remove", "replace", "move", "copy",
		))
	}

	return err
}

// validateFrom checks if "op" field is valid ("move" or "copy") when "from"
// field is provided
func validateFrom(from *string, op string) validators.ValidationError {
	if op == "move" || op == "copy" {
		return validateFromOpMoveCopy(from)
	}

	if from != nil {
		return validators.MakeOneOfErr("from",
			"op",
			"is allowed only",
			[]string{"move", "copy"},
		)
	}

	return validators.OK()
}

func validateFromOpMoveCopy(from *string) validators.ValidationError {
	if from == nil {
		return validators.MakeOneOfErr("from",
			"op",
			validators.MustNotBeEmpty,
			[]string{"move", "copy"},
		)
	}

	return validators.OK()
}

// validateValue checks if "op" field is valid ("add" or "replace") when "value"
// field is provided
func validateValue(value json.RawMessage, op string) validators.ValidationError {
	if op == "add" || op == "replace" {
		return validateValueOpAddReplace(value)
	}

	if value != nil {
		return validators.MakeOneOfErr("value",
			"op",
			"is allowed only",
			[]string{"add", "replace"},
		)
	}

	return validators.OK()
}

func validateValueOpAddReplace(value json.RawMessage) validators.ValidationError {
	if value == nil {
		return validators.MakeOneOfErr("value",
			"op",
			validators.MustNotBeEmpty,
			[]string{"add", "replace"},
		)
	}

	return validators.OK()
}

func validatePath(path *string, op string) validators.ValidationError {
	var err validators.ValidationError

	if path == nil {
		err.AddViolationAt(validators.Root(), validators.MustBeDefined)
	} else if (*path == "" || *path == "/") && op == "remove" {
		err.AddViolationAt(validators.Root(), "root path cannot be removed")
	}

	return err
}

func TopLevelTargetRefDeprecations(targetRef *common_api.TargetRef) []string {
	if targetRef == nil {
		return nil
	}
	if targetRef.Kind == common_api.MeshService {
		return []string{
			fmt.Sprintf("%s value for 'targetRef.kind' is deprecated, use %s with '%s' tag instead", common_api.MeshService, common_api.MeshSubset, mesh_proto.ServiceTag),
		}
	}
	return nil
}
