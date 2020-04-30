#!/usr/bin/env bash

kubectl -n monitoring create configmap kubeconfigs \
--from-file=config-177=config-177 \
--from-file=config-175=config-175 \
--from-file=config-110=config-110 \
--from-file=config-168=config-168 --dry-run -o yaml | kubectl apply -f -


kubectl -n monitoring create configmap storesd-config \
--from-file=config.yaml=config.yaml --dry-run -o yaml | kubectl apply -f -