#!/usr/bin/env bash

grep -v "*" "$1" | cut -d '!' -f 2 | xargs | sha256sum | cut -d ' ' -f 1
