#!/bin/bash
# set -x

DOCKER_AUTH=$(jq -r '.auths | keys | .[]' $HOME/.docker/config.json | grep docker)
CRED=$(jq -r ".auths[\"$DOCKER_AUTH\"].auth //empty" $HOME/.docker/config.json)
if [[ "$CRED" == "" ]]; then
    >&2 echo "No existing docker auth information found on the host"
else
    DOCKER_USER=$(echo $CRED | base64 --decode | cut -d ':' -f 1)
    DOCKER_PWD=$(echo $CRED | base64 --decode | cut -d ':' -f 2)
    cat <<EOF > $HOME/.docker/k3d-registry.yaml
configs:
  registry-1.docker.io:
    auth:
      username: $DOCKER_USER
      password: $DOCKER_PWD
EOF
    echo $HOME/.docker/k3d-registry.yaml

    echo "{\"auths\":{\"https://index.docker.io/v1/\":{\"auth\":\"$CRED\"}}}" > $HOME/.docker/kind-config.json
    EXTRA_MOUNTS="[ {\"containerPath\": \"/var/lib/kubelet/config.json\", \"hostPath\": \"$HOME/.docker/kind-config.json\"}]"
    for FILE in $(ls test/kind/*.yaml); do
        if [[ ! -z "$(yq '.nodes[] | select(.role == "control-plane") | length' $FILE)" ]]; then
            yq -i ".nodes[] | select(.role == \"control-plane\") *= {\"extraMounts\":$EXTRA_MOUNTS} | parent | parent" $FILE
        else
            yq -i ".nodes = [ { \"role\": \"control-plane\", \"extraMounts\": $EXTRA_MOUNTS }]" $FILE
        fi
    done
fi
