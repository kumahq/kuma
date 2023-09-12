set dotenv-load

use_debugger := if env_var_or_default("DEBUGGER", "") != "" { "yes" } else { "no" }
artifacts := "./build/artifacts-darwin-amd64"
kumactl := artifacts / "kumactl/kumactl"
kuma-cp := artifacts / "kuma-cp/kuma-cp"
kuma-dp := artifacts / "kuma-dp/kuma-dp"

dev-kumactl := if use_debugger == "yes" { "dlv debug app/kumactl/main.go --" } else { kumactl }
dev-kuma-cp := if use_debugger == "yes" { "dlv debug app/kuma-cp/main.go --" } else { kuma-cp }
dev-kuma-dp := if use_debugger == "yes" { "dlv debug app/kuma-dp/main.go --" } else { kuma-dp }

kumactl-configs := "./koyeb/samples/kumactl-configs"

glb-cert-path := "./build/koyeb/glb-certs"
ca-cert-path := "./build/koyeb/certs"
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

gen-certs: _generate-ca-cert _generate-client-certs _generate-glb-cert

# This generates a CA cert and its key. Those are expected to be loaded into each mesh
_generate-ca-cert:
  mkdir -p {{ca-cert-path}}
  echo "\n[req]\ndistinguished_name=dn\n[ dn ]\n[ ext ]\nbasicConstraints=CA:TRUE,pathlen:0\nkeyUsage=keyCertSign\n" > /tmp/ca_config
  openssl req -config /tmp/ca_config -new -newkey rsa:2048 -nodes -subj "/CN=Hello" -x509 -extensions ext -keyout {{ca-cert-path}}/key.pem -out {{ca-cert-path}}/crt.pem

# Generate a client cert so we can verify that we can talk in mTLS to the Ingress Gateway
# NOTE: the verification of the IGW server certif does not seem to work super well though
_generate-client-certs:
  mkdir -p {{client-certs-path}}
  openssl genrsa -out {{client-certs-path}}/client.key 2048
  openssl req -new -sha256 -key {{client-certs-path}}/client.key -subj "/CN=Hello" -out {{client-certs-path}}/client.csr
  openssl x509 -req -in {{client-certs-path}}/client.csr -CA {{ca-cert-path}}/crt.pem -CAkey {{ca-cert-path}}/key.pem -CAcreateserial -extfile ./koyeb/samples/client-cert-extensions.ext -out {{client-certs-path}}/client.crt -days 5000 -sha256
  openssl x509 -in {{client-certs-path}}/client.crt -out {{client-certs-path}}/client.pem -text
  openssl x509 -in {{ca-cert-path}}/crt.pem -out {{client-certs-path}}/client-ca.pem -text
  openssl rsa -in {{client-certs-path}}/client.key -text > {{client-certs-path}}/client-key.pem
  openssl x509 -req -in {{client-certs-path}}/client.csr -CA {{ca-cert-path}}/crt.pem -CAkey {{ca-cert-path}}/key.pem -CAcreateserial -extfile ./koyeb/samples/client-cert-extensions-with-spiffe.ext -out {{client-certs-path}}/final.crt -days 5000 -sha256
  openssl x509 -in {{client-certs-path}}/client.crt -out {{client-certs-path}}/final.pem -text

_generate-glb-cert:
  mkdir -p {{glb-cert-path}}
  {{kumactl}} generate tls-certificate --type=server --hostname=koyeb.app,localhost  --cert-file {{glb-cert-path}}/glb.crt --key-file {{glb-cert-path}}/glb.key

_inject_glb-cert-secret:
  export SECRET_NAME=glb-cert ; \
  export SECRET_MESH=default ; \
  export SECRET_BASE64_ENCODED=$(cat {{glb-cert-path}}/glb.key {{glb-cert-path}}/glb.crt | base64) ; \
  cat koyeb/samples/secret-template.yaml | envsubst | {{kumactl}} apply --config-file {{kumactl-configs}}/kumactl-configs/global-cp.yaml -f -

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
  cat koyeb/samples/secret-template.yaml | envsubst | {{kumactl}} apply --config-file {{kumactl-configs}}/kumactl-configs/global-cp.yaml -f -

_inject_ca mesh:
  @just _create_secret manually-generated-ca-cert {{mesh}} {{ca-cert-path}}/crt.pem
  @just _create_secret manually-generated-ca-key {{mesh}} {{ca-cert-path}}/key.pem

