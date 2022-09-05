#!/bin/bash

OUTPUT_DIR=$1/bin
VERSION="13.0.0"
CLANG_FORMAT=${OUTPUT_DIR}/clang-format
# No arm64 linux so let's do a dummy script
if [ "${ARCH}" == "arm64" ] && [ "${OS}" == "linux" ]; then
  printf "#!/bin/bash\necho clang-format not suported on arm linux" > "${CLANG_FORMAT}"
  chmod u+x "${CLANG_FORMAT}"
  exit
fi
# There's no clang-format for arm64 mac so let's just use the amd64
if [ "${OS}" == "darwin" ]; then
  ARCH="amd64"
  OS="macosx"
fi

VERSION_NAME="clang-format-13_${OS}-${ARCH}"
if [ -e "${CLANG_FORMAT}" ] && [ "v$("${CLANG_FORMAT}" --version | cut -f3 -d ' ')" == "v${VERSION}" ]; then
  echo "$("${CLANG_FORMAT}" --version | head -1) is already installed at ${OUTPUT_DIR}"
  exit
fi
echo "Installing clang-format ${VERSION}..."
set -x
curl --location --fail -s -o "${CLANG_FORMAT}" https://github.com/muttleyxd/clang-tools-static-binaries/releases/download/master-208096c1/"${VERSION_NAME}"
chmod u+x "${CLANG_FORMAT}"
set +x
