package install

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const appendTemplate = `mesh:53 {
        errors
        cache 30
        %s . %s:5653
    }`

var resourceHeader = []byte("apiVersion: v1\nkind: ConfigMap\n")

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

			corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
				"coredns", metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !strings.Contains(corednsConfigMap.Data["Corefile"], "mesh:53") {
				forwardVerb := getCoreFileForwardVerb(corednsConfigMap.Data["Corefile"])
				toappend := fmt.Sprintf(appendTemplate, forwardVerb, cpaddress)
				corednsConfigMap.Data["Corefile"] += toappend
			}

			corednsConfigMapYAML, err := yaml.Marshal(corednsConfigMap)
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
