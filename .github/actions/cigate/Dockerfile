# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /cigate .

FROM --platform=$BUILDPLATFORM alpine:latest@sha256:5b02b42e375f7426f8d65c3af331ca05d9878f9989230354504e0b9dfd431f60 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/kumahq/kuma"
LABEL org.opencontainers.image.description="Decides what a Kuma CI run does, and whether it may be trusted"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /cigate /usr/local/bin/cigate

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["cigate"]
