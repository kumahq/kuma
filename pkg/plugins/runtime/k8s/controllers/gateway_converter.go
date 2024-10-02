package controllers

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

// createorUpdateBuiltinGatewayDataplane manages the dataplane for a pod
// belonging to a built-in Kuma gateway.
func (r *PodReconciler) createorUpdateBuiltinGatewayDataplane(ctx context.Context, pod *kube_core.Pod, ns *kube_core.Namespace) error {
	dataplane := &mesh_k8s.Dataplane{
		ObjectMeta: kube_meta.ObjectMeta{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		},
		Mesh: k8s_util.MeshOfByAnnotation(pod, ns),
	}

	tagsAnnotation, ok := pod.Annotations[metadata.KumaTagsAnnotation]

	if !ok {
		r.EventRecorder.Eventf(
			pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Missing %s annotation on Pod serving a builtin Gateway", metadata.KumaTagsAnnotation,
		)
		return nil
	}

	var tags map[string]string
	if err := json.Unmarshal([]byte(tagsAnnotation), &tags); err != nil || tags == nil {
		r.EventRecorder.Eventf(
			pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Invalid %s annotation on Pod serving a builtin Gateway", metadata.KumaTagsAnnotation,
		)
		return nil
	}

	dataplaneProto, err := r.PodConverter.BuiltinGatewayDataplane(pod, tags)
	if err != nil {
		return errors.Wrap(err, "unable to translate a Gateway Pod into a Dataplane")
	} else if dataplaneProto == nil {
		// we don't want a dataplane, the existing object will be deleted
		// through owner refs
		return nil
	}

	log := r.Log.WithValues("pod", kube_types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	operationResult, err := kube_controllerutil.CreateOrUpdate(ctx, r.Client, dataplane, func() error {
		currentSpec, err := dataplane.GetSpec()
		if err != nil {
			return errors.Wrap(err, "unable to get current Dataplane's spec")
		}
		if model.Equal(currentSpec, dataplaneProto) {
			log.V(1).Info("resource hasn't changed, skip")
			return nil
		}

		dataplane.SetSpec(dataplaneProto)

		if err := kube_controllerutil.SetControllerReference(pod, dataplane, r.Scheme); err != nil {
			return errors.Wrap(err, "unable to set Dataplane's controller reference to Pod")
		}
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create/update Dataplane", "operationResult", operationResult)
		r.EventRecorder.Eventf(pod, kube_core.EventTypeWarning, FailedToGenerateKumaDataplaneReason, "Failed to generate Kuma Dataplane: %s", err.Error())
		return err
	}

	switch operationResult {
	case kube_controllerutil.OperationResultCreated:
		log.Info("Dataplane created")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, CreatedKumaDataplaneReason, "Created Kuma Dataplane: %s", dataplane.Name)
	case kube_controllerutil.OperationResultUpdated:
		log.Info("Dataplane updated")
		r.EventRecorder.Eventf(pod, kube_core.EventTypeNormal, UpdatedKumaDataplaneReason, "Updated Kuma Dataplane: %s", dataplane.Name)
	}
	return nil
}

func (p *PodConverter) BuiltinGatewayDataplane(
	pod *kube_core.Pod,
	tags map[string]string,
) (*mesh_proto.Dataplane, error) {
	if p.Zone != "" {
		tags[mesh_proto.ZoneTag] = p.Zone
	}
	dataplaneProto := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: pod.Status.PodIP,
			Gateway: &mesh_proto.Dataplane_Networking_Gateway{
				Tags: tags,
				Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
			},
		},
	}

	return &dataplaneProto, nil
}
