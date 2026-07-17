set shell := ["bash", "-c"]

# images
SCYLLA_IMAGE := "scylladb/scylla"
MINIO_IMAGE  := "quay.io/minio/minio"

# configurations
SCYLLA_CONFIG := invocation_directory() / "test-resources/scylla-config"

# Default recipe
fresh: stop up

# Start CI environment
up: scylla minio prepare-buckets prepare-scylla start-kind-cluster shards-kubeconfig

# Start ScyllaDB with health checks
scylla:
    @echo "🚀 Starting Scylla..."
    docker run -d \
      --name scylla \
      -p 9042:9042 \
      -p 10000:10000 \
      --health-cmd "nodetool statusgossip | grep -q 'running'" \
      --health-interval 5s \
      --health-retries 10 \
      {{SCYLLA_IMAGE}} \
      --listen-address 127.0.0.1 \
      --rpc-address 0.0.0.0 \
      --broadcast-rpc-address 0.0.0.0 \
      --smp 1 \
      --developer-mode 1

# Start Minio with health checks
minio:
    @echo "📦 Starting Minio..."
    docker run -d \
      --name minio \
      -p 9000:9000 \
      -p 9123:9123 \
      --restart always \
      --health-cmd "curl -f http://localhost:9000/minio/health/live" \
      --health-interval 5s \
      {{MINIO_IMAGE}} \
      server /data --console-address :9123

# Create Minio buckets using built-in health waiting
prepare-buckets:
    @echo "⏳ Waiting for Minio..."
    until [ "$(docker inspect -f '{{ "{{" }}.State.Health.Status{{ "}}" }}' minio)" == "healthy" ]; do sleep 2; done
    docker run --rm \
      --network host \
      --entrypoint "/bin/sh" \
      {{MINIO_IMAGE}} \
      -c "mc alias set e2e http://localhost:9000 minioadmin minioadmin && \
          mc mb --ignore-existing e2e/tmp e2e/nexus"

# Run Scylla initialization
prepare-scylla:
    @echo "⏳ Waiting for Scylla..."
    until [ "$(docker inspect -f '{{ "{{" }}.State.Health.Status{{ "}}" }}' scylla)" == "healthy" ]; do sleep 2; done
    docker run --rm \
      --network host \
      -v {{SCYLLA_CONFIG}}:/opt/storage \
      --entrypoint /opt/storage/prepare-scylla.sh \
      {{SCYLLA_IMAGE}}

start-kind-cluster:
    kind create cluster --config=test-resources/kind.yaml --name nexus-shard-0

shards-kubeconfig:
    mkdir -p ./test-resources/kind && \
    kind export kubeconfig --name nexus-shard-0 --kubeconfig ./test-resources/kind/kind-nexus-shard-0.kubeconfig

# Run all tests
test:
    go test -v ./...

# Cleanup CI environment
stop:
    @echo "🧹 Cleaning up..."
    docker rm -f scylla minio 2>/dev/null || true
    kind delete cluster --name nexus-shard-0

# View logs
logs name="":
    docker logs -f {{if name == "" { "scylla" } else { name }}}