_init-default: (_inject_ca "default")
  # Upsert default mesh
  cat koyeb/samples/mesh-default/mesh.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -
  cat koyeb/samples/mesh-default/meshtrace.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -

ingress: _build-dp _init-default
  {{kumactl}} generate zone-token --zone=par1 --valid-for 720h --scope ingress > /tmp/dp-token-ingress
  {{dev-kuma-dp}} run \
    --proxy-type=ingress \
    --cp-address=https://127.0.0.1:5678 \
    --dataplane-token-file=/tmp/dp-token-ingress \
    --dataplane-file=koyeb/samples/mesh-default/ingress-par1.yaml

glb: _build-dp _init-default _inject_glb-cert-secret
  cat koyeb/samples/mesh-default/mesh-gateway-global-load-balancer.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -
  {{kumactl}} generate dataplane-token -m default --valid-for 720h --config-file {{kumactl-configs}}/par1-cp.yaml > /tmp/glb-par1-token
  {{dev-kuma-dp}} run --dataplane-token-file /tmp/glb-par1-token --log-level info --cp-address https://localhost:5678 -d ./koyeb/samples/mesh-default/global-load-balancer-par1.yaml

igw: _build-dp _init-default
  cat koyeb/samples/mesh-default/mesh-gateway-ingress-gateway.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -

  {{kumactl}} generate dataplane-token -m default --valid-for 720h --config-file {{kumactl-configs}}/par1-cp.yaml > /tmp/igw-par1-token
  {{dev-kuma-dp}} run --dataplane-token-file /tmp/igw-par1-token --log-level info --cp-address https://localhost:5678 -d ./koyeb/samples/mesh-default/ingress-gateway-par1.yaml

test: test-http test-http2 test-grpc test-ws test-glb

test-http:
  # Check that the target container is live
  curl --fail http://localhost:8001/health -s --output /dev/null
  @echo

  # Check that the IGW is live
  curl --fail http://localhost:5601/health
  @echo

  # Check that the IGW is live in HTTPs
  curl --fail -k https://localhost:5602/health --key {{client-certs-path}}/client.key --cert {{client-certs-path}}/final.pem --cacert {{ca-cert-path}}/crt.pem
  @echo

  # Check that the IGW routing works in HTTP
  curl --fail http://localhost:5601 -H "x-koyeb-route: dp-8001_prod" -s --output /dev/null
  @echo

  # Check that the IGW routing works in HTTPS
  curl --fail https://localhost:5602 --key {{client-certs-path}}/client.key --cert {{client-certs-path}}/final.pem --cacert {{ca-cert-path}}/crt.pem -H "x-koyeb-route: dp-8001_prod" -s -k --output /dev/null
  @echo

  # Check that the GLB is live
  curl --fail http://localhost:5600/health
  @echo
  curl --fail -k https://localhost:5603/health
  @echo
  @echo

  # Check that the GLB forwards correctly HTTP1. The grep ensures that the path as perceived by the container is correctly stripped
  curl --fail http://localhost:5600/http -H "host: http.local.koyeb.app" -s | grep 'path: /$'
  curl --fail http://localhost:5600/http -H "host: http.local.koyeb.app" --http2-prior-knowledge -s  | grep 'path: /$'
  curl --fail -k https://localhost:5603/http -H "host: http.local.koyeb.app" -s | grep 'path: /$'
  curl --fail -k https://localhost:5603/http -H "host: http.local.koyeb.app" --http2-prior-knowledge -s  | grep 'path: /$'
  @echo

test-glb:
  # Check that an unknowns Host routes to a 404
  curl http://localhost:5600/ -H "Host: unknown-host.com" --output /dev/null -s -w "%{http_code}\n" | grep 404
  @echo

  # Check that unknown paths return a 404
  curl http://localhost:5600/blibablou -H "host: http.local.koyeb.app" --output /dev/null -s -w "%{http_code}\n" | grep 404
  curl http://localhost:5600           -H "host: http.local.koyeb.app" --output /dev/null -s -w "%{http_code}\n" | grep 404
  @echo

