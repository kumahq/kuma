GO_RUN := CGO_ENABLED=0 go run $(GOFLAGS) $(LD_FLAGS)

EXAMPLE_DATAPLANE_MESH ?= default
EXAMPLE_DATAPLANE_NAME ?= example
EXAMPLE_DATAPLANE_INBOUND_PORT ?= 8000
EXAMPLE_DATAPLANE_SERVICE_PORT ?= 10011
EXAMPLE_DATAPLANE_SERVICE_TAG ?= echo-service
ENVOY_ADMIN_PORT ?= 9901

define EXAMPLE_DATAPLANE_RESOURCE
type: Dataplane
mesh: $(EXAMPLE_DATAPLANE_MESH)
name: $(EXAMPLE_DATAPLANE_NAME)
networking:
  address: 127.0.0.1
  admin:
    port: $(ENVOY_ADMIN_PORT)
  inbound:
  - port: $(EXAMPLE_DATAPLANE_INBOUND_PORT)
    servicePort: $(EXAMPLE_DATAPLANE_SERVICE_PORT)
    tags:
      kuma.io/service: $(EXAMPLE_DATAPLANE_SERVICE_TAG)
      kuma.io/protocol: http
endef


POSTGRES_SSL_MODE ?= disable

NUM_OF_DATAPLANES ?= 100
NUM_OF_SERVICES ?= 80
KUMA_CP_ADDRESS ?= grpcs://localhost:5678

run/universal/postgres/ssl: ## Dev: Run Control Plane locally in universal mode with Postgres store and SSL enabled
	POSTGRES_SSL_MODE=verifyCa \
	POSTGRES_SSL_CERT_PATH=$(TOOLS_DIR)/postgres/certs/postgres.client.crt \
	POSTGRES_SSL_KEY_PATH=$(TOOLS_DIR)/postgres/certs/postgres.client.key \
	POSTGRES_SSL_ROOT_CERT_PATH=$(TOOLS_DIR)/postgres/certs/rootCA.crt \
	$(MAKE) run/universal/postgres

.PHONY: run/universal/postgres
run/universal/postgres: ## Dev: Run Control Plane locally in universal mode with Postgres store
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=postgres \
	KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma \
	KUMA_STORE_POSTGRES_TLS_MODE=$(POSTGRES_SSL_MODE) \
	KUMA_STORE_POSTGRES_TLS_CERT_PATH=$(POSTGRES_SSL_CERT_PATH) \
	KUMA_STORE_POSTGRES_TLS_KEY_PATH=$(POSTGRES_SSL_KEY_PATH) \
	KUMA_STORE_POSTGRES_TLS_CA_PATH=$(POSTGRES_SSL_ROOT_CERT_PATH) \
	$(GO_RUN) ./app/kuma-cp/main.go migrate up --log-level=debug

	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=postgres \
	KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma \
	KUMA_STORE_POSTGRES_TLS_MODE=$(POSTGRES_SSL_MODE) \
	KUMA_STORE_POSTGRES_TLS_CERT_PATH=$(POSTGRES_SSL_CERT_PATH) \
	KUMA_STORE_POSTGRES_TLS_KEY_PATH=$(POSTGRES_SSL_KEY_PATH) \
	KUMA_STORE_POSTGRES_TLS_CA_PATH=$(POSTGRES_SSL_ROOT_CERT_PATH) \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

.PHONY: run/example/envoy/universal
run/example/envoy/universal: run/echo-server run/example/envoy

.PHONY: run/example/envoy
run/example/envoy: export KUMA_DATAPLANE_RUNTIME_RESOURCE=$(EXAMPLE_DATAPLANE_RESOURCE)
run/example/envoy: build/kuma-dp build/kumactl ## Dev: Run Envoy configured against local Control Plane
	${BUILD_ARTIFACTS_DIR}/kumactl/kumactl generate dataplane-token --name=$(EXAMPLE_DATAPLANE_NAME) --mesh=$(EXAMPLE_DATAPLANE_MESH) > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token
	KUMA_DATAPLANE_RUNTIME_TOKEN_PATH=/tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token \
	${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp run --log-level=debug

.PHONY: config_dump/example/envoy
config_dump/example/envoy: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

.PHONY: run/universal/memory
run/universal/memory: ## Dev: Run Control Plane locally in universal mode with in-memory store
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=memory \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

.PHONY: start/postgres
start/postgres: ## Boostrap: start Postgres for Control Plane with initial schema
	docker-compose -f $(TOOLS_DIR)/postgres/docker-compose.yaml up -d --build
	$(TOOLS_DIR)/postgres/wait-for-postgres.sh 15432

.PHONY: stop/postgres
stop/postgres: ## Boostrap: stop Postgres
	docker-compose -f $(TOOLS_DIR)/postgres/docker-compose.yaml down

.PHONY: start/postgres/ssl
start/postgres/ssl: ## Boostrap: start Postgres for Control Plane with initial schema and SSL enabled
	POSTGRES_MODE=tls $(MAKE) start/postgres

.PHONY: stop/postgres/ssl
stop/postgres/ssl: ## Boostrap: stop Postgres with SSL enabled
	$(MAKE) stop/postgres

.PHONY: run/kuma-dp
run/kuma-dp: build/kumactl ## Dev: Run `kuma-dp` locally
	${BUILD_ARTIFACTS_DIR}/kumactl/kumactl generate dataplane-token --name=$(EXAMPLE_DATAPLANE_NAME) --mesh=$(EXAMPLE_DATAPLANE_MESH) > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_RUNTIME_TOKEN_PATH=/tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token \
	$(GO_RUN) ./app/kuma-dp/main.go run --log-level=debug

.PHONY: run/xds-client
run/xds-client:
	go run ./tools/xds-client/... run --dps "${NUM_OF_DATAPLANES}" --services "${NUM_OF_SERVICES}" --xds-server-address "${KUMA_CP_ADDRESS}"

.PHONY: run/echo-server
run/echo-server:
	go run test/server/main.go echo --port=$(EXAMPLE_DATAPLANE_SERVICE_PORT) &
