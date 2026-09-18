# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /cigate .

FROM --platform=$BUILDPLATFORM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/kumahq/kuma"
LABEL org.opencontainers.image.description="Decides what a Kuma CI run does, and whether it may be trusted"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /cigate /usr/local/bin/cigate

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["cigate"]
