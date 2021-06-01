#!/usr/bin/env bash

sudo /usr/sbin/sshd -D -e &
/usr/local/bin/docker-entrypoint.sh postgres
