package types

import (
	"fmt"
	"path"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	k8s_apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	crdgen_version "sigs.k8s.io/controller-tools/pkg/version"
	gateway_api "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/data"
)

// Generate generates YAML file data for all the custom resources in
// supported release of the Kubernetes Gateway API.
func Generate() (data.FileList, error) {
	var files []data.File

	crds, err := generateGatewayResources()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate Gateway API resources")
	}

	for _, c := range crds {
		if c.Spec.Group != gateway_api.GroupName {
			continue
		}

		bytes, err := yaml.Marshal(c)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render %s resource", c.Name)
		}

		name := fmt.Sprintf("%s_%s.yaml", gateway_api.GroupName, c.Spec.Names.Plural)
		files = append(files, data.File{
			Name: name,
			Data: bytes,
			// This needs to be a path below "deployment/charts/kuma".
			FullPath: path.Join("crds", "gateway-api", gateway_api.GroupVersion.Version, name),
		})
	}

	return files, nil
}

func generateGatewayResources() ([]*k8s_apixv1.CustomResourceDefinition, error) {
	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/gateway-api/apis/v1alpha2",
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load packages")
	}

	generator := &crd.Generator{}
	parser := &crd.Parser{
		Collector: &markers.Collector{Registry: &markers.Registry{}},
		Checker: &loader.TypeChecker{
			NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
		},
	}

	if err := generator.RegisterMarkers(parser.Collector.Registry); err != nil {
		return nil, err
	}

	for _, r := range roots {
		parser.NeedPackage(r)
	}

	metav1Pkg := crd.FindMetav1(roots)
	if metav1Pkg == nil {
		return nil, errors.New("missing k8s.io/apimachinery/pkg/apis/meta/v1")
	}

	kubeKinds := crd.FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		return nil, errors.New("no Kubernetes types")
	}

	for pkg := range parser.GroupVersions {
		for _, err := range pkg.Errors {
			panic(fmt.Sprintf("%s: %s\n", pkg.PkgPath, err))
		}
	}

	var crds []*k8s_apixv1.CustomResourceDefinition

	for groupKind := range kubeKinds {
		parser.NeedCRDFor(groupKind, nil)
		crdRaw := parser.CustomResourceDefinitions[groupKind]

		if crdRaw.ObjectMeta.Annotations == nil {
			crdRaw.ObjectMeta.Annotations = map[string]string{}
		}

		// Upstream Gateway API does some terrible sed postprocessing to add the approval
		// annotation. We reproduce it here because is it needed for the reserved group name.
		crdRaw.ObjectMeta.Annotations["controller-gen.kubebuilder.io/version"] = crdgen_version.Version()
		crdRaw.ObjectMeta.Annotations[k8s_apixv1.KubeAPIApprovedAnnotation] = "https://github.com/kubernetes-sigs/gateway-api/pull/891"

		crd.FixTopLevelMetadata(crdRaw)

		conv, err := crd.AsVersion(crdRaw, k8s_apixv1.SchemeGroupVersion)
		if err != nil {
			return nil, err
		}

		crds = append(crds, conv.(*k8s_apixv1.CustomResourceDefinition))
	}

	return crds, nil
}
