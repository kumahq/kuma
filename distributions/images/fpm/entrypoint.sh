#!/bin/sh

set -e

[ -z "${SOURCE_DIR}" ] && echo "SOURCE_DIR must be set" && exit 1
[ -z "${OUTPUT_DIR}" ] && echo "OUTPUT_DIR must be set" && exit 1
[ -z "${PACKAGE_NAME}" ] && echo "PACKAGE_NAME must be set" && exit 1
[ -z "${PACKAGE_VERSION}" ] && echo "PACKAGE_VERSION must be set" && exit 1
[ -z "${PACKAGE_ARCH}" ] && echo "PACKAGE_ARCH must be set" && exit 1
[ -z "${PACKAGE_DESCRIPTION}" ] && echo "PACKAGE_DESCRIPTION must be set" && exit 1
[ -z "${PACKAGE_URL}" ] && echo "PACKAGE_URL must be set" && exit 1
[ -z "${PACKAGE_VENDOR}" ] && echo "PACKAGE_VENDOR must be set" && exit 1
[ -z "${PACKAGE_MAINTAINER}" ] && echo "PACKAGE_MAINTAINER must be set" && exit 1
[ -z "${PACKAGE_LICENSE}" ] && echo "PACKAGE_LICENSE must be set" && exit 1
[ -z "${BASE_IMAGE_NAME}" ] && echo "BASE_IMAGE_NAME must be set" && exit 1
[ -z "${BASE_IMAGE_TAG}" ] && echo "BASE_IMAGE_TAG must be set" && exit 1

[ ! -d "${OUTPUT_DIR}" ] && mkdir -p "${OUTPUT_DIR}"

DEPS="-d ca-certificates"
if [ "${BASE_IMAGE_NAME}" = "ubuntu" ] || [ "${BASE_IMAGE_NAME}" = "debian" ]; then
  PACKAGE_TYPE="deb"
  OUTPUT_FILE_SUFFIX="${BASE_IMAGE_NAME}${BASE_IMAGE_TAG}"
elif [ "${BASE_IMAGE_NAME}" = "centos" ]; then
  PACKAGE_TYPE="rpm"
  OUTPUT_FILE_SUFFIX="el${BASE_IMAGE_TAG}"
elif [ "${BASE_IMAGE_NAME}" = "rhel" ]; then
  PACKAGE_TYPE="rpm"
  OUTPUT_FILE_SUFFIX="rhel${BASE_IMAGE_TAG}"
elif [ "${BASE_IMAGE_NAME}" = "amazonlinux" ]; then
  PACKAGE_TYPE="rpm"
  OUTPUT_FILE_SUFFIX="aws"
elif [ "${BASE_IMAGE_NAME}" = "alpine" ]; then
  PACKAGE_TYPE="tar"
  OUTPUT_FILE_SUFFIX="apk"
else
  PACKAGE_TYPE="tar"
  OUTPUT_FILE_SUFFIX="generic"
fi

OUTPUT_FILE="${OUTPUT_DIR}/${PACKAGE_NAME}-${PACKAGE_VERSION}-linux.${OUTPUT_FILE_SUFFIX}.${PACKAGE_ARCH}.${PACKAGE_TYPE}"

fpm --force \
  --input-type dir \
  --output-type ${PACKAGE_TYPE} \
  --name ${PACKAGE_NAME} \
  --version ${PACKAGE_VERSION} \
  --architecture ${PACKAGE_ARCH} \
  --description "${PACKAGE_DESCRIPTION}" \
  --provides ${PACKAGE_NAME} \
  ${DEPS} \
  --url "${PACKAGE_URL}" \
  --vendor "${PACKAGE_VENDOR}" \
  --maintainer "${PACKAGE_MAINTAINER}" \
  --license "${PACKAGE_LICENSE}" \
  --package "${OUTPUT_FILE}" \
  --chdir ${SOURCE_DIR} \
  usr/ etc/

if [ "${PACKAGE_TYPE}" = "tar" ]; then
  gzip --force "${OUTPUT_FILE}"
fi
