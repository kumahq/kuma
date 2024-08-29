package containers

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	kube_api "k8s.io/apimachinery/pkg/api/resource"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type EnvVarsByName []kube_core.EnvVar

func (a EnvVarsByName) Len() int      { return len(a) }
func (a EnvVarsByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EnvVarsByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

type DataplaneProxyFactory struct {
	ControlPlaneURL           string
	ControlPlaneCACert        string
	DefaultAdminPort          uint32
	ContainerConfig           runtime_k8s.DataplaneContainer
	BuiltinDNS                runtime_k8s.BuiltinDNS
	WaitForDataplane          bool
	sidecarContainersEnabled  bool
	virtualProbesEnabled      bool
	applicationProbeProxyPort uint32
}

func NewDataplaneProxyFactory(
	controlPlaneURL string,
	controlPlaneCACert string,
	defaultAdminPort uint32,
	containerConfig runtime_k8s.DataplaneContainer,
	builtinDNS runtime_k8s.BuiltinDNS,
	waitForDataplane bool,
	sidecarContainersEnabled bool,
	virtualProbesEnabled bool,
	applicationProbeProxyPort uint32,
) *DataplaneProxyFactory {
	return &DataplaneProxyFactory{
		ControlPlaneURL:           controlPlaneURL,
		ControlPlaneCACert:        controlPlaneCACert,
		DefaultAdminPort:          defaultAdminPort,
		ContainerConfig:           containerConfig,
		BuiltinDNS:                builtinDNS,
		WaitForDataplane:          waitForDataplane,
		sidecarContainersEnabled:  sidecarContainersEnabled,
		virtualProbesEnabled:      virtualProbesEnabled,
		applicationProbeProxyPort: applicationProbeProxyPort,
	}
}

func (i *DataplaneProxyFactory) proxyConcurrencyFor(annotations map[string]string) (int64, error) {
	count, ok, err := metadata.Annotations(annotations).GetUint32(metadata.KumaSidecarConcurrencyAnnotation)
	if ok {
		return int64(count), err
	}

	// Note that validation requires the resource limit is not empty.
	cpuRequest := kube_api.MustParse(i.ContainerConfig.Resources.Limits.CPU)
	ncpu := cpuRequest.MilliValue() / 1000
	if ncpu < 2 {
		// Only autotune to down to 2 to mitigate the latency
		// risk if a worker thread blocks.
		ncpu = 2
	}

	return ncpu, nil
}

func (i *DataplaneProxyFactory) envoyAdminPort(annotations map[string]string) (uint32, error) {
	adminPort, _, err := metadata.Annotations(annotations).GetUint32(metadata.KumaEnvoyAdminPort)
	return adminPort, err
}

func (i *DataplaneProxyFactory) drainTime(annotations map[string]string) (time.Duration, error) {
	r, _, err := metadata.Annotations(annotations).GetDurationWithDefault(i.ContainerConfig.DrainTime.Duration, metadata.KumaSidecarDrainTime)
	return r, err
}

func (i *DataplaneProxyFactory) NewContainer(
	owner kube_client.Object,
	mesh string,
) (kube_core.Container, error) {
	annotations := owner.GetAnnotations()

	env, err := i.sidecarEnvVars(mesh, annotations)
	if err != nil {
		return kube_core.Container{}, err
	}

	cpuCount, err := i.proxyConcurrencyFor(annotations)
	if err != nil {
		return kube_core.Container{}, err
	}

	adminPort, err := i.envoyAdminPort(annotations)
	if err != nil {
		return kube_core.Container{}, err
	}
	if adminPort == 0 {
		adminPort = i.DefaultAdminPort
	}

	waitForDataplaneReady, _, err := metadata.Annotations(annotations).GetEnabledWithDefault(i.WaitForDataplane, metadata.KumaWaitForDataplaneReady)
	if err != nil {
		return kube_core.Container{}, err
	}

	args := []string{
		"run",
		"--log-level=info",
	}

	if cpuCount > 0 {
		args = append(args,
			"--concurrency="+strconv.FormatInt(cpuCount, 10))
	}

	container := kube_core.Container{
		Image:           i.ContainerConfig.Image,
		ImagePullPolicy: kube_core.PullIfNotPresent,
		Args:            args,
		Env:             env,
		SecurityContext: &kube_core.SecurityContext{
			RunAsUser:  &i.ContainerConfig.UID,
			RunAsGroup: &i.ContainerConfig.GID,
			Capabilities: &kube_core.Capabilities{
				Drop: []kube_core.Capability{
					"ALL",
				},
			},
		},
		LivenessProbe: &kube_core.Probe{
			ProbeHandler: kube_core.ProbeHandler{
				HTTPGet: &kube_core.HTTPGetAction{
					Path: "/ready",
					Port: kube_intstr.IntOrString{
						IntVal: int32(adminPort),
					},
				},
			},
			InitialDelaySeconds: i.ContainerConfig.LivenessProbe.InitialDelaySeconds,
			TimeoutSeconds:      i.ContainerConfig.LivenessProbe.TimeoutSeconds,
			PeriodSeconds:       i.ContainerConfig.LivenessProbe.PeriodSeconds,
			SuccessThreshold:    1,
			FailureThreshold:    i.ContainerConfig.LivenessProbe.FailureThreshold,
		},
		ReadinessProbe: &kube_core.Probe{
			ProbeHandler: kube_core.ProbeHandler{
				HTTPGet: &kube_core.HTTPGetAction{
					Path: "/ready",
					Port: kube_intstr.IntOrString{
						IntVal: int32(adminPort),
					},
				},
			},
			InitialDelaySeconds: i.ContainerConfig.ReadinessProbe.InitialDelaySeconds,
			TimeoutSeconds:      i.ContainerConfig.ReadinessProbe.TimeoutSeconds,
			PeriodSeconds:       i.ContainerConfig.ReadinessProbe.PeriodSeconds,
			SuccessThreshold:    i.ContainerConfig.ReadinessProbe.SuccessThreshold,
			FailureThreshold:    i.ContainerConfig.ReadinessProbe.FailureThreshold,
		},
		Resources: kube_core.ResourceRequirements{
			Requests: kube_core.ResourceList{
				kube_core.ResourceCPU:              kube_api.MustParse(i.ContainerConfig.Resources.Requests.CPU),
				kube_core.ResourceMemory:           kube_api.MustParse(i.ContainerConfig.Resources.Requests.Memory),
				kube_core.ResourceEphemeralStorage: pointer.Deref(kube_api.NewScaledQuantity(50, kube_api.Mega)),
			},
			Limits: kube_core.ResourceList{
				kube_core.ResourceCPU:              kube_api.MustParse(i.ContainerConfig.Resources.Limits.CPU),
				kube_core.ResourceMemory:           kube_api.MustParse(i.ContainerConfig.Resources.Limits.Memory),
				kube_core.ResourceEphemeralStorage: pointer.Deref(kube_api.NewScaledQuantity(1, kube_api.Giga)),
			},
		},
	}
	if i.sidecarContainersEnabled {
		container.StartupProbe = &kube_core.Probe{
			ProbeHandler: kube_core.ProbeHandler{
				HTTPGet: &kube_core.HTTPGetAction{
					Path: "/ready",
					Port: kube_intstr.IntOrString{
						IntVal: int32(adminPort),
					},
				},
			},
			InitialDelaySeconds: i.ContainerConfig.StartupProbe.InitialDelaySeconds,
			TimeoutSeconds:      i.ContainerConfig.StartupProbe.TimeoutSeconds,
			PeriodSeconds:       i.ContainerConfig.StartupProbe.PeriodSeconds,
			SuccessThreshold:    1,
			FailureThreshold:    i.ContainerConfig.StartupProbe.FailureThreshold,
		}
	}
	if waitForDataplaneReady {
		container.Lifecycle = &kube_core.Lifecycle{
			PostStart: &kube_core.LifecycleHandler{
				Exec: &kube_core.ExecAction{
					Command: []string{"kuma-dp", "wait", "--url", fmt.Sprintf("http://localhost:%d/ready", adminPort)},
				},
			},
		}
	}
	return container, nil
}

func (i *DataplaneProxyFactory) sidecarEnvVars(mesh string, podAnnotations map[string]string) ([]kube_core.EnvVar, error) {
	drainTime, err := i.drainTime(podAnnotations)
	if err != nil {
		return nil, err
	}

	envVars := map[string]kube_core.EnvVar{
		"POD_NAME": {
			Name: "POD_NAME",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		"POD_NAMESPACE": {
			Name: "POD_NAMESPACE",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		"INSTANCE_IP": {
			Name: "INSTANCE_IP",
			ValueFrom: &kube_core.EnvVarSource{
				FieldRef: &kube_core.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		},
		"KUMA_CONTROL_PLANE_URL": {
			Name:  "KUMA_CONTROL_PLANE_URL",
			Value: i.ControlPlaneURL,
		},
		"KUMA_DATAPLANE_MESH": {
			Name:  "KUMA_DATAPLANE_MESH",
			Value: mesh,
		},
		"KUMA_DATAPLANE_DRAIN_TIME": {
			Name:  "KUMA_DATAPLANE_DRAIN_TIME",
			Value: drainTime.String(),
		},
		"KUMA_DATAPLANE_RUNTIME_TOKEN_PATH": {
			Name:  "KUMA_DATAPLANE_RUNTIME_TOKEN_PATH",
			Value: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		},
		"KUMA_CONTROL_PLANE_CA_CERT": {
			Name:  "KUMA_CONTROL_PLANE_CA_CERT",
			Value: i.ControlPlaneCACert,
		},
	}
	if i.BuiltinDNS.Enabled {
		envVars["KUMA_DNS_ENABLED"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENABLED",
			Value: "true",
		}

		envVars["KUMA_DNS_ENABLE_LOGGING"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENABLE_LOGGING",
			Value: strconv.FormatBool(i.BuiltinDNS.Logging),
		}

		envVars["KUMA_DNS_CORE_DNS_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_CORE_DNS_PORT",
			Value: strconv.FormatInt(int64(i.BuiltinDNS.Port), 10),
		}

		envVars["KUMA_DNS_ENVOY_DNS_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENVOY_DNS_PORT",
			Value: strconv.FormatInt(int64(i.BuiltinDNS.Port+2), 10),
		}

		envVars["KUMA_DNS_CORE_DNS_BINARY_PATH"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_CORE_DNS_BINARY_PATH",
			Value: "coredns",
		}
	} else {
		envVars["KUMA_DNS_ENABLED"] = kube_core.EnvVar{
			Name:  "KUMA_DNS_ENABLED",
			Value: "false",
		}
	}

	if logLevel, exist := metadata.Annotations(podAnnotations).GetString(metadata.KumaEnvoyLogLevel); exist {
		envVars["KUMA_DATAPLANE_RUNTIME_ENVOY_LOG_LEVEL"] = kube_core.EnvVar{
			Name:  "KUMA_DATAPLANE_RUNTIME_ENVOY_LOG_LEVEL",
			Value: logLevel,
		}
	}
	if complogLevel, exist := metadata.Annotations(podAnnotations).GetString(metadata.KumaEnvoyComponentLogLevel); exist {
		envVars["KUMA_DATAPLANE_RUNTIME_ENVOY_COMPONENT_LOG_LEVEL"] = kube_core.EnvVar{
			Name:  "KUMA_DATAPLANE_RUNTIME_ENVOY_COMPONENT_LOG_LEVEL",
			Value: complogLevel,
		}
	}

	annotations := make(map[string]string)
	if err := probes.SetVirtualProbesEnabledAnnotation(annotations, podAnnotations, i.virtualProbesEnabled); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaVirtualProbesAnnotation))
	}
	if err := probes.SetApplicationProbeProxyPortAnnotation(annotations, podAnnotations, i.applicationProbeProxyPort); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to set %s", metadata.KumaApplicationProbeProxyPortAnnotation))
	}
	if appProbeProxyPort, _, _ := metadata.Annotations(annotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation); appProbeProxyPort > 0 {
		envVars["KUMA_APPLICATION_PROBE_PROXY_PORT"] = kube_core.EnvVar{
			Name:  "KUMA_APPLICATION_PROBE_PROXY_PORT",
			Value: strconv.Itoa(int(appProbeProxyPort)),
		}
	}

	// override defaults with cfg env vars
	for envName, envVal := range i.ContainerConfig.EnvVars {
		envVars[envName] = kube_core.EnvVar{
			Name:  envName,
			Value: envVal,
		}
	}

	// override defaults and cfg env vars with annotations
	annotationEnvVars, _, err := metadata.Annotations(podAnnotations).GetMap(metadata.KumaSidecarEnvVarsAnnotation)
	if err != nil {
		return nil, err
	}
	for envName, envVal := range annotationEnvVars {
		envVars[envName] = kube_core.EnvVar{
			Name:  envName,
			Value: envVal,
		}
	}

	var result []kube_core.EnvVar
	for _, v := range envVars {
		result = append(result, v)
	}
	sort.Stable(EnvVarsByName(result))

	return result, nil
}
