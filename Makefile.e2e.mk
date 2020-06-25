.PHONY: build/example/docker-compose load/example/docker-compose \
		deploy/example/docker-compose undeploy/example/docker-compose \
		wait/example/docker-compose curl/example/docker-compose stats/example/docker-compose \
		verify/example/docker-compose/inbound verify/example/docker-compose/outbound verify/example/docker-compose \
		verify/traffic-routing/docker-compose/without-mtls verify/traffic-routing/docker-compose/with-mtls \
		apply/traffic-routing/docker-compose/mtls wait/traffic-routing/docker-compose/mtls \
		apply/traffic-routing/docker-compose/no-mtls wait/traffic-routing/docker-compose/no-mtls \
		verify/traffic-routing/docker-compose/workflow \
		wait/traffic-routing/docker-compose \
		verify/traffic-routing/docker-compose/default-route \
		apply/traffic-routing/docker-compose/web-to-backend-route \
		wait/traffic-routing/docker-compose/web-to-backend-route \
		verify/traffic-routing/docker-compose/web-to-backend-route \
		delete/traffic-routing/docker-compose/web-to-backend-route \
		wait/traffic-routing/docker-compose/no-web-to-backend-route \
		build/example/minikube load/example/minikube \
		deploy/kuma/minikube \
		deploy/example/minikube wait/example/minikube \
		deploy/example/minikube/metrics \
		undeploy/example/minikube \
		apply/example/minikube/mtls wait/example/minikube/mtls \
		curl/example/minikube stats/example/minikube \
		verify/example/minikube/inbound verify/example/minikube/outbound verify/example/minikube \
		verify/example/minikube/mtls/outbound verify/example/minikube/mtls \
		undeploy/example/minikube \
		deploy/traffic-routing/minikube \
		verify/traffic-routing/minikube/without-mtls verify/traffic-routing/minikube/with-mtls \
		verify/traffic-routing/minikube/workflow \
		apply/traffic-routing/minikube/mtls wait/traffic-routing/minikube/mtls \
		apply/traffic-routing/minikube/no-mtls wait/traffic-routing/minikube/no-mtls \
		wait/traffic-routing/minikube \
		verify/traffic-routing/minikube/default-route \
		apply/traffic-routing/minikube/web-to-backend-route wait/traffic-routing/minikube/web-to-backend-route \
		verify/traffic-routing/minikube/web-to-backend-route \
		delete/traffic-routing/minikube/web-to-backend-route wait/traffic-routing/minikube/no-web-to-backend-route \
		undeploy/traffic-routing/minikube

DOCKER_COMPOSE_OPTIONS ?=

#
# Re-usable snippets
#

define pull_docker_images
	if [ "$(KUMACTL_INSTALL_USE_LOCAL_IMAGES)" != "true" ]; then \
		echo "Pulling Docker images ..." \
		&& set -x \
		&& docker pull $(KUMA_CP_DOCKER_IMAGE) \
		&& docker pull $(KUMA_DP_DOCKER_IMAGE) \
		&& docker pull $(KUMA_INIT_DOCKER_IMAGE) \
		&& docker pull $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) \
		&& docker pull $(KUMACTL_DOCKER_IMAGE) \
		&& set +x \
		&& echo "Pulling is now complete" ; \
	fi
endef

define docker_compose
	docker-compose -f tools/e2e/examples/docker-compose/docker-compose.yaml
endef

define kubectl_exec
	kubectl -n $(1) exec -ti $$( kubectl -n $(1) get pods -l app=$(2) -o=jsonpath='{.items[0].metadata.name}' ) -c $(3) --
endef

define wait_for_client_service
	sh -c ' \
		for i in `seq 1 60`; do \
			echo -n "try #$$i: " ; \
			curl --silent --show-error --fail --include http://localhost:$(1) ; \
			if [[ $$? -eq 0 ]]; then \
				exit 0; \
			fi; \
			sleep 1; \
		done; \
		echo -e "\nError: failed to get a successful response" ; \
		exit 1 ; \
	'
endef

