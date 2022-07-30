#!/bin/bash

OLD_SHELL="$1"
TARGET="$2"
KUMA_DIR="$3"
shift
shift
shift

# time in ms
START=$(gdate +%s%3N)
"$OLD_SHELL" "$@"
STOP=$(gdate +%s%3N)
TOTAL=$((STOP-START))

if (( TOTAL > 0 )); then
    echo "{\"$TARGET\": $TOTAL}" >> "$KUMA_DIR"/build/build-times
fi
