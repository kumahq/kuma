#!/usr/bin/env bash

KUMA_SYSTEM=${KUMA_SYSTEM:-kuma-system}

kubectl delete --namespace ${KUMA_SYSTEM} secret/kuma-injector-tls-cert
kubectl delete --namespace ${KUMA_SYSTEM} configmap/kuma-injector-config
kubectl delete --namespace ${KUMA_SYSTEM} serviceaccount/kuma-injector
kubectl delete --namespace ${KUMA_SYSTEM} service/kuma-injector
kubectl delete --namespace ${KUMA_SYSTEM} deployment.apps/kuma-injector

kubectl delete mutatingwebhookconfiguration.admissionregistration.k8s.io/kuma-injector-webhook-configuration
kubectl delete clusterrole.rbac.authorization.k8s.io/kuma:injector
kubectl delete clusterrolebinding.rbac.authorization.k8s.io/kuma:injector
