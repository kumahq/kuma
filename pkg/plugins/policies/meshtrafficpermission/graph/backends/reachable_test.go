package backends

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	test_kumactl "github.com/kumahq/kuma/app/kumactl/pkg/test"
	ms_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/k8s/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	v1alpha12 "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/k8s/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_http "github.com/kumahq/kuma/pkg/util/http"
)

func TestB(t *testing.T) {
	subset := core_rules.SubsetFromTags(generateFromTags())
	// [{app nginx false}
	// {k8s.kuma.io/namespace kuma-one false}
	// {kuma.io/zone default false}
	// {kuma.io/service nginx-service_kuma-one_svc_80 false}
	// {pod-template-hash 69fc5cd7d5 false}
	// {k8s.kuma.io/service-name nginx-service false}
	// {k8s.kuma.io/service-port 80 false}
	// {kubernetes.io/hostname k3d-kuma-1-server-0 false}
	// {kuma.io/protocol http false}]
	fmt.Println(subset)
}

func TestA(t *testing.T) {
	{
		var rootCtx *kumactl_cmd.RootContext
		var store core_store.ResourceStore
		rootCtx = test_kumactl.MakeMinimalRootContext()
		rootCtx.Runtime.Registry = registry.Global()
		rootCtx.Runtime.NewResourceStore = func(util_http.Client) core_store.ResourceStore {
			return store
		}
		store = core_store.NewPaginationStore(memory_resources.NewStore())
	}

	rules := BuildRules(generateMeshServices(), generateMTP())
	for k, v := range rules {
		fmt.Println(k)
		fmt.Println("========")
		for index := range v {
			fmt.Println("++++index: ", index)
			tmpRule := *v[index]
			fmt.Printf("rule: %#v \n", tmpRule)
		}
	}

	tmpIdentifierRule := rules[generateTypedResourceIdentifier()]

	// {MTP-tags} compute {MeshService-tags}
	// MTP-tags 与 MeshService-tags 中key重叠的k-v拿出来比较，看看后者是否属于前者的子集
	rule := tmpIdentifierRule.Compute(core_rules.SubsetFromTags(generateFromTags()))
	if rule == nil {
		fmt.Println("++++++++++++++++  NIL")
		return
	}

	action := rule.Conf.(mtp_api.Conf).Action
	fmt.Println(action) // should be Allow
}

func generateTypedResourceIdentifier() core_model.TypedResourceIdentifier {
	return core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.ResourceIdentifier{
			Name:      "nginx-service",
			Mesh:      "default",
			Namespace: "kuma-one",
			Zone:      "default",
		},
		ResourceType: "MeshService",
	}
}

func generateFromTags() map[string]string {
	dic := map[string]string{
		"app":                      "nginx",
		"k8s.kuma.io/namespace":    "kuma-one",
		"k8s.kuma.io/service-name": "nginx-service",
		"k8s.kuma.io/service-port": "80",
		"kubernetes.io/hostname":   "k3d-kuma-1-server-0",
		"kuma.io/protocol":         "http",
		"kuma.io/service":          "nginx-service_kuma-one_svc_80",
		"kuma.io/zone":             "default",
		"pod-template-hash":        "69fc5cd7d5",
	}
	return dic
}

func generateMeshServices() []*ms_api.MeshServiceResource {
	return []*ms_api.MeshServiceResource{
		//generateMeshServiceOne(),
		generateMeshServiceTwo(),
	}
}

