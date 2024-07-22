#!/bin/bash
set -e

if [[ "$DOCKERHUB_PULL_CREDENTIAL" == "" ]]; then
    >&2 echo "No docker pull credential information specified"
else
    DOCKER_USER=$(echo "$DOCKERHUB_PULL_CREDENTIAL" | base64 --decode | cut -d ':' -f 1)
    DOCKER_PWD=$(echo "$DOCKERHUB_PULL_CREDENTIAL" | base64 --decode | cut -d ':' -f 2)
    echo -n "$DOCKER_PWD" | docker login -u "$DOCKER_USER" --password-stdin > /dev/null
    cat <<EOF > "$HOME/.docker/k3d-registry.yaml"
configs:
  registry-1.docker.io:
    auth:
      username: $DOCKER_USER
      password: $DOCKER_PWD
EOF
    echo "$HOME/.docker/k3d-registry.yaml"

    echo "{\"auths\":{\"https://index.docker.io/v1/\":{\"auth\":\"$CRED\"}}}" > "$HOME/.docker/kind-config.json"
    EXTRA_MOUNTS="[ {\"containerPath\": \"/var/lib/kubelet/config.json\", \"hostPath\": \"$HOME/.docker/kind-config.json\"}]"
    for FILE in test/kind/*.yaml; do
        if [[ "$("$CI_TOOLS_DIR"/bin/yq '.nodes[] | select(.role == "control-plane") | length' "$FILE")" != "" ]]; then
            "$CI_TOOLS_DIR"/bin/yq -i ".nodes[] | select(.role == \"control-plane\") *= {\"extraMounts\":$EXTRA_MOUNTS} | parent | parent" "$FILE"
        else
            "$CI_TOOLS_DIR"/bin/yq -i ".nodes = [ { \"role\": \"control-plane\", \"extraMounts\": $EXTRA_MOUNTS }]" "$FILE"
        fi
    done
fi
