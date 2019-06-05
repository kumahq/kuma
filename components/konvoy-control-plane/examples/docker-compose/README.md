Konvoy Control Plane inside Docker Compose 
====================

## Usage

Build and run the containers:

```bash
docker-compose pull && \
docker-compose up --build --no-start && \
docker-compose up
```

Make test requests:

```bash
docker run --network docker-compose_envoymesh --rm -ti tutum/curl sh -c 'while true ; do curl http://demo-app:8080 && sleep 1 ; done'
```

Observe Envoy stats:

```bash
watch 'docker exec docker-compose_envoy_1 curl -s localhost:9901/stats | grep upstream_rq_total'
```

E.g.,
```
cluster.ads_cluster.upstream_rq_total: 1
cluster.localhost_8080.upstream_rq_total: 7
cluster.pass_through.upstream_rq_total: 7
```

where

* `cluster.localhost_8080.upstream_rq_total` is a number of `inbound` requests
* `cluster.pass_through.upstream_rq_total` is a number of `outbound` requests
