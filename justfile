set shell := ["bash", "-c"]

# images
SCYLLA_IMAGE := "scylladb/scylla"
MINIO_IMAGE  := "quay.io/minio/minio"

# configurations
SCYLLA_CONFIG := invocation_directory() / "test-resources/scylla-config"
MANIFESTS := invocation_directory() / "test-resources/manifests"

# helm for Nexus
NEXUS_CHART_NAME := "nexus"
NEXUS_CHART_PATH := "./.helm"
NEXUS_CHART_IMAGE_NAME := "nexus-dev"
NEXUS_CHART_IMAGE_TAG  := "latest"

# cluster
NEXUS_CLUSTER_NAME := "nexus-controller-0"

# Default recipe
fresh: stop up

# Start CI environment
up: start-kind-cluster install-ingress-controller create-namespace create-ingress scylla-kind minio-kind crd build-image load-image deploy-chart

start-kind-cluster:
    kind create cluster --config=test-resources/kind.yaml --name {{NEXUS_CLUSTER_NAME}}

# Run all tests
test:
    go test -v ./...

# Cleanup CI environment
stop:
    @echo "🧹 Cleaning up..."
    docker rm -f scylla minio 2>/dev/null || true
    kind delete cluster --name {{NEXUS_CLUSTER_NAME}}

# View logs
logs name="":
    docker logs -f {{if name == "" { "scylla" } else { name }}}

# build the local Docker image
build-image:
    docker build -t {{NEXUS_CHART_IMAGE_NAME}}:{{NEXUS_CHART_IMAGE_TAG}} -f .container/Dockerfile .

# load image into the cluster
load-image:
    kind load docker-image {{NEXUS_CHART_IMAGE_NAME}}:{{NEXUS_CHART_IMAGE_TAG}} --name  {{NEXUS_CLUSTER_NAME}}

create-namespace:
    kubectl create namespace nexus --dry-run=client -o yaml | kubectl apply -f -

# install chart
deploy-chart:
    kubectl create secret generic cassandra-credentials \
        --namespace nexus \
        --from-literal=NEXUS__SCYLLA_CQL_STORE__HOSTS="scylla.nexus.svc.cluster.local" \
        --from-literal=NEXUS__SCYLLA_CQL_STORE__INDEXES_SUPPORTED="true" \
        --from-literal=NEXUS__SCYLLA_CQL_STORE__USER="cassandra" \
        --from-literal=NEXUS__SCYLLA_CQL_STORE__PASSWORD="cassandra" \
        --from-literal=NEXUS__SCYLLA_CQL_STORE__KEYSPACE="nexus" --dry-run=client -o yaml | kubectl apply -f -

    kubectl create secret generic nexus-s3 \
        --namespace nexus \
        --from-literal=NEXUS__S3_BUFFER__REGION="us-east-1" \
        --from-literal=NEXUS__S3_BUFFER__ACCESS_KEY_ID="minioadmin" \
        --from-literal=NEXUS__S3_BUFFER__SECRET_ACCESS_KEY="minioadmin" \
        --from-literal=NEXUS__S3_BUFFER__ENDPOINT="http://minio.nexus.svc.cluster.local:9000" --dry-run=client -o yaml | kubectl apply -f -

    kubectl create secret generic nexus-sign-key \
        --namespace nexus \
        --from-literal=NEXUS__S3_BUFFER__REQUEST_PAYLOAD_PROXY_CONFIGURATION__SIGN_SECRET="test" --dry-run=client -o yaml | kubectl apply -f -

    helm upgrade --install --create-namespace --namespace nexus {{NEXUS_CHART_NAME}} {{NEXUS_CHART_PATH}} \
        --set image.repository={{NEXUS_CHART_IMAGE_NAME}} \
        --set image.tag={{NEXUS_CHART_IMAGE_TAG}} \
        --set image.pullPolicy=Never \
        --set scheduler.config.checkpointStore.type=cassandra-scylla \
        --set scheduler.config.checkpointStore.secretName="cassandra-credentials" \
        --set scheduler.config.s3Buffer.s3Credentials.secretName="nexus-s3"

# cleanup
remove-chart:
    helm uninstall -n nexus {{NEXUS_CHART_NAME}}

install-ingress-controller:
    kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml
    kubectl rollout status deployment/ingress-nginx-controller -n ingress-nginx --timeout=180s

create-ingress:
    # Create ingress rules for services
    for i in $(seq 1 30); do \
      kubectl apply -f {{MANIFESTS}}/ingress.yaml && break || \
      (echo "Retry $i/30: failed to apply ingress, retrying in 1s..." && sleep 1); \
    done; \
    if [ $i -eq 30 ]; then \
      echo "Failed to apply ingress after 30 attempts."; \
      exit 1; \
    fi

scylla-kind:
    kubectl apply -f {{MANIFESTS}}/scylladb.yaml
    kubectl -n nexus rollout status deployment/scylla --timeout=180s

minio-kind:
    kubectl apply -f {{MANIFESTS}}/minio.yaml
    kubectl -n nexus rollout status deployment/minio --timeout=180s

crd:
    helm upgrade --install --namespace nexus nexus-crd  oci://ghcr.io/sneaksanddata/helm/nexus-crd --version v1.0.0-4-gefa0d24
