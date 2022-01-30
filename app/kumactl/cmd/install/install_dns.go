package install

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/app/kumactl/pkg/install/k8s"
)

const (
	corednsAppendTemplate = `mesh:53 {
        errors
        cache 3
        %s . %s
    }`
)

var resourceHeader = []byte("---\napiVersion: v1\nkind: ConfigMap\n")

type InstallDNSArgs struct {
	Namespace string
	Service   string
	Port      string
}

var DefaultInstallDNSArgs = InstallDNSArgs{
	Namespace: "kuma-system",
	Service:   "kuma-control-plane",
	Port:      "5653",
}

func newInstallDNS() *cobra.Command {
	args := DefaultInstallDNSArgs

	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Install DNS to Kubernetes",
		Long: `Install the DNS forwarding to the CoreDNS ConfigMap in the configured Kubernetes Cluster.
This command requires that the KUBECONFIG environment is set`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kubeClientConfig, err := k8s.DefaultClientConfig("", "")
			if err != nil {
				return errors.Wrap(err, "could not detect Kubernetes configuration")
			}

			clientset, err := kubernetes.NewForConfig(kubeClientConfig)
			if err != nil {
				return err
			}

			kumaCPSVC, err := clientset.CoreV1().Services(args.Namespace).Get(context.TODO(),
				args.Service, metav1.GetOptions{})
			if err != nil {
				return err
			}

			kumaDNSAddress := net.JoinHostPort(kumaCPSVC.Spec.ClusterIP, args.Port)

			var errs error
			generated := false
			corednsConfigMap, err := handleCoreDNS(clientset, kumaDNSAddress)
			if err != nil {
				errs = multierr.Append(errs, err)
			} else {
				err = outputYaml(cmd, corednsConfigMap)
				if err != nil {
					errs = multierr.Append(errs, err)
				}
				generated = true
			}

			kubednsConfigMap, err := handleKubeDNS(clientset, kumaDNSAddress)
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
	cmd.Flags().StringVar(&args.Namespace, "namespace", args.Namespace, "namespace to look for Kuma Control Plane service")
	cmd.Flags().StringVar(&args.Port, "port", args.Port, "port of the Kuma DNS server")

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

func handleCoreDNS(clientset *kubernetes.Clientset, kumaDNSAddress string) (*v1.ConfigMap, error) {
	corednsConfigMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(context.TODO(),
		"coredns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cleanupCoreConfigMap(corednsConfigMap)

	forwardVerb := getCoreFileForwardVerb(corednsConfigMap.Data["Corefile"])
	toappend := fmt.Sprintf(corednsAppendTemplate, forwardVerb, kumaDNSAddress)
	corednsConfigMap.Data["Corefile"] += toappend

	return corednsConfigMap, nil
}

func handleKubeDNS(clientset *kubernetes.Clientset, kumaDNSAddress string) (*v1.ConfigMap, error) {
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

	stubDomains["mesh"] = []string{kumaDNSAddress}

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
