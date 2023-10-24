## kumactl install transparent-proxy

Install Transparent Proxy pre-requisites on the host

### Synopsis

Install Transparent Proxy by modifying the hosts iptables.

Follow the following steps to use the Kuma data plane proxy in Transparent Proxy mode:

 1) create a dedicated user for the Kuma data plane proxy, e.g. 'kuma-dp'
 2) run this command as a 'root' user to modify the host's iptables and /etc/resolv.conf
    - supply the dedicated username with '--kuma-dp-uid'
    - all changes are easly revertible by issuing 'kumactl uninstall transparent-proxy'
    - by default the SSH port tcp/22 will not be redirected to Envoy, but everything else will.
      Use '--exclude-inbound-ports' to provide a comma separated list of ports that should also be excluded

 sudo kumactl install transparent-proxy \
          --kuma-dp-user kuma-dp \
          --exclude-inbound-ports 443

 3) prepare a Dataplane resource yaml like this:

type: Dataplane
mesh: default
name: {{ name }}
networking:
  address: {{ address }}
  inbound:
  - port: {{ port }}
    tags:
      kuma.io/service: demo-client
  transparentProxying:
    redirectPortInbound: 15006
    redirectPortOutbound: 15001

The values in 'transparentProxying' section are the defaults set by this command and if needed be changed by supplying 
'--redirect-inbound-port' and '--redirect-outbound-port' respectively.

 4) the kuma-dp command shall be run with the designated user. 
    - if using systemd to run add 'User=kuma-dp' in the '[Service]' section of the service file
    - leverage 'runuser' similar to (assuming aforementioned yaml):

runuser -u kuma-dp -- \
  /usr/bin/kuma-dp run \
    --cp-address=https://172.19.0.2:5678 \
    --dataplane-token-file=/kuma/token-demo \
    --dataplane-file=/kuma/dpyaml-demo \
    --dataplane-var name=dp-demo \
    --dataplane-var address=172.19.0.4 \
    --dataplane-var port=80  \
    --binary-path /usr/local/bin/envoy



```
kumactl install transparent-proxy [flags]
```

### Options

```
      --dry-run                                                                         dry run
      --ebpf-bpffs-path string                                                          the path of the BPF filesystem (default "/sys/fs/bpf")
      --ebpf-cgroup-path string                                                         the path of cgroup2 (default "/sys/fs/cgroup")
      --ebpf-enabled                                                                    use ebpf instead of iptables to install transparent proxy
      --ebpf-instance-ip string                                                         IP address of the instance (pod/vm) where transparent proxy will be installed
      --ebpf-programs-source-path string                                                path where compiled ebpf programs and other necessary for ebpf mode files can be found (default "/kuma/ebpf")
      --ebpf-tc-attach-iface string                                                     name of the interface which TC eBPF programs should be attached to
      --exclude-inbound-ports string                                                    a comma separated list of inbound ports to exclude from redirect to Envoy
      --exclude-outbound-ports string                                                   a comma separated list of outbound ports to exclude from redirect to Envoy
      --exclude-outbound-ports-for-uids stringArray                                     outbound ports to exclude for specific uids in a format of protocol:ports:uids where protocol and ports can be omitted or have value tcp or udp and ports can be a single value, a list, a range or a combination of all or * and uid can be a value or a range e.g. 53,3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for both TCP and UDP for uids 106, 107, 108
      --exclude-outbound-tcp-ports-for-uids stringArray                                 [DEPRECATED (use --exclude-outbound-ports-for-uids)] tcp outbound ports to exclude for specific uids in a format of ports:uids where ports can be a single value, a list, a range or a combination of all and uid can be a value or a range e.g. 53,3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for uids 106, 107, 108
      --exclude-outbound-udp-ports-for-uids stringArray                                 [DEPRECATED (use --exclude-outbound-ports-for-uids)] udp outbound ports to exclude for specific uids in a format of ports:uids where ports can be a single value, a list, a range or a combination of all and uid can be a value or a range e.g. 53, 3000-5000:106-108 would mean exclude ports 53 and from 3000 to 5000 for uids 106, 107, 108
  -h, --help                                                                            help for transparent-proxy
      --kuma-dp-uid string                                                              the uid of the user that will run kuma-dp
      --kuma-dp-user string                                                             the user that will run kuma-dp
      --max-retries int                                                                 flag can be used to specify the maximum number of times to retry an installation before giving up (default 5)
      --redirect-all-dns-traffic                                                        redirect all DNS traffic to a specified port, unlike --redirect-dns this will not be limited to the dns servers identified in /etc/resolve.conf
      --redirect-dns                                                                    redirect only DNS requests targeted to the servers listed in /etc/resolv.conf to a specified port
      --redirect-dns-port string                                                        the port where the DNS agent is listening (default "15053")
      --redirect-dns-upstream-target-chain string                                       (optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care. (default "RETURN")
      --redirect-inbound                                                                redirect the inbound traffic to the Envoy. Should be disabled for Gateway data plane proxies. (default true)
      --redirect-inbound-port networking.transparentProxying.redirectPortInbound        inbound port redirected to Envoy, as specified in dataplane's networking.transparentProxying.redirectPortInbound (default "15006")
      --redirect-inbound-port-v6 networking.transparentProxying.redirectPortInboundV6   IPv6 inbound port redirected to Envoy, as specified in dataplane's networking.transparentProxying.redirectPortInboundV6 (default "15010")
      --redirect-outbound-port networking.transparentProxying.redirectPortOutbound      outbound port redirected to Envoy, as specified in dataplane's networking.transparentProxying.redirectPortOutbound (default "15001")
      --skip-dns-conntrack-zone-split                                                   skip applying conntrack zone splitting iptables rules
      --sleep-between-retries duration                                                  flag can be used to specify the amount of time to sleep between retries (default 2s)
      --store-firewalld                                                                 store the iptables changes with firewalld
      --verbose                                                                         verbose
      --vnet stringArray                                                                virtual networks in a format of interfaceNameRegex:CIDR split by ':' where interface name doesn't have to be exact name e.g. docker0:172.17.0.0/16, br+:172.18.0.0/16, iface:::1/64
      --wait uint                                                                       specify the amount of time, in seconds, that the application should wait for the xtables exclusive lock before exiting. If the lock is not available within the specified time, the application will exit with an error (default 5)
      --wait-interval uint                                                              flag can be used to specify the amount of time, in microseconds, that iptables should wait between each iteration of the lock acquisition loop. This can be useful if the xtables lock is being held by another application for a long time, and you want to reduce the amount of CPU that iptables uses while waiting for the lock
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

