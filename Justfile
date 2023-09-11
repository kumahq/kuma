
set dotenv-load

use_debugger := if env_var_or_default("DEBUGGER", "") != "" { "yes" } else { "no" }
artifacts := "./build/artifacts-darwin-amd64"
kumactl := artifacts / "kumactl/kumactl"
kuma-cp := artifacts / "kuma-cp/kuma-cp"
kuma-dp := artifacts / "kuma-dp/kuma-dp"

dev-kumactl := if use_debugger == "yes" { "dlv debug app/kumactl/main.go --" } else { kumactl }
dev-kuma-cp := if use_debugger == "yes" { "dlv debug app/kuma-cp/main.go --" } else { kuma-cp }
dev-kuma-dp := if use_debugger == "yes" { "dlv debug app/kuma-dp/main.go --" } else { kuma-dp }

certs-path := "./build/koyeb/certs"
client-certs-path := "./build/koyeb/client-certs"

default:
  @just --list

run-db:
  -docker run -p 5532:5432 --rm --name kuma-db -e POSTGRES_USER=kuma -e POSTGRES_PASSWORD=kuma -e POSTGRES_DB=kuma -d postgres:14
  sleep 5
  docker run --rm --link kuma-db:kuma-db -e PGHOST=kuma-db -e PGUSER=kuma -e PGPASSWORD=kuma postgres:14 /bin/bash -c "echo 'CREATE DATABASE \"cp-global\" ENCODING UTF8;' | psql"
  docker run --rm --link kuma-db:kuma-db -e PGHOST=kuma-db -e PGUSER=kuma -e PGPASSWORD=kuma postgres:14 /bin/bash -c "echo 'CREATE DATABASE \"cp-par1\" ENCODING UTF8;' | psql"

stop-db:
  docker stop kuma-db

# This generates a CA cert and its key. Those are expected to be loaded into each mesh
generate-ca-cert:
  mkdir -p {{certs-path}}
  echo "\n[req]\ndistinguished_name=dn\n[ dn ]\n[ ext ]\nbasicConstraints=CA:TRUE,pathlen:0\nkeyUsage=keyCertSign\n" > /tmp/ca_config
  openssl req -config /tmp/ca_config -new -newkey rsa:2048 -nodes -subj "/CN=Hello" -x509 -extensions ext -keyout {{certs-path}}/key.pem -out {{certs-path}}/crt.pem

# NOTE: the verification of the IGW server certif does not seem to work super well though
generate-client-certs:
  mkdir -p {{client-certs-path}}
  openssl genrsa -out {{client-certs-path}}/client.key 2048
  openssl req -new -sha256 -key {{client-certs-path}}/client.key -subj "/CN=Hello" -out {{client-certs-path}}/client.csr
  openssl x509 -req -in {{client-certs-path}}/client.csr -CA {{certs-path}}/crt.pem -CAkey {{certs-path}}/key.pem -CAcreateserial -extfile ./koyeb/samples/client-cert-extensions.ext -out {{client-certs-path}}/client.crt -days 5000 -sha256
  openssl x509 -in {{client-certs-path}}/client.crt -out {{client-certs-path}}/client.pem -text
  openssl x509 -in {{certs-path}}/crt.pem -out {{client-certs-path}}/client-ca.pem -text
  openssl rsa -in {{client-certs-path}}/client.key -text > {{client-certs-path}}/client-key.pem
  openssl x509 -req -in {{client-certs-path}}/client.csr -CA {{certs-path}}/crt.pem -CAkey {{certs-path}}/key.pem -CAcreateserial -extfile ./koyeb/samples/client-cert-extensions-with-spiffe.ext -out {{client-certs-path}}/final.crt -days 5000 -sha256
  openssl x509 -in {{client-certs-path}}/client.crt -out {{client-certs-path}}/final.pem -text

_build-cp:
  make build/kuma-cp

_build-dp:
  make build/kuma-dp
  make build/kumactl

