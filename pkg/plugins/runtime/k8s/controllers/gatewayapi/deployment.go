package gatewayapi

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	kube_apps "k8s.io/api/apps/v1"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func k8sResourceName(name string) string {
	return fmt.Sprintf("%s-kuma-gateway", name)
}

func k8sSelector(name string) map[string]string {
	return map[string]string{"app": k8sResourceName(name)}
}

func (r *GatewayReconciler) createOrUpdateService(
	ctx context.Context,
	gateway *core_mesh.GatewayResource,
	k8sGateway *gatewayapi.Gateway,
) (*kube_core.Service, error) {
	service := &kube_core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k8sGateway.Namespace,
			Name:      k8sResourceName(k8sGateway.Name),
		},
	}

	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: k8sGateway.Namespace}, &ns); err != nil {
		return nil, errors.Wrap(err, "unable to get Namespace for gateway")
	}

	if _, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		var ports []kube_core.ServicePort

		for _, listener := range gateway.Spec.GetConf().GetListeners() {
			ports = append(ports, kube_core.ServicePort{
				Name:     strconv.Itoa(int(listener.Port)),
				Protocol: kube_core.ProtocolTCP,
				Port:     int32(listener.Port),
			})
		}

		service.Spec = kube_core.ServiceSpec{
			Selector: k8sSelector(k8sGateway.Name),
			Ports:    ports,
			Type:     kube_core.ServiceTypeLoadBalancer,
		}

		err := kube_controllerutil.SetControllerReference(k8sGateway, service, r.Scheme)
		return errors.Wrap(err, "unable to set Service's controller reference to Gateway")
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create or update Service for Gateway")
	}

	return service, nil
}

func (r *GatewayReconciler) createOrUpdateDeployment(
	ctx context.Context,
	gateway *core_mesh.GatewayResource,
	k8sGateway *gatewayapi.Gateway,
) (*kube_apps.Deployment, error) {
	ns := kube_core.Namespace{}
	if err := r.Client.Get(ctx, kube_types.NamespacedName{Name: k8sGateway.Namespace}, &ns); err != nil {
		return nil, errors.Wrap(err, "unable to get Namespace for gateway")
	}

	deployment := &kube_apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: k8sGateway.GetNamespace(),
			Name:      k8sResourceName(k8sGateway.GetName()),
		},
	}

	if _, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// TODO(michaelbeaumont) fix the resource limits ro fit use as a gateway
		// proxy
		container, err := r.ProxyFactory.NewContainer(k8sGateway.Annotations, &ns)
		if err != nil {
			return errors.Wrap(err, "unable to create Gateway container")
		}

		container.Name = util_k8s.KumaGatewayContainerName

		podSpec := kube_core.PodSpec{
			Containers: []kube_core.Container{container},
		}

		annotations := map[string]string{
			metadata.KumaGatewayAnnotation:          metadata.AnnotationBuiltin,
			metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationDisabled,
		}

		if mesh := util_k8s.MeshFor(k8sGateway); mesh != model.DefaultMesh {
			annotations[metadata.KumaMeshAnnotation] = mesh
		}

		var replicas int32 = 1

		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: k8sSelector(k8sGateway.Name),
		}
		deployment.Spec.Template = kube_core.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      k8sSelector(k8sGateway.Name),
				Annotations: annotations,
			},
			Spec: podSpec,
		}

		err = kube_controllerutil.SetControllerReference(k8sGateway, deployment, r.Scheme)
		return errors.Wrap(err, "unable to set Deployments's controller reference to Gateway")
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create or update Deployment for Gateway")
	}

	return deployment, nil
}