define curl_example_client
	sh -c ' \
		set -e ; \
		for i in `seq 1 5`; do \
			if [[ $$(curl -s http://localhost:3000 | jq -r ".headers[\"host\"]" ) = "mockbin.org" ]]; then \
				echo "request #$$i successful" ; \
			else \
				echo "request #$$i failed" ; \
				exit 1 ; \
			fi ; \
			sleep 1 ; \
		done \
	'
endef

define envoy_active_mtls_listeners_count
	curl -s localhost:9901/config_dump \
	| jq ".configs[] \
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v3.ListenersConfigDump\") \
	| .dynamic_listeners[] \
	| select(.name | startswith(\"$(1)\")) \
	| select(.active_state.listener.address.socket_address.port_value == $(2)) \
	| select(.active_state.listener.filter_chains[] \
		| (.transport_socket.typed_config.common_tls_context \
			and .transport_socket.typed_config.common_tls_context.tls_certificate_sds_secret_configs[] .name == \"identity_cert\") \
			and (.transport_socket.typed_config.common_tls_context.combined_validation_context.validation_context_sds_secret_config.name == \"mesh_ca\") \
			and (.transport_socket.typed_config.require_client_certificate == true) \
	  ) " \
	| jq -s ". | length"
endef

define envoy_active_mtls_clusters_count
	curl -s localhost:9901/config_dump \
	| jq ".configs[] \
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v3.ClustersConfigDump\") \
	| .dynamic_active_clusters[] \
	| select(.cluster.name == \"$(1)\") \
	| select(.cluster.transport_socket.typed_config.common_tls_context) \
	| select(.cluster.transport_socket.typed_config.common_tls_context | \
		 (.tls_certificate_sds_secret_configs[] | .name == \"identity_cert\") and (.combined_validation_context.validation_context_sds_secret_config.name == \"mesh_ca\") \
	  ) " \
	| jq -s ". | length"
endef

