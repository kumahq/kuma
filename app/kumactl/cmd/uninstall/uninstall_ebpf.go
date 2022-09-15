package uninstall

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/client/k8s"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	kuma_version "github.com/kumahq/kuma/pkg/version"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	typedbatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	AppLabel          = "kuma.io/bpf-cleanup"
	BpfCleanupJobName = "kuma-bpf-cleanup"
	BpfCleanupImage   = "kumahq/kuma-init"
)

var (
	DeleteImmediately         = metav1.DeleteOptions{GracePeriodSeconds: new(int64)}
	KumaBpfLabelSelector      = fmt.Sprintf("%s=%s", AppLabel, BpfCleanupJobName)
	KumaBpfCleanupAppSelector = metav1.ListOptions{LabelSelector: KumaBpfLabelSelector}
)

type JobResource struct {
	jobClient typedbatchv1.JobInterface
	podClient typedcorev1.PodInterface
}

type ebpfArgs struct {
	BPFFsPath           string
	Timeout             time.Duration
	CleanupImageVersion string
	RemoveOnly          bool
	Namespace           string
}

func newUninstallEbpf(root *kumactl_cmd.RootContext) *cobra.Command {
	args := ebpfArgs{
		// default value that we inject in pod injector
		BPFFsPath:           root.InstallCpContext.Args.Ebpf_bpffspath,
		Timeout:             time.Duration(120 * time.Second),
		CleanupImageVersion: kuma_version.Build.Version,
		RemoveOnly:          false,
		Namespace:           root.InstallCpContext.Args.Namespace,
	}
	cmd := &cobra.Command{
		Use:   "ebpf",
		Short: "Uninstall BPF files from the nodes",
		Long:  `Uninstall BPF files from the nodes by removing BPF programs from all the nodes`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kubeClientConfig, err := k8s.DefaultClientConfig("", "")
			if err != nil {
				return errors.Wrap(err, "Could not detect Kubernetes configuration")
			}

			k8sClient, err := kubernetes.NewForConfig(kubeClientConfig)
			if err != nil {
				return errors.Wrap(err, "Could not create Kubernetes client")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), args.Timeout)
			defer cancel()

			nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "Failed obtaining nodes from Kubernetes cluster")
			}
			jobResource := JobResource{
				jobClient: k8sClient.BatchV1().Jobs(args.Namespace),
				podClient: k8sClient.CoreV1().Pods(args.Namespace),
			}

			if args.RemoveOnly {
				if err := jobResource.Cleanup(ctx); err != nil {
					return errors.Wrap(err, "Failed cleaning jobs")
				}
				return nil
			}

			for id, node := range nodes.Items {
				_, err := jobResource.jobClient.Create(ctx, getJobSpec(id, node.Name, args.BPFFsPath, args.CleanupImageVersion), metav1.CreateOptions{})
				if err != nil {
					return errors.Wrap(err, "Failed creating jobs")
				}
			}
			watcher, err := jobResource.podClient.Watch(ctx, metav1.ListOptions{
				LabelSelector: KumaBpfLabelSelector,
				Watch:         true,
			})
			if err != nil {
				return errors.Wrap(err, "failed to create pod watcher")
			}

			defer func() {
				if e := jobResource.Cleanup(context.Background()); e != nil {
					err = e
				}
			}()

			errCh := make(chan error, 1)
			go func() {
				errCh <- jobResource.Watch(ctx, watcher)
			}()

			select {
			case <-ctx.Done():
				return nil
			case err := <-errCh:
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace where job is created")
	cmd.Flags().StringVar(&args.BPFFsPath, "bpf-fs-path", args.BPFFsPath, "path where bpf programs were installed")
	cmd.Flags().DurationVar(&args.Timeout, "timeout", args.Timeout, "timeout for whole process of removing left files")
	cmd.Flags().StringVar(&args.CleanupImageVersion, "cleanup-image-version", args.CleanupImageVersion, "version of cleanup ebpf job image")
	cmd.Flags().BoolVar(&args.RemoveOnly, "remove-only", args.RemoveOnly, "cleanup jobs and pods only")
	return cmd
}

func getJobSpec(id int, nodeName, mountPath, version string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", BpfCleanupJobName, id),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						AppLabel: BpfCleanupJobName,
					},
				},
				Spec: corev1.PodSpec{
					NodeName: nodeName,
					Containers: []corev1.Container{
						{
							Name:    BpfCleanupJobName,
							Image:   fmt.Sprintf("%s:%s", BpfCleanupImage, version),
							Command: strings.Split("sleep 15", " "),
							SecurityContext: &corev1.SecurityContext{
								Privileged: new(bool),
							},
							VolumeMounts: []corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "bpf-fs-path",
									MountPath: mountPath,
								},
								corev1.VolumeMount{
									Name:      "sys-fs-cgroup",
									MountPath: "/sys/fs/cgroup",
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "bpf-fs-path",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: mountPath,
								},
							},
						},
						corev1.Volume{
							Name: "sys-fs-cgroup",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys/fs/cgroup",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (j *JobResource) Watch(ctx context.Context, watcher watch.Interface) (e error) {
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var phase corev1.PodPhase
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				return nil
			}
			if pod.Status.Phase == phase {
				continue
			}
			switch pod.Status.Phase {
			case corev1.PodSucceeded:
				return nil
			case corev1.PodFailed:
				return fmt.Errorf("pod failed")
			}
			phase = pod.Status.Phase
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (jr *JobResource) Cleanup(ctx context.Context) error {
	err := jr.jobClient.DeleteCollection(ctx, DeleteImmediately, KumaBpfCleanupAppSelector)
	if err != nil {
		return fmt.Errorf("failed to delete jobs %s", err)
	}
	err = jr.podClient.DeleteCollection(ctx, DeleteImmediately, KumaBpfCleanupAppSelector)
	if err != nil {
		return fmt.Errorf("failed to delete pods %s", err)
	}
	return nil
}
