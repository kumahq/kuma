EXAMPLE_DATAPLANE_MESH ?= default
EXAMPLE_DATAPLANE_NAME ?= example
EXAMPLE_DATAPLANE_INBOUND_PORT ?= 8000
EXAMPLE_DATAPLANE_SERVICE_PORT ?= 10011
EXAMPLE_DATAPLANE_SERVICE_TAG ?= echo-service
ENVOY_ADMIN_PORT = $(shell expr $(EXAMPLE_DATAPLANE_INBOUND_PORT) - 8000 + 9901)

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
    serviceProbe:
       tcp: {}
    tags:
      kuma.io/service: $(EXAMPLE_DATAPLANE_SERVICE_TAG)
      kuma.io/protocol: http
endef

POSTGRES_MODE = standard
CP_STORE = memory
CP_ENV += KUMA_ENVIRONMENT=universal KUMA_MULTIZONE_ZONE_NAME=zone-1 KUMA_STORE_TYPE=$(CP_STORE)
ifeq ($(CP_STORE),postgres)
CP_ENV += KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma
endif
ifeq ($(POSTGRES_MODE),tls)
CP_ENV += KUMA_STORE_POSTGRES_TLS_MODE=verifyCa \
	KUMA_STORE_POSTGRES_TLS_CERT_PATH=$(TOOLS_DIR)/postgres/certs/postgres.client.crt \
	KUMA_STORE_POSTGRES_TLS_KEY_PATH=$(TOOLS_DIR)/postgres/certs/postgres.client.key \
	KUMA_STORE_POSTGRES_TLS_CA_PATH=$(TOOLS_DIR)/postgres/certs/rootCA.crt
endif
CP_ENV += $(EXTRA_CP_ENV)

export EXAMPLE_DATAPLANE_RESOURCE

NUM_OF_DATAPLANES ?= 100
NUM_OF_SERVICES ?= 80
KUMA_CP_ADDRESS ?= grpcs://localhost:5678

.PHONY: run/postgres/start
run/postgres/start: ## Dev: start Postgres for Control Plane with initial schema (Use POSTGRES_MODE=tls to enable TLS)
	$(CP_ENV) POSTGRES_MODE=$(POSTGRES_MODE) docker-compose -f $(TOOLS_DIR)/postgres/docker-compose.yaml up -d --build

.PHONY: run/postgres/stop
run/postgres/stop: ## Dev: stop Postgres
	docker-compose -f $(TOOLS_DIR)/postgres/docker-compose.yaml down

.PHONY: run/kuma-cp
run/kuma-cp: $(DISTRIBUTION_FOLDER) ## Dev: Run `kuma-cp` locally (use CP_STORE=postgres to use postgres as a store and POSTGRES_MODE=tls to enabled TLS, use EXTRA_CP_ENV to add extra Kuma env vars)
ifeq ($(CP_STORE),postgres)
	$(CP_ENV) $(DISTRIBUTION_FOLDER)/bin/kuma-cp migrate up --log-level=debug
endif
	$(CP_ENV) $(DISTRIBUTION_FOLDER)/bin/kuma-cp run --log-level=debug

.PHONY: run/kuma-dp
run/kuma-dp: $(DISTRIBUTION_FOLDER) ## Dev: Run `kuma-dp` locally
	$(DISTRIBUTION_FOLDER)/bin/kumactl generate dataplane-token --name=$(EXAMPLE_DATAPLANE_NAME) --mesh=$(EXAMPLE_DATAPLANE_MESH) --valid-for=24h > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token
	echo "$$EXAMPLE_DATAPLANE_RESOURCE" > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH).yaml
	$(DISTRIBUTION_FOLDER)/bin/kuma-dp run --log-level=debug --dataplane-token-file /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token --dataplane-file /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH).yaml

.PHONY: run/xds-client
run/xds-client:
	go run ./tools/xds-client/... run --dps "${NUM_OF_DATAPLANES}" --services "${NUM_OF_SERVICES}" --xds-server-address "${KUMA_CP_ADDRESS}"

.PHONY: run/echo-server
run/echo-server: build/test-server
	$(BUILD_ARTIFACTS_DIR)/test-server/test-server echo --port=$(EXAMPLE_DATAPLANE_SERVICE_PORT)

.PHONY: run/envoy/config_dump
run/envoy/config_dump: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump
