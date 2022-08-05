## kumactl install control-plane

Install Kuma Control Plane on Kubernetes

### Synopsis

Install Kuma Control Plane on Kubernetes in its own namespace.
This command requires that the KUBECONFIG environment is set

```
kumactl install control-plane [flags]
```

### Options

```
      --cni-bin-dir string                           set the CNI binary directory (default "/var/lib/cni/bin")
      --cni-chained                                  enable chained CNI installation
      --cni-conf-name string                         set the CNI configuration name (default "kuma-cni.conf")
      --cni-enabled                                  install Kuma with CNI instead of proxy init container
      --cni-net-dir string                           set the CNI install directory (default "/etc/cni/multus/net.d")
      --cni-node-selector stringToString             node selector for CNI deployment (default [])
      --cni-registry string                          registry for the image of the Kuma CNI component
      --cni-repository string                        repository for the image of the Kuma CNI component (default "install-cni")
      --cni-version string                           version of the image of the Kuma CNI component (default "0.0.10")
      --control-plane-node-selector stringToString   node selector for Kuma Control Plane (default [])
      --control-plane-registry string                registry for the image of the Kuma Control Plane component
      --control-plane-repository string              repository for the image of the Kuma Control Plane component (default "kuma-cp")
      --control-plane-service-name string            Service name of the Kuma Control Plane (default "kuma-control-plane")
      --control-plane-version string                 version of the image of the Kuma Control Plane component (default "unknown")
      --dataplane-init-registry string               registry for the init image of the Kuma DataPlane component
      --dataplane-init-repository string             repository for the init image of the Kuma DataPlane component (default "kuma-init")
      --dataplane-init-version string                version of the init image of the Kuma DataPlane component (default "unknown")
      --dataplane-registry string                    registry for the image of the Kuma DataPlane component
      --dataplane-repository string                  repository for the image of the Kuma DataPlane component (default "kuma-dp")
      --dataplane-version string                     version of the image of the Kuma DataPlane component (default "unknown")
      --dump-values                                  output all possible values for the configuration. This is similar to `helm show values <chart>
      --egress-drain-time string                     drain time for Envoy proxy (default "30s")
      --egress-enabled                               install Kuma with an Egress deployment, using the Data Plane image
      --egress-node-selector stringToString          node selector for Zone Egress (default [])
      --egress-service-type string                   the type for the Egress Service (ie. ClusterIP, NodePort, LoadBalancer) (default "ClusterIP")
      --env-var stringToString                       environment variables that will be passed to the control plane (default [])
      --experimental-gatewayapi                      install experimental Gateway API support
  -h, --help                                         help for control-plane
      --hooks-node-selector stringToString           node selector for Helm hooks (default [])
      --image-pull-policy string                     image pull policy that applies to all components of the Kuma Control Plane (default "IfNotPresent")
      --ingress-drain-time string                    drain time for Envoy proxy (default "30s")
      --ingress-enabled                              install Kuma with an Ingress deployment, using the Data Plane image
      --ingress-node-selector stringToString         node selector for Zone Ingress (default [])
      --ingress-use-node-port                        use NodePort instead of LoadBalancer for the Ingress Service
      --injector-failure-policy string               failure policy of the mutating web hook implemented by the Kuma Injector component (default "Fail")
      --kds-global-address string                    URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)
      --mode string                                  kuma cp modes: one of standalone|zone|global (default "standalone")
      --namespace string                             namespace to install Kuma Control Plane to (default "kuma-system")
      --registry string                              registry for all images (default "docker.io/kumahq")
      --set helm template <chart> --set ...          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2), This is similar to helm template <chart> --set ... to use set-file or set-string just use helm instead
      --tls-api-server-client-certs-secret string    Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS
      --tls-api-server-secret string                 Secret that contains tls.crt, tls.key for protecting Kuma API on HTTPS
      --tls-general-ca-bundle string                 Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)
      --tls-general-ca-secret string                 Secret that contains ca.crt that was used to sign cert for protecting Kuma in-cluster communication (ca.crt present in this secret have precedence over the one provided in --tls-general-secret)
      --tls-general-secret string                    Secret that contains tls.crt, tls.key [and ca.crt when no --tls-general-ca-secret specified] for protecting Kuma in-cluster communication
      --tls-kds-global-server-secret string          Secret that contains tls.crt, tls.key for protecting cross cluster communication
      --tls-kds-zone-client-secret string            Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification
      --use-node-port                                use NodePort instead of LoadBalancer
  -f, --values helm template <chart> -f ...          specify values in a YAML file or '-' for stdin. This is similar to helm template <chart> -f ...
      --version string                               version of Kuma Control Plane components
      --without-kubernetes-connection                install without connection to Kubernetes cluster. This can be used for initial Kuma installation, but not for upgrades
      --zone string                                  set the Kuma zone name
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl install](kumactl_install.md)	 - Install various Kuma components.

