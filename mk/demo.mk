.PHONY: demo/dp-sever
demo/dp-sever: # start dp-server
	$(KUMA_DIR)/demo/grpc/server/start_grpc_server_dp.sh > $(KUMA_DIR)/log/dp-server-$(shell date '+%Y%m%d%H%M%S').log 2>&1 &


.PHONY: demo/grpc-sever
demo/grpc-sever: # start grpc-server
	$(KUMA_DIR)/build/artifacts-darwin-amd64/test-server/test-server grpc server --ip 10.53.36.94 --port 2345 > $(KUMA_DIR)/log/grpc-server-$(shell date '+%Y%m%d%H%M%S').log 2>&1 &


.PHONY: demo/dp-client
demo/dp-client: # start dp-client
	$(KUMA_DIR)/demo/grpc/client/start_grpc_client_dp.sh > $(KUMA_DIR)/log/dp-client-$(shell date '+%Y%m%d%H%M%S').log 2>&1 &


.PHONY: demo/grpc-client
demo/grpc-client: # start grpc-client
	$(KUMA_DIR)/build/artifacts-darwin-amd64/test-server/test-server grpc client --address 127.0.0.1:8989  > $(KUMA_DIR)/log/grpc-client-$(shell date '+%Y%m%d%H%M%S').log 2>&1 &


.PHONY: demo/start
demo/start: demo/dp-sever  demo/grpc-sever demo/dp-client demo/grpc-client

.PHONY: demo/stop
demo/stop:
	ps axu | grep -E 'test-server|envoy' |awk -F ' ' '{print $$2}' | xargs kill -9

