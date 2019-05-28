## build image
FROM golang:1.12.5
ENV GOPATH=
WORKDIR /go/src/github.com/Kong/konvoy/components/konvoy-control-plane
COPY . .

RUN make build

## runtime image
FROM alpine:latest

ENV PATH=$PATH:/konvoy-control-plane
COPY --from=0 /go/src/github.com/Kong/konvoy/components/konvoy-control-plane/build/artifacts/konvoy-control-plane/konvoy-control-plane /konvoy-control-plane/konvoy-control-plane

RUN addgroup -S -g 5678 konvoy-control-plane \
 && adduser -S -D -G konvoy-control-plane -u 5678 konvoy-control-plane

USER konvoy-control-plane

ENTRYPOINT ["konvoy-control-plane"]
CMD ["run"]