func generateMeshServiceOne() *ms_api.MeshServiceResource {
	k8sMeshServiceStr := `apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  annotations:
    kuma.io/display-name: service-2048-app
  creationTimestamp: "2024-12-11T15:10:05Z"
  generation: 5
  labels:
    k8s.kuma.io/is-headless-service: "false"
    k8s.kuma.io/namespace: kuma-demo
    k8s.kuma.io/service-name: service-2048-app
    kuma.io/env: kubernetes
    kuma.io/managed-by: k8s-controller
    kuma.io/mesh: default
    kuma.io/origin: zone
    kuma.io/zone: default
  name: service-2048-app
  namespace: kuma-demo
  ownerReferences:
  - apiVersion: v1
    kind: Service
    name: service-2048-app
    uid: 97c366ef-4bfd-4667-8151-2ec6b31afa8c
  resourceVersion: "16352"
  uid: 5dd6d86c-f594-4ee6-9cc3-442c9b0f8b3c
spec:
  identities:
  - type: ServiceTag
    value: service-2048-app_kuma-demo_svc_80
  ports:
  - appProtocol: http
    name: http
    port: 80
    targetPort: 80
  selector:
    dataplaneTags:
      app: 2048-app
      k8s.kuma.io/namespace: kuma-demo
  state: Available
status:
  addresses:
  - hostname: service-2048-app.kuma-demo.default.gg.mesh
    hostnameGeneratorRef:
      coreName: default.kuma-system
    origin: HostnameGenerator
  dataplaneProxies:
    connected: 1
    healthy: 1
    total: 1
  hostnameGenerators:
  - conditions:
    - message: ""
      reason: Generated
      status: "True"
      type: Generated
    hostnameGeneratorRef:
      coreName: default.kuma-system
  tls:
    status: Ready
  vips:
  - ip: 10.43.129.183`

	reader := strings.NewReader(k8sMeshServiceStr)
	all, err := io.ReadAll(reader)
	if nil != err {
		panic(err)
	}

	var meshService v1alpha1.MeshService

	err = yaml.Unmarshal(all, &meshService)
	if nil != err {
		panic(err)
	}

	meshServiceResource := ms_api.NewMeshServiceResource()
	meshServiceResource.Spec = meshService.Spec
	meshServiceResource.Meta = &k8s.KubernetesMetaAdapter{
		ObjectMeta: meshService.ObjectMeta,
		Mesh:       meshService.GetMesh(),
	}

	return meshServiceResource
}

func generateMeshServiceTwo() *ms_api.MeshServiceResource {
	k8sMeshServiceStr := `
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  annotations:
    kuma.io/display-name: nginx-service
  creationTimestamp: "2024-12-11T15:11:53Z"
  generation: 5
  labels:
    k8s.kuma.io/is-headless-service: "false"
    k8s.kuma.io/namespace: kuma-one
    k8s.kuma.io/service-name: nginx-service
    kuma.io/env: kubernetes
    kuma.io/managed-by: k8s-controller
    kuma.io/mesh: default
    kuma.io/origin: zone
    kuma.io/zone: default
  name: nginx-service
  namespace: kuma-one
  ownerReferences:
  - apiVersion: v1
    kind: Service
    name: nginx-service
    uid: 7abbe08d-0bf3-427c-956d-b5f5b1d1baab
  resourceVersion: "16448"
  uid: 279545b7-7bee-4184-a1be-4b6a342a259b
spec:
  identities:
  - type: ServiceTag
    value: nginx-service_kuma-one_svc_80
  ports:
  - appProtocol: http
    name: name-of-service-port
    port: 80
    targetPort: http-web-svc
  selector:
    dataplaneTags:
      app: nginx
      k8s.kuma.io/namespace: kuma-one
  state: Available
status:
  addresses:
  - hostname: nginx-service.kuma-one.default.gg.mesh
    hostnameGeneratorRef:
      coreName: default.kuma-system
    origin: HostnameGenerator
  dataplaneProxies:
    connected: 1
    healthy: 1
    total: 1
  hostnameGenerators:
  - conditions:
    - message: ""
      reason: Generated
      status: "True"
      type: Generated
    hostnameGeneratorRef:
      coreName: default.kuma-system
  tls:
    status: Ready
  vips:
  - ip: 10.43.99.95
`

	reader := strings.NewReader(k8sMeshServiceStr)
	all, err := io.ReadAll(reader)
	if nil != err {
		panic(err)
	}

	var meshService v1alpha1.MeshService

	err = yaml.Unmarshal(all, &meshService)
	if nil != err {
		panic(err)
	}

	meshServiceResource := ms_api.NewMeshServiceResource()
	meshServiceResource.Spec = meshService.Spec
	meshServiceResource.Meta = &k8s.KubernetesMetaAdapter{
		ObjectMeta: meshService.ObjectMeta,
		Mesh:       meshService.GetMesh(),
	}

	return meshServiceResource
}

func generateMTP() []*mtp_api.MeshTrafficPermissionResource {
	data := []*mtp_api.MeshTrafficPermissionResource{
		goodMTP(),
		badMTP(),
	}

	return data
}

