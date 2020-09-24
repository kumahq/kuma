.PHONY: ecs/clean ecs/deploy \
		ecs/kuma-vpc/deploy ecs/kuma-cp/deploy ecs/kuma-cp-global/deploy ecs/kuma-cp-remote/deploy \
		ecs/workloads/deploy ecs/httpbin/deploy ecs/ssh-client/deploy

ecs/clean: ecs/httpbin/clean ecs/ssh-client/clean ecs/kuma-cp/clean ecs/kuma-vpc/clean

ecs/%/clean:
	aws cloudformation delete-stack --stack-name $* 2>&1 >/dev/null

ecs/deploy: ecs/kuma-vpc/deploy ecs/kuma-cp/deploy ecs/workloads/deploy

ecs/kuma-vpc/deploy:
	aws cloudformation deploy \
        --capabilities CAPABILITY_IAM \
        --stack-name kuma-vpc \
        --template-file kuma-vpc.yaml

ecs/kuma-cp/deploy:
	aws cloudformation deploy \
		--capabilities CAPABILITY_IAM \
		--stack-name kuma-cp \
		--template-file kuma-cp-standalone.yaml

ecs/kuma-cp-global/deploy:
	aws cloudformation deploy \
        --capabilities CAPABILITY_IAM \
        --stack-name kuma-cp-global \
        --template-file kuma-cp-global.yaml

ecs/kuma-cp-remote/deploy:
	aws cloudformation deploy \
            --capabilities CAPABILITY_IAM \
            --stack-name kuma-cp-remote \
            --template-file kuma-cp-remote.yaml

ecs/workloads/deploy: ecs/httpbin/deploy ecs/ssh-client/deploy

ecs/httpbin/deploy:
	aws cloudformation deploy \
		--capabilities CAPABILITY_IAM \
		--stack-name httpbin \
		--template-file workload.yaml \
		--parameter-overrides \
		  DesiredCount=1 \
		  DPToken=${DATAPLANE_TOKEN_HTTPBIN} \
		  WorkloadImage=kennethreitz/httpbin \
		  WorkloadName=httpbin \
		  WorkloadProtocol=tcp

ecs/ssh-client/deploy:
	aws cloudformation deploy \
		--capabilities CAPABILITY_IAM \
		--stack-name ssh-client \
		--template-file workload.yaml \
		--parameter-overrides \
		  DesiredCount=1 \
		  DPToken=${DATAPLANE_TOKEN_SSHCLIENT} \
		  WorkloadImage=sickp/alpine-sshd:latest \
		  WorkloadName=ssh-client \
		  WorkloadProtocol=tcp