test-grpc:
  # Check that the target container is live
  grpcurl --plaintext localhost:8004 main.HelloWorld/Greeting
  @echo

  # Check that the IGW routes correctly
  grpcurl --plaintext -H "X-Koyeb-Route: dp-8004_prod" localhost:5601 main.HelloWorld/Greeting
  # echo '{}' | INSECURE=yes /Users/nicolas/perso/evans/evans --verbose --host localhost --port 5602 --tls --certkey ./build/koyeb/client-certs/client.key --cert ./build/koyeb/client-certs/final.pem --cacert ./build/koyeb/client-certs/client-ca.pem -r cli --header "x-koyeb-route=dp-8004_prod" call main.HelloWorld.Greeting
  @echo

  # Check that the GLB routes correctly
  grpcurl -authority grpc.koyeb.app -insecure localhost:5603 list
  grpcurl -authority grpc.koyeb.app -insecure localhost:5603 main.HelloWorld/Greeting
  @echo


# Note that we do not test in SSL because websocat seems not to allow it. e2e tests will catch errors on SSL
test-ws:
  # Test that the IGW routes correctly
  echo "hello" | websocat ws://127.0.0.1:5601 -H 'x-koyeb-route: dp-8011_prod'
  # Test that the GLB routes correctly
  echo "hello" | websocat  -t -     ws-ll-c:http-request:tcp:127.0.0.1:5600      --request-header 'Host: http.local.koyeb.app'     --request-header 'Upgrade: websocket'     --request-header 'Sec-WebSocket-Key: mYUkMl6bemnLatx/g7ySfw=='     --request-header 'Sec-WebSocket-Version: 13'     --request-header 'Connection: Upgrade'     --request-uri=/ws
  @echo

test-http2:
  # Check that the target container is live
  curl --fail http://localhost:8002/health -s --output /dev/null
  curl --fail http://localhost:8002/health -s --http2 -s --output /dev/null
  curl --fail http://localhost:8002/health --http2-prior-knowledge -s --output /dev/null
  @echo

  # Check that the IGW routes correctly
  curl --fail http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" -s --output /dev/null
  curl --fail http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" --http2 -s --output /dev/null
  curl --fail http://localhost:5601/health -H "x-koyeb-route: dp-8002_prod" --http2-prior-knowledge -s --output /dev/null
  @echo

  # Check that the GLB routes correctly
  curl --fail http://localhost:5600/http2 -H "host: http.local.koyeb.app" -s --output /dev/null
  curl --fail http://localhost:5600/http2 -H "host: http.local.koyeb.app" --http2 -s --output /dev/null
  curl --fail http://localhost:5600/http2 -H "host: http.local.koyeb.app" --http2-prior-knowledge -s --output /dev/null
  curl --fail https://localhost:5603/http2 -H "host: http.local.koyeb.app" -k -s --output /dev/null
  curl --fail https://localhost:5603/http2 -H "host: http.local.koyeb.app" -k --http2 -s --output /dev/null
  curl --fail https://localhost:5603/http2 -H "host: http.local.koyeb.app" -k --http2-prior-knowledge -s --output /dev/null
  @echo

dp-container-ws:
  docker run -ti -p 8011:8080 jmalloc/echo-server

dp-container1:
  docker run -ti -p 8001-8010:8001-8010 kalmhq/echoserver:latest

dp-container2:
  # This container should never receive any request. It should be ensured by abc's TrafficRoute policy
  docker run -p 8012:5678 hashicorp/http-echo -text="ERROR!! I'm a leftover container for a service. I do not have the right koyeb.com/global-deployment tag, hence I should not receive any request!"

dp: _build-dp (_inject_ca "abc")
  # Upsert default mesh
  cat ./koyeb/samples/mesh-default/mesh.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -

  # Upsert abc mesh
  cat ./koyeb/samples/mesh-abc/mesh.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -
  # Upsert abc Virtual Outbound
  cat ./koyeb/samples/mesh-abc/virtual-outbound.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -
  # Upsert abc TrafficRoute
  cat ./koyeb/samples/mesh-abc/traffic-route.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -
  # Upsert abc MeshTrace
  cat koyeb/samples/mesh-abc/meshtrace.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/global-cp.yaml -f -

  # Wait some time for the mesh to be propagated to the zonal CP...
  sleep 2

  # Finally, upsert the dataplane
  cat ./koyeb/samples/mesh-abc/dp.yaml | {{kumactl}} apply --config-file {{kumactl-configs}}/par1-cp.yaml -f -


  # Generate a token for the dataplane instance
  {{kumactl}} generate dataplane-token -m abc --valid-for 720h --config-file {{kumactl-configs}}/par1-cp.yaml > /tmp/abc-dp-token
  # Run the dataplane
  {{dev-kuma-dp}} run --dataplane-token-file /tmp/abc-dp-token --name dp --log-level info --cp-address https://localhost:5678 --mesh abc
