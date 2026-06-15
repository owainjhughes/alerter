#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing Skaffold"
curl -fsSL -o /tmp/skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install /tmp/skaffold /usr/local/bin/skaffold

echo "==> Installing kubeseal"
KUBESEAL_VERSION="0.27.1"
curl -fsSL -o /tmp/kubeseal.tar.gz \
  "https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-linux-amd64.tar.gz"
tar -xzf /tmp/kubeseal.tar.gz -C /tmp kubeseal
sudo install /tmp/kubeseal /usr/local/bin/kubeseal

echo "==> Installing ArgoCD CLI"
curl -fsSL -o /tmp/argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install /tmp/argocd /usr/local/bin/argocd

echo "==> Installing golangci-lint"
curl -fsSL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
  sh -s -- -b "$(go env GOPATH)/bin"

skaffold version || true
kubeseal --version || true
argocd version --client || true
