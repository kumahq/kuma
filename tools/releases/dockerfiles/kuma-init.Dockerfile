FROM gcr.io/k8s-staging-build-image/distroless-iptables:v0.6.4
ARG ARCH

COPY /build/artifacts-linux-$ARCH/kumactl/kumactl /usr/bin

# this will be from a base image once it is done
COPY /tools/releases/templates/LICENSE \
    /tools/releases/templates/README \
    /kuma/

COPY /tools/releases/templates/NOTICE /kuma/NOTICE

# Manually add user equivalent to:
# adduser --system --disabled-password --group kumactl --uid 5678
RUN cat <<EOF >> /etc/passwd \
kumactl:x:5678:5678::/home/kumactl:/usr/sbin/nologin \
EOF

RUN cat <<EOF >> /etc/shadow \
kumactl:*:19000:0:99999:7::: \
EOF

RUN cat <<EOF >> /etc/group \
kumactl:x:5678: \
EOF

RUN cat <<EOF >> /etc/gshadow \
kumactl:!::: \
EOF

ENTRYPOINT ["/usr/bin/kumactl"]
CMD ["install", "transparent-proxy"]