cp-global: _build-cp
  {{kuma-cp}} -c koyeb/samples/cp-global.yaml migrate up
  {{dev-kuma-cp}} -c koyeb/samples/cp-global.yaml run

cp-par1: _build-cp
  {{kuma-cp}} -c koyeb/samples/cp-par1.yaml migrate up
  {{dev-kuma-cp}} -c koyeb/samples/cp-par1.yaml run

_create_secret name mesh path:
  export SECRET_NAME={{name}} ; \
  export SECRET_MESH={{mesh}} ; \
  export SECRET_BASE64_ENCODED=$(cat {{path}} | base64) ; \
  cat koyeb/samples/secret-template.yaml | envsubst | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -

_inject_ca mesh:
  @just _create_secret manually-generated-ca-cert {{mesh}} ./build/koyeb/certs/crt.pem
  @just _create_secret manually-generated-ca-key {{mesh}} ./build/koyeb/certs/key.pem

_init-default: (_inject_ca "default")
  # Upsert default mesh
  cat koyeb/samples/default-mesh.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -
  cat koyeb/samples/default-meshtrace.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -

ingress: _build-dp _init-default
  {{kumactl}} generate zone-token --zone=par1 --valid-for 720h --scope ingress > /tmp/dp-token-ingress
  {{dev-kuma-dp}} run \
    --proxy-type=ingress \
    --cp-address=https://127.0.0.1:5678 \
    --dataplane-token-file=/tmp/dp-token-ingress \
    --dataplane-file=koyeb/samples/ingress-par1.yaml


#glb: _build-dp _init-default
#  {{artifacts}}/kumactl/kumactl generate zone-ingress-token --zone par1  > /tmp/dp-token-glb
#  {{artifacts}}/kuma-dp/kuma-dp run --dataplane-token-file /tmp/dp-token-glb --dns-coredns-config-template-path ./koyeb/samples/Corefile --dns-coredns-port 10053 --dns-envoy-port 10050 --log-level info --cp-address https://localhost:5678 --ca-cert-file ./build/koyeb/tls-cert/ca.pem --admin-port 4243 -d ./koyeb/samples/ingress-glb.yaml  --proxy-type ingress

igw: _build-dp _init-default
  cat koyeb/samples/default-meshgateway.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -

  {{kumactl}} generate dataplane-token -m default --valid-for 720h --config-file koyeb/samples/kumactl-configs/par1-cp.yaml > /tmp/igw-par1-token
  {{dev-kuma-dp}} run --dataplane-token-file /tmp/igw-par1-token --log-level info --cp-address https://localhost:5678 -d ./koyeb/samples/ingress-gateway-par1.yaml

test-grpc:
  # Check that the target container is live
  grpcurl --plaintext localhost:8004 main.HelloWorld/Greeting
  @echo

  # Check that the IGW routes correctly
  grpcurl --plaintext -H "X-Koyeb-Route: dp-8004_prod" localhost:5601 main.HelloWorld/Greeting
  @echo

#  grpcurl -authority grpc.local.koyeb.app -plaintext -servername grpc.local.koyeb.app localhost:5600 list
#  grpcurl -authority grpc.local.koyeb.app -plaintext -servername grpc.local.koyeb.app localhost:5600 main.HelloWorld/Greeting

test-ws:
  echo "hello" | websocat  ws://127.0.0.1:5601 -H 'x-koyeb-route: dp-8011_prod'

test-http2:
  # Check that the target container is live
  curl http://localhost:8002/health -s --fail --output /dev/null
  curl http://localhost:8002/health -s --http2 -s --fail  --output /dev/null
  curl http://localhost:8002/health --http2-prior-knowledge -s --fail --output /dev/null
  @echo

  # Check that the IGW routes correctly
  curl http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" -s --fail --output /dev/null
  curl http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" --http2 -s --fail --output /dev/null
  curl http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" --http2-prior-knowledge -s --fail --output /dev/null
  @echo

