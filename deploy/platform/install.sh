#!/usr/bin/env bash
# Bootstraps the cluster platform (Knative Serving+Eventing, Kourier, Strimzi
# Kafka, Knative Kafka Broker) on a fresh minikube cluster. Run from repo root.
# This is the manual M3 platform install, codified — later replaced by Terraform (M7).
set -euo pipefail

SERVING=knative-v1.22.1
KOURIER=knative-v1.22.1
EVENTING=knative-v1.22.2
KAFKA_BROKER=knative-v1.22.1

echo "==> Knative Serving"
kubectl apply -f https://github.com/knative/serving/releases/download/$SERVING/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/$SERVING/serving-core.yaml
kubectl -n knative-serving wait --for=condition=Available deploy --all --timeout=420s

echo "==> Kourier networking"
kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/download/$KOURIER/kourier.yaml
kubectl -n kourier-system wait --for=condition=Available deploy --all --timeout=300s
kubectl patch configmap/config-network -n knative-serving --type merge \
  -p '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
kubectl patch configmap/config-deployment -n knative-serving --type merge \
  -p '{"data":{"registries-skipping-tag-resolving":"kind.local,ko.local,dev.local"}}'

echo "==> Knative Eventing"
kubectl apply -f https://github.com/knative/eventing/releases/download/$EVENTING/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/$EVENTING/eventing-core.yaml
kubectl -n knative-eventing wait --for=condition=Available deploy --all --timeout=420s

echo "==> Strimzi (Kafka operator)"
kubectl create namespace kafka --dry-run=client -o yaml | kubectl apply -f -
kubectl apply --server-side -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
kubectl -n kafka wait --for=condition=Available deploy/strimzi-cluster-operator --timeout=300s

echo "==> Kafka cluster"
kubectl wait --for=condition=Established crd/kafkas.kafka.strimzi.io --timeout=120s
kubectl apply -f deploy/k8s/kafka/kafka-cluster.yaml
kubectl -n kafka wait kafka/alerter-kafka --for=condition=Ready --timeout=480s

echo "==> Knative Kafka Broker"
kubectl apply -f https://github.com/knative-extensions/eventing-kafka-broker/releases/download/$KAFKA_BROKER/eventing-kafka-controller.yaml
kubectl apply -f https://github.com/knative-extensions/eventing-kafka-broker/releases/download/$KAFKA_BROKER/eventing-kafka-broker.yaml
kubectl patch configmap/kafka-broker-config -n knative-eventing --type merge \
  -p '{"data":{"bootstrap.servers":"alerter-kafka-kafka-bootstrap.kafka:9092","default.topic.replication.factor":"1"}}'
kubectl -n knative-eventing wait --for=condition=Available deploy/kafka-broker-receiver deploy/kafka-controller --timeout=300s
kubectl -n knative-eventing rollout status statefulset/kafka-broker-dispatcher --timeout=300s

echo "==> app namespace"
kubectl create namespace alerter --dry-run=client -o yaml | kubectl apply -f -

echo "Platform ready."