func goodMTP() *mtp_api.MeshTrafficPermissionResource {
	goodMTPStr := `
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kuma.io/v1alpha1","kind":"MeshTrafficPermission","metadata":{"annotations":{},"labels":{"kuma.io/mesh":"default"},"name":"mtp-allow-kuma-one","namespace":"kuma-system"},"spec":{"from":[{"default":{"action":"Allow"},"targetRef":{"kind":"MeshSubset","tags":{"k8s.kuma.io/namespace":"kuma-one"}}}],"targetRef":{"kind":"Mesh"}}}
    kuma.io/display-name: mtp-allow-kuma-one
  creationTimestamp: "2024-12-11T08:43:58Z"
  generation: 1
  labels:
    k8s.kuma.io/namespace: kuma-system
    kuma.io/env: kubernetes
    kuma.io/mesh: default
    kuma.io/origin: zone
    kuma.io/policy-role: system
    kuma.io/zone: default
  name: mtp-allow-kuma-one
  namespace: kuma-system
  ownerReferences:
  - apiVersion: kuma.io/v1alpha1
    kind: Mesh
    name: default
    uid: ec9627ca-1846-4875-b31b-1ad13eb0838d
  resourceVersion: "6068"
  uid: 5a426857-54b2-47e6-a2ca-8c4441d3077a
spec:
  from:
  - targetRef:
      kind: MeshSubset
      tags:
        k8s.kuma.io/namespace: kuma-one
    default:
      action: Allow
  - targetRef:
      kind: MeshSubset
      tags:
        app: backend
        k8s.kuma.io/namespace: kuma-one
    default:
      action: Deny
  targetRef:
    kind: Mesh
`
	newReader := strings.NewReader(goodMTPStr)
	all, err := io.ReadAll(newReader)
	if nil != err {
		panic(err)
	}

	var meshTP v1alpha12.MeshTrafficPermission
	err = yaml.Unmarshal(all, &meshTP)
	if nil != err {
		panic(err)
	}

	var result mtp_api.MeshTrafficPermissionResource
	result.Spec = meshTP.Spec
	result.SetMeta(&k8s.KubernetesMetaAdapter{
		ObjectMeta: meshTP.ObjectMeta,
		Mesh:       meshTP.GetMesh(),
	})

	return &result
}

func badMTP() *mtp_api.MeshTrafficPermissionResource {
	badMTPStr := `
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kuma.io/v1alpha1","kind":"MeshTrafficPermission","metadata":{"annotations":{},"labels":{"kuma.io/mesh":"default"},"name":"mtp-allow-kuma-other-ns-and-tag","namespace":"kuma-system"},"spec":{"from":[{"default":{"action":"Allow"},"targetRef":{"kind":"MeshSubset","tags":{"asdfasdfasdf":"asdfasdfasdf","k8s.kuma.io/namespace":"kuma-other"}}}],"targetRef":{"kind":"Mesh"}}}
    kuma.io/display-name: mtp-allow-kuma-other-ns-and-tag
  creationTimestamp: "2024-12-11T08:44:24Z"
  generation: 1
  labels:
    k8s.kuma.io/namespace: kuma-system
    kuma.io/env: kubernetes
    kuma.io/mesh: default
    kuma.io/origin: zone
    kuma.io/policy-role: system
    kuma.io/zone: default
  name: mtp-allow-kuma-other-ns-and-tag
  namespace: kuma-system
  ownerReferences:
  - apiVersion: kuma.io/v1alpha1
    kind: Mesh
    name: default
    uid: ec9627ca-1846-4875-b31b-1ad13eb0838d
  resourceVersion: "6091"
  uid: f38541e7-78a6-4538-9ff9-837b5935b8f8
spec:
  from:
  - default:
      action: Allow
    targetRef:
      kind: MeshSubset
      tags:
        asdfasdfasdf: asdfasdfasdf
        k8s.kuma.io/namespace: kuma-other
  targetRef:
    kind: Mesh
`
	newReader := strings.NewReader(badMTPStr)
	all, err := io.ReadAll(newReader)
	if nil != err {
		panic(err)
	}

	var meshTP v1alpha12.MeshTrafficPermission
	err = yaml.Unmarshal(all, &meshTP)
	if nil != err {
		panic(err)
	}

	var result mtp_api.MeshTrafficPermissionResource
	result.Spec = meshTP.Spec
	result.SetMeta(&k8s.KubernetesMetaAdapter{
		ObjectMeta: meshTP.ObjectMeta,
		Mesh:       meshTP.GetMesh(),
	})

	return &result
}
