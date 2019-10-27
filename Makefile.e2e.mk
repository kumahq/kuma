.PHONY: build/example/minikube load/example/minikube deploy/example/minikube wait/example/minikube apply/example/minikube/mtls wait/example/minikube/mtls curl/example/minikube stats/example/minikube \
		verify/example/minikube/inbound verify/example/minikube/outbound verify/example/minikube \
		verify/example/minikube/mtls/outbound verify/example/minikube/mtls

#
# Re-usable snippets
#

define wait_for_example_client
	sh -c ' \
		for i in `seq 1 60`; do \
			echo -n "try #$$i: " ; \
			curl --silent --show-error --fail --include http://localhost:3000 ; \
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
			if [[ $$(curl -s http://localhost:3000 | jq -r ".headers[\"kong-client-id\"]" ) = "mockbin" ]]; then \
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
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v2alpha.ListenersConfigDump\") \
	| .dynamic_active_listeners[] \
	| select(.listener.name | startswith(\"$(1)\")) \
	| select(.listener.address.socket_address.port_value == $(2)) \
	| select(.listener.filter_chains[] \
		| (.tls_context.common_tls_context.tls_certificate_sds_secret_configs[] .name == \"identity_cert\") \
		  and (.tls_context.common_tls_context.validation_context_sds_secret_config.name == \"mesh_ca\") \
		  and (.tls_context.require_client_certificate == true) ) " \
	| jq -s ". | length"
endef

define envoy_active_mtls_clusters_count
	curl -s localhost:9901/config_dump \
	| jq ".configs[] \
    | select(.[\"@type\"] == \"type.googleapis.com/envoy.admin.v2alpha.ClustersConfigDump\") \
	| .dynamic_active_clusters[] \
	| select(.cluster.name == \"$(1)\") \
	| select(.cluster.tls_context.common_tls_context | \
		 (.tls_certificate_sds_secret_configs[] | .name == \"identity_cert\") and (.validation_context_sds_secret_config.name == \"mesh_ca\") \
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

#
# Minikube setup
#

build/example/minikube: ## Minikube: Build Docker images inside Minikube
	eval $$(minikube docker-env) && $(MAKE) images

load/example/minikube: ## Minikube: Load Docker images into Minikube
	eval $$(minikube docker-env) && $(MAKE) docker/load

deploy/example/minikube: ## Minikube: Deploy example setup
	docker run --rm $(KUMACTL_DOCKER_IMAGE) kumactl install control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | kubectl apply -f -
	kubectl wait --timeout=60s --for=condition=Available -n kuma-system deployment/kuma-injector
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-system pods -l app=kuma-injector
	kubectl apply -f tools/e2e/examples/minikube/kuma-demo/
	kubectl wait --timeout=60s --for=condition=Available -n kuma-demo deployment/demo-app
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-demo pods -l app=demo-app
	kubectl wait --timeout=60s --for=condition=Available -n kuma-demo deployment/demo-client
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-demo pods -l app=demo-client

apply/example/minikube/mtls: ## Minikube: enable mTLS
	kubectl apply -f tools/e2e/examples/minikube/policies/mtls.yaml

wait/example/minikube: ## Minikube: Wait for demo setup to get ready
	kubectl -n kuma-demo exec -ti $$( kubectl -n kuma-demo get pods -l app=demo-client -o=jsonpath='{.items[0].metadata.name}' ) -c demo-client -- $(call wait_for_example_client)

wait/example/minikube/mtls: ## Minikube: Wait until incoming Listener and outgoing Cluster have been configured for mTLS
	kubectl -n kuma-demo exec -ti $$( kubectl -n kuma-demo get pods -l app=demo-client -o=jsonpath='{.items[0].metadata.name}' ) -c demo-client -- sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_listeners_count,inbound,3000) ) -eq 1 ]]; then echo "listener has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: listener has not been configured for mTLS" ; exit 1'
	kubectl -n kuma-demo exec -ti $$( kubectl -n kuma-demo get pods -l app=demo-client -o=jsonpath='{.items[0].metadata.name}' ) -c demo-client -- sh -c 'for i in `seq 1 10`; do echo -n "try #$$i: " ; if [[ $$( $(call envoy_active_mtls_clusters_count,demo-app.kuma-demo.svc:8000) ) -eq 1 ]]; then echo "cluster has been configured for mTLS "; exit 0; fi; sleep 1; done; echo -e "\nError: cluster has not been configured for mTLS" ; exit 1'

curl/example/minikube: ## Minikube: Make sample requests to demo setup
	kubectl -n kuma-demo exec -ti $$( kubectl -n kuma-demo get pods -l app=demo-client -o=jsonpath='{.items[0].metadata.name}' ) -c demo-client -- $(call curl_example_client)

stats/example/minikube: ## Minikube: Observe Envoy metrics from demo setup
	kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep upstream_rq_total

verify/example/minikube/inbound:
	$(call verify_example_inbound,kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- )

verify/example/minikube/outbound:
	$(call verify_example_outbound,kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- )

verify/example/minikube: verify/example/minikube/inbound verify/example/minikube/outbound ## Minikube: Verify Envoy stats (after sample requests)

verify/example/minikube/mtls: verify/example/minikube/mtls/outbound ## Minikube: Verify Envoy mTLS stats (after sample requests)

verify/example/minikube/mtls/outbound:
	@echo "Checking number of Outbound mTLS requests via Envoy ..."
	test $$( kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-client -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_kuma_demo_svc_8000_ssl_handshake{envoy_cluster_name="demo-app"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 5
	@echo "Check passed!"

kumactl/example/minikube:
	cat tools/e2e/examples/minikube/kumactl_workflow.sh | docker run -i --rm --user $$(id -u):$$(id -g) --network host -v $$HOME/.kube:/tmp/.kube -v $$HOME/.minikube:$$HOME/.minikube -e HOME=/tmp -w /tmp $(KUMACTL_DOCKER_IMAGE)
