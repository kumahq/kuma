GO_RUN := CGO_ENABLED=0 go run $(GOFLAGS) $(LD_FLAGS)

EXAMPLE_DATAPLANE_MESH ?= default
EXAMPLE_DATAPLANE_NAME ?= example
ENVOY_ADMIN_PORT ?= 9901

POSTGRES_SSL_MODE ?= disable

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
	KUMA_EXPERIMENTAL_MESHGATEWAY=true \
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
run/example/envoy/universal: run/example/envoy

.PHONY: run/example/envoy
run/example/envoy: build/kuma-dp build/kumactl ## Dev: Run Envoy configured against local Control Plane
	${BUILD_ARTIFACTS_DIR}/kumactl/kumactl generate dataplane-token --name=$(EXAMPLE_DATAPLANE_NAME) --mesh=$(EXAMPLE_DATAPLANE_MESH) > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	KUMA_DATAPLANE_RUNTIME_TOKEN_PATH=/tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token \
	${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp run --log-level=debug

.PHONY: config_dump/example/envoy
config_dump/example/envoy: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

.PHONY: run/universal/memory
run/universal/memory: ## Dev: Run Control Plane locally in universal mode with in-memory store
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=memory \
	KUMA_EXPERIMENTAL_MESHGATEWAY=true \
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
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	KUMA_DATAPLANE_RUNTIME_TOKEN_PATH=/tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token \
	$(GO_RUN) ./app/kuma-dp/main.go run --log-level=debug