#  curl http://localhost:8002/health -s --fail --output /dev/null
#  curl http://localhost:8002/health --http2 -s --fail  --output /dev/null
#  curl http://localhost:8002/health --http2-prior-knowledge -s --fail --output /dev/null
#  curl http://localhost:5600/health -H "host: http2.local.koyeb.app" -s --fail --output /dev/null
#  curl http://localhost:5600/health -H "host: http2.local.koyeb.app" --http2 -s --fail --output /dev/null
#  curl http://localhost:5600/health -H "host: http2.local.koyeb.app" --http2-prior-knowledge -s --fail --output /dev/null

test-http:
  # Check that the target container is live
  curl http://localhost:8001/health -s --fail --output /dev/null
  @echo

  # Check that the IGW is live
  curl http://localhost:5601/health
  @echo

  # Check that the IGW is live in HTTPs
  curl -k https://localhost:5602/health --key {{client-certs-path}}/client.key --cert {{client-certs-path}}/final.pem --cacert {{certs-path}}/crt.pem
  @echo

  # Check that the IGW routing works in HTTP
  curl -k http://localhost:5601 -H "x-koyeb-route: dp-8001_prod" -s --fail --output /dev/null
  @echo

  # Check that the IGW routing works in HTTPS
  curl -k https://localhost:5602 --key {{client-certs-path}}/client.key --cert {{client-certs-path}}/final.pem --cacert {{certs-path}}/crt.pem -H "x-koyeb-route: dp-8001_prod" -s --fail --output /dev/null
  @echo

#  curl http://localhost:5600/health
#  @echo
#  curl http://127.0.0.1:5600/health
#  curl http://localhost:5600 -H "host: http.local.koyeb.app" -s --fail --output /dev/null
#  curl http://localhost:5600 -H "host: http.local.koyeb.app" --http2-prior-knowledge -s --fail --output /dev/null
#
#test: test-http test-http2 test-grpc

dp-container-ws:
  docker run -ti -p 8011:8080 jmalloc/echo-server

dp-container1:
  docker run -ti -p 8001-8010:8001-8010 kalmhq/echoserver:latest

dp-container2:
  # This container should never receive any request. It should be ensured by abc's TrafficRoute policy
  docker run -p 8012:5678 hashicorp/http-echo -text="ERROR!! I'm a leftover container for a service. I do not have the right koyeb.com/global-deployment tag, hence I should not receive any request!"

dp dp_name="dp" mesh="abc": _build-dp
  @just _inject_ca {{mesh}}

  # Upsert default mesh
  cat koyeb/samples/default-mesh.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -

  # Upsert abc mesh
  cat ./koyeb/samples/{{mesh}}-mesh.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -
  # Upsert abc Virtual Outbound
  cat ./koyeb/samples/abc-virtual-outbound.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -
  # Upsert abc TrafficRoute
  cat ./koyeb/samples/abc-traffic-route.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -
  # Upsert abc MeshTrace
  cat koyeb/samples/abc-meshtrace.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/global-cp.yaml -f -

  # Wait some time for the mesh to be propagated to the zonal CP...
  sleep 2

  # Finally, upsert the dataplane
  cat ./koyeb/samples/{{mesh}}-{{dp_name}}.yaml | {{kumactl}} apply --config-file koyeb/samples/kumactl-configs/par1-cp.yaml -f -


  # Generate a token for the dataplane instance
  {{kumactl}} generate dataplane-token -m {{mesh}} --valid-for 720h --config-file koyeb/samples/kumactl-configs/par1-cp.yaml > /tmp/{{mesh}}-{{dp_name}}-token
  # Run the dataplane
  {{dev-kuma-dp}} run --dataplane-token-file /tmp/{{mesh}}-{{dp_name}}-token --name {{dp_name}} --log-level info --cp-address https://localhost:5678 --mesh {{mesh}}
