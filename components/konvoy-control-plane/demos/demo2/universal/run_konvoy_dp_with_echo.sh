#!/usr/bin/env sh

wget -O /tmp/http-echo.zip  https://github.com/hashicorp/http-echo/releases/download/v0.2.3/http-echo_0.2.3_linux_amd64.zip
unzip -o /tmp/http-echo.zip -d /tmp

/tmp/http-echo -text "$1" -listen="$2" &
/konvoy-dataplane/konvoy-dataplane run
