package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	corednsAppendTemplate = `mesh:53 {
        errors
        cache 30
        %s . %s:5653
    }`
	kubednsAppendTemplate = `{"mesh": %s}\n`
)

var resourceHeader = []byte("---\napiVersion: v1\nkind: ConfigMap\n")

func newInstallDNS() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Install DNS to Kubernetes",
		Long: `Install the DNS forwarding to the CoreDNS ConfigMap in the configured Kubernetes Cluster.
This command requires that the KUBECONFIG environment is set`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kubeconfigPath := os.Getenv("KUBECONFIG")
			if kubeconfigPath == "" {
				return errors.Errorf("Please set KUBECONFIG before running the commands")
			}

			config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				return err
			}

			clientset, _ := kubernetes.NewForConfig(config)
			if err != nil {
				return err
			}

			kumaCPSVC, err := clientset.CoreV1().Services("kuma-system").Get(context.TODO(),
				"kuma-control-plane", metav1.GetOptions{})
			if err != nil {
				return err
			}

			cpaddress := kumaCPSVC.Spec.ClusterIP

			var errs error
			generated := false
			corednsConfigMap, err := handleCoreDNS(clientset, cpaddress)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				err = outputYaml(cmd, corednsConfigMap)
				if err != nil {
					errs = multierr.Append(errs, err)
				}
				generated = true
			}

			kubednsConfigMap, err := handleKubeDNS(clientset, cpaddress)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				err = outputYaml(cmd, kubednsConfigMap)
				if err != nil {
					errs = multierr.Append(errs, err)
				}
				generated = true
			}

			if !generated {
				return errs
			}

			return nil
		},
	}

	return cmd
}

// CoreDNS before 1.4 (K8s < 1.16) uses the verb `proxy` to denote the forwarding
// All later versions user `forward`. Detect what is used in the ConfigMap's Corefile
func getCoreFileForwardVerb(corefile string) string {
	if strings.Contains(corefile, "proxy . /etc/resolv.conf") {
		return "proxy"
	}

	return "forward"
}

func handleCoreDNS(clientset *kubernetes.Clientset, cpaddress string) (*v1.ConfigMap, error) {
	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
		"coredns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if !strings.Contains(corednsConfigMap.Data["Corefile"], "mesh:53") {
		forwardVerb := getCoreFileForwardVerb(corednsConfigMap.Data["Corefile"])
		toappend := fmt.Sprintf(corednsAppendTemplate, forwardVerb, cpaddress)
		corednsConfigMap.Data["Corefile"] += toappend
	}
	return corednsConfigMap, nil
}

func handleKubeDNS(clientset *kubernetes.Clientset, cpaddress string) (*v1.ConfigMap, error) {
	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
		"kube-dns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if !strings.Contains(corednsConfigMap.Data["stubDomains"], "\"mesh\"") {
		toappend := fmt.Sprintf(kubednsAppendTemplate, cpaddress)
		corednsConfigMap.Data["stubDomains"] += toappend
	}
	return corednsConfigMap, nil
}

func outputYaml(cmd *cobra.Command, configMap *v1.ConfigMap) error {
	corednsConfigMapYAML, err := yaml.Marshal(configMap)
	if err != nil {
		return err
	}

	if _, err := cmd.OutOrStdout().Write(resourceHeader); err != nil {
		return errors.Wrap(err, "Failed to output the generated resources")
	}

	if _, err := cmd.OutOrStdout().Write(corednsConfigMapYAML); err != nil {
		return errors.Wrap(err, "Failed to output the generated resources")
	}
	return nil
}