define verify_example_inbound
	@echo "Checking number of Inbound requests via Envoy ..."
	test $$( $(1) \
		wget -qO- http://localhost:9901/stats/prometheus | \
		grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8000"}' | \
		awk '{print $$2}' | tr -d [:space:] \
	) -ge 5
	@echo "Check passed!"
endef

define verify_example_outbound
	@echo "Checking number of Outbound requests via Envoy ..."
	test $$( $(1) \
		wget -qO- http://localhost:9901/stats/prometheus | \
		grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}' | \
		awk '{print $$2}' | tr -d [:space:] \
	) -ge 1
	@echo "Check passed!"
endef

define curl_web_to_backend
	sh -c ' \
		set -e ; \
		for i in `seq 1 100`; do \
			curl --silent --show-error --fail http://localhost:6060 ; \
		done; \
	'
endef

define envoy_active_routing_listeners_count
	curl -s localhost:9901/config_dump \
	| jq ".configs[] \
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v3.ListenersConfigDump\") \
	| .dynamic_listeners[] \
	| select(.name | startswith(\"$(1)\")) \
	| select(.active_state.listener.address.socket_address.port_value == $(2)) \
	| select(.active_state.listener.filter_chains[] | .filters[] \
		 | select((.name = \"envoy.tcp_proxy\") \
			and (.typed_config.cluster == \"$(3)\")) \
	  ) " \
	| jq -s ". | length"
endef

define envoy_active_routing_clusters_count
	curl -s localhost:9901/config_dump \
	| jq ".configs[] \
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v3.ClustersConfigDump\") \
	| .dynamic_active_clusters[] \
	| select(.cluster.name == \"$(1)\")" \
	| jq -s ". | length"
endef

#
# Docker Compose setup
#

build/example/docker-compose: images ## Docker Compose: Build Docker images of the Control Plane

load/example/docker-compose: docker/load ## Docker Compose: Load Docker images

deploy/example/docker-compose: ## Docker Compose: Run example setup
	$(call pull_docker_images)
	if [ "$(KUMACTL_INSTALL_USE_LOCAL_IMAGES)" != "true" ]; then \
		$(call docker_compose) pull ; \
	fi
	$(call docker_compose) up $(DOCKER_COMPOSE_OPTIONS)

undeploy/example/docker-compose: ## Docker Compose: Remove example setup
	$(call docker_compose) down -v

wait/example/docker-compose: ## Docker Compose: Wait for example setup to get ready
	$(call docker_compose) exec kuma-example-client $(call wait_for_client_service,3000)

curl/example/docker-compose: ## Docker Compose: Make sample requests to the example setup
	$(call docker_compose) exec kuma-example-client $(call curl_example_client)

verify/example/docker-compose/inbound:
	$(call verify_example_inbound,$(call docker_compose) exec kuma-example-app)

verify/example/docker-compose/outbound:
	@echo "Checking number of Outbound requests via Envoy ..."
	@echo "Not implemented"

verify/example/docker-compose: verify/example/docker-compose/inbound verify/example/docker-compose/outbound ## Docker Compose: Verify Envoy stats (after sample requests)

stats/example/docker-compose: ## Docker Compose: Observe Envoy metrics from the example setup
	$(call docker_compose) exec kuma-example-app curl -s localhost:9901/stats/prometheus | grep upstream_rq_total

verify/traffic-routing/docker-compose/without-mtls: \
	apply/traffic-routing/docker-compose/no-mtls \
	wait/traffic-routing/docker-compose/no-mtls \
	verify/traffic-routing/docker-compose/workflow

verify/traffic-routing/docker-compose/with-mtls: \
	apply/traffic-routing/docker-compose/mtls \
	wait/traffic-routing/docker-compose/mtls \
	verify/traffic-routing/docker-compose/workflow

verify/traffic-routing/docker-compose/workflow: \
	wait/traffic-routing/docker-compose \
	verify/traffic-routing/docker-compose/default-route \
	apply/traffic-routing/docker-compose/web-to-backend-route \
	wait/traffic-routing/docker-compose/web-to-backend-route \
	verify/traffic-routing/docker-compose/web-to-backend-route \
	delete/traffic-routing/docker-compose/web-to-backend-route \
	wait/traffic-routing/docker-compose/no-web-to-backend-route \
	verify/traffic-routing/docker-compose/default-route

apply/traffic-routing/docker-compose/mtls: ## Docker Compose: enable mTLS
	@echo
	@echo "Enabling mTLS ..."
	@echo
	$(call docker_compose) exec kumactl kumactl apply -f /kuma-example/policies/mtls.yaml
	$(call docker_compose) exec kumactl kumactl apply -f /kuma-example/policies/everyone-to-everyone.traffic-permission.yaml

wait/traffic-routing/docker-compose/mtls: ## Docker Compose: Wait until incoming Listener and outgoing Cluster have been configured for mTLS
	@echo
	@echo "Waiting until incoming Listener and outgoing Cluster have been configured for mTLS ..."
	@echo
	$(call docker_compose) exec kuma-example-web sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,6060) ) -eq 1 ]]; then echo "listener has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener has not been configured for mTLS" ; exit 1'
	$(call docker_compose) exec kuma-example-web sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,kuma-example-backend) ) -eq 1 ]]; then echo "cluster has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster has not been configured for mTLS" ; exit 1'

apply/traffic-routing/docker-compose/no-mtls: ## Docker Compose: disable mTLS
	@echo
	@echo "Disabling mTLS ..."
	@echo
	$(call docker_compose) exec kumactl kumactl apply -f /kuma-example/policies/no-mtls.yaml

wait/traffic-routing/docker-compose/no-mtls: ## Docker Compose: Wait until mTLS has been disabled on incoming Listener and outgoing Cluster
	@echo
	@echo "Waiting until mTLS has been disabled on incoming Listener and outgoing Cluster ..."
	@echo
	$(call docker_compose) exec kuma-example-web sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,6060) ) -eq 0 ]]; then echo "listener is no longer configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener is still configured for mTLS" ; exit 1'
	$(call docker_compose) exec kuma-example-web sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,kuma-example-backend) ) -eq 0 ]]; then echo "cluster is no longer configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster is still configured for mTLS" ; exit 1'

wait/traffic-routing/docker-compose: ## Docker Compose: Wait for example setup for TrafficRoute to get ready
	@echo
	@echo "Waiting for example setup for TrafficRoute to get ready ..."
	@echo
	$(call docker_compose) exec kuma-example-web $(call wait_for_client_service,6060)

