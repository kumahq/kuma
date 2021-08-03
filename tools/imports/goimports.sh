#!/bin/bash

# goimports preserves the blank lines in the 'import' section
# and because of that doesn't sort properly. This script deletes all
# the blank lines in the 'import' section and calls goimports afterwards

if [ $(uname -s) == "Darwin" ]; then
  sed -i '' '/^import/,/)/ {
      /^$/ d
    }
  ' $@
else
    sed -i '/^import/,/)/ {
      /^$/ d
    }
  ' $@
fi

goimports -w -local github.com/kumahq/kuma $@