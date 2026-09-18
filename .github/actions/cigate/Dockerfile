# syntax=docker/dockerfile:1@sha256:ecfaec9ed6d810b56388c508f4121597bfbba70d41a6dfeee4d8cad5f295fc32

FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /cigate .

FROM scratch

LABEL org.opencontainers.image.source="https://github.com/kumahq/kuma"
LABEL org.opencontainers.image.description="Decides what a Kuma CI run does, and whether it may be trusted"
LABEL org.opencontainers.image.licenses="Apache-2.0"

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /cigate /usr/local/bin/cigate

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["cigate"]