verify/traffic-routing/docker-compose/default-route: ## Docker Compose: Make sample requests to example setup for TrafficRoute
	@echo
	@echo "Checking default traffic routing policy (round robin) ..."
	@echo
	test $$( $(call docker_compose) exec kuma-example-web $(call curl_web_to_backend) | sort | uniq | wc -l ) -eq 2
	@echo "Check passed!"

apply/traffic-routing/docker-compose/web-to-backend-route: ## Docker Compose: create "web-to-backend" route
	@echo
	@echo "Creating 'web-to-backend' route ..."
	@echo
	$(call docker_compose) exec kumactl kumactl apply -f /kuma-example/policies/web-to-backend.traffic-route.yaml

wait/traffic-routing/docker-compose/web-to-backend-route: ## Docker Compose: Wait until custom "web-to-backend" TrafficRoute is applied
	@echo
	@echo "Waiting until custom 'web-to-backend' TrafficRoute is applied ..."
	@echo
	sleep 10 # todo(jakubdyszkiewicz) I don't want to build a logic of detecting lb split in cluster since this test will be soon rewritten to E2E Go Framework

verify/traffic-routing/docker-compose/web-to-backend-route: ## Docker Compose: Make sample requests to example setup for TrafficRoute
	@echo
	@echo "Checking custom traffic routing policy (100% to v2) ..."
	@echo
	test $$( $(call docker_compose) exec kuma-example-web $(call curl_web_to_backend) | sort | uniq | tr -d '\r\n' ) = '{"version":"v2"}'
	@echo "Check passed!"

delete/traffic-routing/docker-compose/web-to-backend-route: ## Docker Compose: delete "web-to-backend" route
	@echo
	@echo "Deleting 'web-to-backend' route ..."
	@echo
	$(call docker_compose) exec kumactl kumactl delete traffic-route web-to-backend

wait/traffic-routing/docker-compose/no-web-to-backend-route: ## Docker Compose: Wait until custom "web-to-backend" TrafficRoute is removed
	@echo
	@echo "Waiting until custom 'web-to-backend' TrafficRoute is removed ..."
	@echo
	sleep 10 # todo(jakubdyszkiewicz) I don't want to build a logic of detecting lb split in cluster since this test will be soon rewritten to E2E Go Framework

#
# Minikube setup
#

build/example/minikube: ## Minikube: Build Docker images inside Minikube
	eval $$(minikube docker-env) && $(MAKE) images

load/example/minikube: ## Minikube: Load Docker images into Minikube
	eval $$(minikube docker-env) && $(MAKE) docker/load

deploy/kuma/minikube: ## Minikube: Deploy Kuma with no demo app
	eval $$(minikube docker-env) && $(call pull_docker_images)
	eval $$(minikube docker-env) && docker run --rm $(KUMACTL_DOCKER_IMAGE) kumactl install control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | kubectl apply -f -
	kubectl wait --timeout=60s --for=condition=Available -n kuma-system deployment/kuma-control-plane
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-system pods -l app=kuma-control-plane

deploy/example/minikube: deploy/kuma/minikube ## Minikube: Deploy example setup
	kubectl apply -f tools/e2e/examples/minikube/kuma-demo/
	kubectl wait --timeout=60s --for=condition=Available -n kuma-demo deployment/demo-app
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-demo pods -l app=demo-app
	kubectl wait --timeout=60s --for=condition=Available -n kuma-demo deployment/demo-client
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-demo pods -l app=demo-client

deploy/example/minikube/metrics: ## Minikube: Deploy metrics setup
	eval $$(minikube docker-env) && $(call pull_docker_images)
	eval $$(minikube docker-env) && docker run --rm $(KUMACTL_DOCKER_IMAGE) kumactl install metrics $(KUMACTL_INSTALL_METRICS_IMAGES) | kubectl apply -f -
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-metrics pods -l app=prometheus

apply/example/minikube/mtls: ## Minikube: enable mTLS
	kubectl apply -f tools/e2e/examples/minikube/policies/mtls.yaml

wait/example/minikube: ## Minikube: Wait for demo setup to get ready
	$(call kubectl_exec,kuma-demo,demo-client,demo-client) $(call wait_for_client_service,3000)

