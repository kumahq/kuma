package install

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	corednsAppendTemplate = `mesh:53 {
        errors
        cache 3
        %s . %s:5653
    }`
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

			clientset, err := kubernetes.NewForConfig(config)
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

func cleanupCoreConfigMap(configMap *v1.ConfigMap) {
	inConfigMap, found := configMap.Data["Corefile"]
	if !found {
		return
	}
	regex := regexp.MustCompile(`mesh:53 {[\s\S]*}`)
	configMap.Data["Corefile"] = string(regex.ReplaceAll([]byte(inConfigMap), []byte("")))
}

func handleCoreDNS(clientset *kubernetes.Clientset, cpaddress string) (*v1.ConfigMap, error) {
	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
		"coredns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cleanupCoreConfigMap(corednsConfigMap)

	forwardVerb := getCoreFileForwardVerb(corednsConfigMap.Data["Corefile"])
	toappend := fmt.Sprintf(corednsAppendTemplate, forwardVerb, cpaddress)
	corednsConfigMap.Data["Corefile"] += toappend

	return corednsConfigMap, nil
}

func handleKubeDNS(clientset *kubernetes.Clientset, cpaddress string) (*v1.ConfigMap, error) {
	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
		"kube-dns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	rawStubDomanins, found := corednsConfigMap.Data["stubDomains"]
	if !found || rawStubDomanins == "" {
		rawStubDomanins = "{}"
	}

	stubDomains := map[string][]string{}
	err = json.Unmarshal([]byte(rawStubDomanins), &stubDomains)
	if err != nil {
		return nil, err
	}

	if _, found := stubDomains["mesh"]; !found {
		stubDomains["mesh"] = []string{cpaddress}
	}

	json, err := json.Marshal(stubDomains)
	if err != nil {
		return nil, err
	}
	if corednsConfigMap.Data == nil {
		corednsConfigMap.Data = map[string]string{}
	}
	corednsConfigMap.Data["stubDomains"] = string(json)

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
