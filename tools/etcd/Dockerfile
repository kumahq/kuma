FROM quay.io/coreos/etcd:latest

COPY etcd.yaml /etc/etcd.yaml

CMD ["/usr/local/bin/etcd","--config-file", "/etc/etcd.yaml"]