wait/example/minikube/mtls: ## Minikube: Wait until incoming Listener and outgoing Cluster have been configured for mTLS
	$(call kubectl_exec,kuma-demo,demo-client,demo-client) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,3000) ) -eq 1 ]]; then echo "listener has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener has not been configured for mTLS" ; exit 1'
	$(call kubectl_exec,kuma-demo,demo-client,demo-client) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,demo-app_kuma-demo_svc_8000) ) -eq 1 ]]; then echo "cluster has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster has not been configured for mTLS" ; exit 1'

curl/example/minikube: ## Minikube: Make sample requests to demo setup
	$(call kubectl_exec,kuma-demo,demo-client,demo-client) $(call curl_example_client)

stats/example/minikube: ## Minikube: Observe Envoy metrics from demo setup
	$(call kubectl_exec,kuma-demo,demo-app,kuma-sidecar) wget -qO- http://localhost:9901/stats/prometheus | grep upstream_rq_total

verify/example/minikube/inbound:
	$(call verify_example_inbound,$(call kubectl_exec,kuma-demo,demo-app,kuma-sidecar))

verify/example/minikube/outbound:
	$(call verify_example_outbound,$(call kubectl_exec,kuma-demo,demo-app,kuma-sidecar))

verify/example/minikube: verify/example/minikube/inbound verify/example/minikube/outbound ## Minikube: Verify Envoy stats (after sample requests)

verify/example/minikube/mtls: verify/example/minikube/mtls/outbound ## Minikube: Verify Envoy mTLS stats (after sample requests)

verify/example/minikube/mtls/outbound:
	@echo "Checking number of Outbound mTLS requests via Envoy ..."
	test $$( $(call kubectl_exec,kuma-demo,demo-client,kuma-sidecar) wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_ssl_handshake{envoy_cluster_name="demo-app_kuma-demo_svc_8000"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 5
	@echo "Check passed!"

kumactl/example/minikube:
	cat tools/e2e/examples/minikube/kumactl_workflow.sh | docker run -i --rm --user $$(id -u):$$(id -g) --network host -v $$HOME/.kube:/tmp/.kube -v $$HOME/.minikube:$$HOME/.minikube -e HOME=/tmp -w /tmp $(KUMACTL_DOCKER_IMAGE)

undeploy/example/minikube: ## Minikube: Undeploy example setup
	kubectl delete -f tools/e2e/examples/minikube/kuma-demo/

deploy/traffic-routing/minikube: ## Minikube: Deploy example setup for TrafficRoute
	@echo
	@echo "Deploying example setup for TrafficRoute ..."
	@echo
	kubectl apply -f tools/e2e/examples/minikube/kuma-routing/
	kubectl wait --timeout=60s --for=condition=Available -n kuma-example deployment/kuma-example-web
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-example pods -l app=kuma-example-web
	kubectl wait --timeout=60s --for=condition=Available -n kuma-example deployment/kuma-example-backend-v1
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-example pods -l app=kuma-example-backend,version=v1
	kubectl wait --timeout=60s --for=condition=Available -n kuma-example deployment/kuma-example-backend-v2
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-example pods -l app=kuma-example-backend,version=v2

verify/traffic-routing/minikube/without-mtls: \
	apply/traffic-routing/minikube/no-mtls \
	wait/traffic-routing/minikube/no-mtls \
	verify/traffic-routing/minikube/workflow

verify/traffic-routing/minikube/with-mtls: \
	apply/traffic-routing/minikube/mtls \
	wait/traffic-routing/minikube/mtls \
	verify/traffic-routing/minikube/workflow

verify/traffic-routing/minikube/workflow: \
	wait/traffic-routing/minikube \
	verify/traffic-routing/minikube/default-route \
	apply/traffic-routing/minikube/web-to-backend-route \
	wait/traffic-routing/minikube/web-to-backend-route \
	verify/traffic-routing/minikube/web-to-backend-route \
	delete/traffic-routing/minikube/web-to-backend-route \
	wait/traffic-routing/minikube/no-web-to-backend-route \
	verify/traffic-routing/minikube/default-route

apply/traffic-routing/minikube/mtls: ## Minikube: enable mTLS
	@echo
	@echo "Enabling mTLS ..."
	@echo
	kubectl apply -f tools/e2e/examples/minikube/policies/mtls.yaml

wait/traffic-routing/minikube/mtls: ## Minikube: Wait until incoming Listener and outgoing Cluster have been configured for mTLS
	@echo
	@echo "Waiting until incoming Listener and outgoing Cluster have been configured for mTLS ..."
	@echo
	$(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,6060) ) -eq 1 ]]; then echo "listener has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener has not been configured for mTLS" ; exit 1'
	$(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,kuma-example-backend_kuma-example_svc_7070) ) -eq 1 ]]; then echo "cluster has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster has not been configured for mTLS" ; exit 1'

