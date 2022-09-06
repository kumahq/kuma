ENVOY_IMPORTS := ./pkg/xds/envoy/imports.go
PROTO_DIRS := ./pkg/config ./api

CONTROLLER_GEN := go run -mod=mod sigs.k8s.io/controller-tools/cmd/controller-gen
RESOURCE_GEN := go run -mod=mod ./tools/resource-gen/main.go

.PHONY: clean/proto
clean/proto: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIRS) -name '*.pb.go' -delete
	find $(PROTO_DIRS) -name '*.pb.validate.go' -delete

.PHONY: generate
generate:  ## Dev: Run code generators
generate: clean/proto generate/api protoc/pkg/config/app/kumactl/v1alpha1 protoc/pkg/test/apis/sample/v1alpha1 protoc/plugins resources/type generate/policies

.PHONY: resources/type
resources/type:
	$(GO_RUN) ./tools/resource-gen/main.go -package mesh -generator type > pkg/core/resources/apis/mesh/zz_generated.resources.go
	$(GO_RUN) ./tools/resource-gen/main.go -package system -generator type > pkg/core/resources/apis/system/zz_generated.resources.go

.PHONY: protoc/pkg/config/app/kumactl/v1alpha1
protoc/pkg/config/app/kumactl/v1alpha1:
	$(PROTOC_GO) pkg/config/app/kumactl/v1alpha1/*.proto

.PHONY: protoc/pkg/test/apis/sample/v1alpha1
protoc/pkg/test/apis/sample/v1alpha1:
	$(PROTOC_GO) pkg/test/apis/sample/v1alpha1/*.proto

.PHONY: protoc/plugins
protoc/plugins:
	$(PROTOC_GO) --proto_path=./api pkg/plugins/ca/provided/config/*.proto
	$(PROTOC_GO) --proto_path=./api pkg/plugins/ca/builtin/config/*.proto

POLICIES_DIR := pkg/plugins/policies

policies = $(foreach dir,$(shell find pkg/plugins/policies -maxdepth 1 -mindepth 1 -type d | grep -v core | grep -v matchers),$(notdir $(dir)))
generate_policy_targets = $(addprefix generate/policy/,$(policies))
cleanup_policy_targets = $(addprefix cleanup/policy/,$(policies))

generate/policies: cleanup/crds $(cleanup_policy_targets) $(generate_policy_targets) generate/policy-import generate/policy-helm generate/builtin-crds

cleanup/crds:
	rm -f ./deployments/charts/kuma/crds/*
	rm -f ./pkg/plugins/resources/k8s/native/test/config/crd/bases/*

# deletes all files in policy directory except *.proto and validator.go
cleanup/policy/%:
	$(shell find $(POLICIES_DIR)/$* \( -name '*.pb.go' -o -name '*.yaml' -o -name 'zz_generated.*'  \) -type f -delete)
	@rm -r $(POLICIES_DIR)/$*/k8s || true

generate/policy/%: generate/schema/%
	@echo "Policy $* successfully generated"

generate/schema/%: generate/controller-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		PATH=$(CI_TOOLS_BIN_DIR):$$PATH tools/policy-gen/crd-extract-openapi.sh $* $$version ; \
	done

generate/policy-import:
	tools/policy-gen/generate-policy-import.sh $(policies)

generate/policy-helm:
	PATH=$(CI_TOOLS_BIN_DIR):$$PATH tools/policy-gen/generate-policy-helm.sh $(policies)

generate/controller-gen/%: generate/kumapolicy-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) "crd:crdVersions=v1,ignoreUnexportedFields=true" paths="./$(POLICIES_DIR)/$*/k8s/..." output:crd:artifacts:config=$(POLICIES_DIR)/$*/k8s/crd && \
		$(CONTROLLER_GEN) object:headerFile=./tools/policy-gen/boilerplate.go.txt,year=$$(date +%Y) paths=$(POLICIES_DIR)/$*/k8s/$$version/zz_generated.types.go ; \
	done

generate/kumapolicy-gen/%: generate/dirs/%
	cd tools/policy-gen/protoc-gen-kumapolicy && go build && cd - ; \
	$(PROTOC_GO) \
		--proto_path=./api \
		--kumapolicy_opt=endpoints-template=tools/policy-gen/templates/endpoints.yaml \
		--kumapolicy_out=$(POLICIES_DIR)/$* \
		--plugin=protoc-gen-kumapolicy=tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy \
		$(POLICIES_DIR)/$*/api/*/*.proto ; \
	rm tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy ; \

generate/dirs/%:
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		mkdir -p $(POLICIES_DIR)/$*/api/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/crd ; \
	done

.PHONY: generate/builtin-crds
generate/builtin-crds:
	$(RESOURCE_GEN) -package mesh -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.mesh.go
	$(RESOURCE_GEN) -package system -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.system.go
	$(MAKE) OUT_CRD=./deployments/charts/kuma/crds IN_CRD=./pkg/plugins/resources/k8s/native/api/... crd/controller-gen
	$(MAKE) OUT_CRD=./pkg/plugins/resources/k8s/native/test/config/crd/bases IN_CRD=./pkg/plugins/resources/k8s/native/test/api/sample/... crd/controller-gen

.PHONY: crd/controller-gen
crd/controller-gen:
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths=$(IN_CRD) output:crd:artifacts:config=$(OUT_CRD)
	$(CONTROLLER_GEN) object:headerFile=./tools/policy-gen/boilerplate.go.txt,year=$$(date +%Y) paths=$(IN_CRD)

.PHONY: generate/envoy-imports
generate/envoy-imports:
	echo 'package envoy\n' > ${ENVOY_IMPORTS}
	echo '// Import all Envoy packages so protobuf are registered and are ready to used in functions such as MarshalAny.' >> ${ENVOY_IMPORTS}
	echo '// This file is autogenerated. run "make generate/envoy-imports" to regenerate it after go-control-plane upgrade' >> ${ENVOY_IMPORTS}
	echo 'import (' >> ${ENVOY_IMPORTS}
	go list github.com/envoyproxy/go-control-plane/... | grep "github.com/envoyproxy/go-control-plane/envoy/" | awk '{printf "\t_ \"%s\"\n", $$1}' >> ${ENVOY_IMPORTS}
	echo ')' >> ${ENVOY_IMPORTS}

.PHONY: generate/api
generate/api: protoc/common/v1alpha1 protoc/mesh protoc/mesh/v1alpha1 protoc/observability/v1 protoc/system/v1alpha1 ## Process Kuma API .proto definitions

generate/test-server:
	$(PROTOC_GO) test/server/grpc/api/*.proto
