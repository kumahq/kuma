ARG BASE_IMAGE=alpine:latest
FROM ${BASE_IMAGE}
ARG ARCH
ARG KUBERNETES_RELEASE
RUN set -x \
 && wget -q -O /bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBERNETES_RELEASE}/bin/linux/${ARCH}/kubectl \
 && chmod +x /bin/kubectl

USER 65534

ENTRYPOINT ["kubectl"]
CMD ["--help"]