apply/traffic-routing/minikube/no-mtls: ## Minikube: disable mTLS
	@echo
	@echo "Disabling mTLS ..."
	@echo
	kubectl apply -f tools/e2e/examples/minikube/policies/no-mtls.yaml

wait/traffic-routing/minikube/no-mtls: ## Minikube: Wait until mTLS has been disabled on incoming Listener and outgoing Cluster
	@echo
	@echo "Waiting until mTLS has been disabled on incoming Listener and outgoing Cluster ..."
	@echo
	$(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,6060) ) -eq 0 ]]; then echo "listener is no longer configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener is still configured for mTLS" ; exit 1'
	$(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,kuma-example-backend_kuma-example_svc_7070) ) -eq 0 ]]; then echo "cluster is no longer configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster is still configured for mTLS" ; exit 1'

wait/traffic-routing/minikube: ## Minikube: Wait for example setup for TrafficRoute to get ready
	@echo
	@echo "Waiting for example setup for TrafficRoute to get ready ..."
	@echo
	$(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) $(call wait_for_client_service,6060)

verify/traffic-routing/minikube/default-route: ## Minikube: Make sample requests to example setup for TrafficRoute
	@echo
	@echo "Checking default traffic routing policy (round robin) ..."
	@echo
	test $$( $(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) $(call curl_web_to_backend) | sort | uniq | wc -l ) -eq 2
	@echo "Check passed!"

apply/traffic-routing/minikube/web-to-backend-route: ## Minikube: create "web-to-backend" route
	@echo
	@echo "Creating 'web-to-backend' route ..."
	@echo
	kubectl apply -f tools/e2e/examples/minikube/policies/web-to-backend.traffic-route.yaml

wait/traffic-routing/minikube/web-to-backend-route: ## Minikube: Wait until custom "web-to-backend" TrafficRoute is applied
	@echo
	@echo "Waiting until custom 'web-to-backend' TrafficRoute is applied ..."
	@echo
	sleep 10 # todo(jakubdyszkiewicz) I don't want to build a logic of detecting lb split in cluster since this test will be soon rewritten to E2E Go Framework

verify/traffic-routing/minikube/web-to-backend-route: ## Minikube: Make sample requests to example setup for TrafficRoute
	@echo
	@echo "Checking custom traffic routing policy (100% to v2) ..."
	@echo
	test $$( $(call kubectl_exec,kuma-example,kuma-example-web,kuma-example-web) $(call curl_web_to_backend) | sort | uniq | tr -d '\r\n' ) = '{"version":"v2"}'
	@echo "Check passed!"

delete/traffic-routing/minikube/web-to-backend-route: ## Minikube: delete "web-to-backend" route
	@echo
	@echo "Deleting 'web-to-backend' route ..."
	@echo
	kubectl delete -f tools/e2e/examples/minikube/policies/web-to-backend.traffic-route.yaml

wait/traffic-routing/minikube/no-web-to-backend-route: ## Minikube: Wait until custom "web-to-backend" TrafficRoute is removed
	@echo
	@echo "Waiting until custom 'web-to-backend' TrafficRoute is removed ..."
	@echo
	sleep 10 # todo(jakubdyszkiewicz) I don't want to build a logic of detecting lb split in cluster since this test will be soon rewritten to E2E Go Framework

undeploy/traffic-routing/minikube: ## Minikube: Undeploy example setup for TrafficRoute
	@echo
	@echo "Undeploying example setup for TrafficRoute ..."
	@echo
	kubectl delete -f tools/e2e/examples/minikube/kuma-routing/
