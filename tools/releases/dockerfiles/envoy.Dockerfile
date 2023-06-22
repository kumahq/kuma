FROM debian:12 as envoy
ARG ARCH

COPY /build/artifacts-linux-$ARCH/envoy/envoy /envoy

RUN apt-get update && \
    apt-get install --no-install-recommends -y libcap2-bin=1:2.66-4 && \
    # extended permissions are stored into file inode and copied by docker when using buildx
    setcap cap_net_bind_service+ep /envoy \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
