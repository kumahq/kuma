#!/bin/bash

OUTPUT_DIR=$1/bin
VERSION="1.23.5"
KUBECTL=${OUTPUT_DIR}/kubectl
if [ -e "${KUBECTL}" ] && [ "$(${KUBECTL} version -o yaml --client=true | grep gitVersion | cut -f4 -d ' ')" == "v${VERSION}" ]; then
  echo "kubectl version ${VERSION} is already installed at ${OUTPUT_DIR}"
  exit
fi
echo "Installing Kubernetes ${CI_KUBEBUILDER_VERSION} ..."
set -x
for component in kube-apiserver kubectl; do
  rm -f "${OUTPUT_DIR}"/${component}
  if [ "${OS}" == "darwin" ] && [ ${component} == "kube-apiserver" ]; then
    # There's no official build of kube-apiserver on darwin so we'll just get the one from kubebuilder
    KUBEBUILDER_VERSION=2.3.2
    VERSION_NAME=kubebuilder_${KUBEBUILDER_VERSION}_${OS}_amd64
    curl --location --fail -s https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/"${VERSION_NAME}".tar.gz | tar --strip-components=2 -xz -C "${OUTPUT_DIR}" "${VERSION_NAME}"/bin/kube-apiserver
  else
    curl --location -o "${OUTPUT_DIR}"/${component} --fail -s  https://dl.k8s.io/v${VERSION}/bin/"${OS}"/"${ARCH}"/${component}
  fi
  chmod +x "${OUTPUT_DIR}"/${component}
done
set +x
echo "kubebuilder ${CI_KUBEBUILDER_VERSION} and dependencies has been installed at ${KUBEBUILDER}"
