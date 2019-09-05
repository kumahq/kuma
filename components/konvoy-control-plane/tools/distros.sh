#!/usr/bin/env bash

GOOS=( darwin linux )
GOARCH=( amd64 )

for os in "${GOOS[@]}"
do
    for arch in "${GOARCH[@]}"
    do
        make GOOS=$os GOARCH=$arch build/artifact-tarball
        curl -T build/artifacts-$os-$arch/kuma-$os-$arch.tar.gz -u $BINTRAY_USERNAME:$BINTRAY_API_KEY "https://api.bintray.com/content/kong/kuma/$os/$RELEASE_TAG-$arch/kuma-$RELEASE_TAG-$os-$arch.tar.gz?publish=1&override=1"
    done
done
